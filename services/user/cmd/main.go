package main

import (
	"context"
	"os/signal"
	"restaurant/pkg/config"
	"restaurant/pkg/logger"
	"restaurant/pkg/tls"
	"restaurant/services/user/internal/client"
	"restaurant/services/user/internal/delivery/grpc"
	"restaurant/services/user/internal/service"
	"restaurant/services/user/internal/storage/sqlite3"
	"restaurant/services/user/pkg/hasher"
	"syscall"
	"time"
)

func main() {
	/* Конфиг */
	config.Load("user")

	/* Логгер */
	logger.Init("User")

	/* Хранилище юзеров */
	storage, err := sqlite3.New("data/users.db")
	if err != nil {
		logger.Error.Printf("ошибка создания хранилища: %v", err)
	}
	defer func() {
		if err := storage.Close(); err != nil {
			logger.Warn.Printf("ошибка закрытия хранилища: %v", err)
		}
	}()

	/* TLS Client */
	clientCreds, err := tls.ClientCreds(
		config.Get[string]("CA_CERT", "certs/ca/ca-cert.pem"),
		config.Get[string]("USER_CERT", "certs/user/server-cert.pem"),
		config.Get[string]("USER_KEY", "certs/user/server-key.pem"),
		"auth")
	if err != nil {
		logger.Error.Printf("ошибка создания mTLS: %v", err)
	}

	/* TLS Server */
	serverCreds, err := tls.ServerCreds(
		config.Get[string]("CA_CERT", "certs/ca/ca-cert.pem"),
		config.Get[string]("USER_CERT", "certs/user/server-cert.pem"),
		config.Get[string]("USER_KEY", "certs/user/server-key.pem"))
	if err != nil {
		logger.Error.Printf("ошибка создания mTLS: %v", err)
	}

	/* gRPC-auth client */
	authClient, err := client.NewAuthClient(clientCreds, config.Get[string]("AUTH_GRPC_LISTENER", "localhost:50051"))
	if err != nil {
		logger.Error.Printf("ошибка gRPC клиента: %v", err)
	}
	defer authClient.Close()

	/* Hasher */
	hasher := hasher.New()

	/* User service */
	userService := service.NewUserService(storage, hasher)

	/* gRPC Server */
	srv := delivery.NewGRPCServer(serverCreds, userService, authClient,
		config.Get[string]("USER_GRPC_LISTENER", "localhost:50052"),
		time.Duration(config.Get[int]("USER_SHUTDOWN", 30))*time.Second)

	go func() {
		if err = srv.Run(); err != nil {
			logger.Error.Printf("ошибка запуска gRPC сервера: %v", err)
		}
	}()

	/* Graceful shutdown */
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	<-ctx.Done()

	logger.Info.Println("получен сигнал завершение сервера, останавливаем..")

	if err = srv.Stop(); err != nil {
		logger.Warn.Printf("программа завершилась не gracefully: %v", err)
		return
	}
	logger.Info.Println("программа завершилась gracefully")
}
