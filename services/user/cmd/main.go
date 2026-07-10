package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"restaurant/pkg/config"
	"restaurant/pkg/logger"
	"restaurant/pkg/tls"
	"restaurant/services/user/internal/client"
	delivery "restaurant/services/user/internal/delivery/grpc"
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
	logger.SetupLogger("dev", "User")

	/* Хранилище юзеров */
	storage, err := sqlite3.New("data/users.db")
	if err != nil {
		slog.Error("failed to create storage",
			slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer func() {
		if err := storage.Close(); err != nil {
			slog.Warn("failed to close storage",
				slog.String("error", err.Error()),
				slog.String("storage_type", "sqlite3"))
		}
	}()

	/* TLS Client */
	clientCreds, err := tls.ClientCreds(
		config.Get[string]("CA_CERT", "certs/ca/ca-cert.pem"),
		config.Get[string]("USER_CERT", "certs/user/server-cert.pem"),
		config.Get[string]("USER_KEY", "certs/user/server-key.pem"),
		"auth")
	if err != nil {
		slog.Error("failed to create mTLS",
			slog.String("error", err.Error()),
			slog.String("type", "auth_client"))
		os.Exit(1)
	}

	/* TLS Server */
	serverCreds, err := tls.ServerCreds(
		config.Get[string]("CA_CERT", "certs/ca/ca-cert.pem"),
		config.Get[string]("USER_CERT", "certs/user/server-cert.pem"),
		config.Get[string]("USER_KEY", "certs/user/server-key.pem"))
	if err != nil {
		slog.Error("failed to create mTLS",
			slog.String("error", err.Error()),
			slog.String("type", "user_server"))
		os.Exit(1)
	}

	/* gRPC-auth client */
	authClient, err := client.NewAuthClient(clientCreds, config.Get[string]("AUTH_GRPC_LISTENER", "localhost:50051"))
	if err != nil {
		slog.Error("failed to run gRPC client",
			slog.String("error", err.Error()))
		os.Exit(1)
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
			slog.Error("failed to run gRPC server",
				slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	/* Graceful shutdown */
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	<-ctx.Done()

	slog.Info("got signal-notify")

	if err = srv.Stop(); err != nil {
		slog.Warn("failed to gracefully shutdown",
			slog.String("error", err.Error()))
		return
	}
	slog.Info("gracefully shutdown")
}
