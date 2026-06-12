package main

import (
	"context"
	"os/signal"
	"restaurant/pkg/config"
	"restaurant/pkg/logger"
	"restaurant/pkg/tls"
	"restaurant/services/auth/internal/delivery/grpc"
	"restaurant/services/auth/internal/service"
	"restaurant/services/auth/internal/storage/sqlite3"
	"syscall"
	"time"
)

func main() {
	/* Конфиг */
	config.Load("auth")

	/* Логгер */
	logger.Init("Auth")

	/* Хранилище refresh-token */
	storage, err := sqlite3.New("tokens.db")
	if err != nil {
		logger.Error.Printf("ошибка создания хранилища: %v", err)
	}
	defer func() {
		if err := storage.Close(); err != nil {
			logger.Warn.Printf("ошибка закрытия хранилища: %v", err)
		}
	}()

	/* JWT-Service */
	jwtService := service.NewJWT(
		config.Get[string]("AUTH_SECRET_KEY", "default-secret-key-min-32-chars"),
		time.Duration(config.Get[int]("AUTH_ACCESS_TTL", 15))*time.Minute,
		time.Duration(config.Get[int]("AUTH_REFRESH_TTL", 7))*time.Hour*24,
		storage)

	/* TLS Server */
	creds, err := tls.ServerCreds(
		config.Get[string]("CA_CERT", "certs/ca/ca-cert.pem"),
		config.Get[string]("AUTH_CERT", "certs/auth/server-cert.pem"),
		config.Get[string]("AUTH_KEY", "certs/auth/server-key.pem"))
	if err != nil {
		logger.Error.Printf("ошибка создания mTLS: %v", err)
	}

	/* gRPC Server */
	srv := delivery.NewGRPCServer(creds, jwtService,
		config.Get[string]("AUTH_GRPC_LISTENER", ":50051"),
		time.Duration(config.Get[int]("AUTH_SHUTDOWN", 30))*time.Second)

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
