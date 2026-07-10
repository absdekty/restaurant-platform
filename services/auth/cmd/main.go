package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"restaurant/pkg/config"
	"restaurant/pkg/logger"
	"restaurant/pkg/tls"
	delivery "restaurant/services/auth/internal/delivery/grpc"
	"restaurant/services/auth/internal/service"
	"restaurant/services/auth/internal/storage/sqlite3"
	"syscall"
	"time"
)

func main() {
	/* Конфиг */
	config.Load("auth")

	/* Логгер */
	logger.SetupLogger("dev", "auth")

	/* Хранилище refresh-token */
	storage, err := sqlite3.New("data/tokens.db")
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
		slog.Error("failed to create mTLS",
			slog.String("error", err.Error()),
			slog.String("type", "auth_server"))
		os.Exit(1)
	}

	/* gRPC Server */
	srv := delivery.NewGRPCServer(creds, jwtService,
		config.Get[string]("AUTH_GRPC_LISTENER", "localhost:50051"),
		time.Duration(config.Get[int]("AUTH_SHUTDOWN", 30))*time.Second)

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
