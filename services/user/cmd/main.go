package main

import (
	"context"
	"log"
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
)

func main() {
	/* Конфиг */
	cfg := &config.UserConfig{}
	if err := config.Load("./configs/config.yaml", "ENV", cfg); err != nil {
		log.Fatalf("config load: %v", err)
	}

	/* Логгер */
	logger.SetupLogger(cfg.ENV, "user")

	slog.Info("Server data:",
		slog.String("ENV", cfg.ENV))

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
	authCreds, err := tls.ClientCreds(
		cfg.CACert,
		cfg.User.CertsClient.Cert, cfg.User.CertsClient.CertKey,
		"auth")
	if err != nil {
		slog.Error("failed to create mTLS",
			slog.String("error", err.Error()),
			slog.String("type", "auth_client"))
		os.Exit(1)
	}

	/* TLS Server */
	serverCreds, err := tls.ServerCreds(
		cfg.CACert,
		cfg.User.CertsServer.Cert, cfg.User.CertsServer.CertKey)
	if err != nil {
		slog.Error("failed to create mTLS",
			slog.String("error", err.Error()),
			slog.String("type", "user_server"))
		os.Exit(1)
	}

	/* gRPC-auth client */
	authClient, err := client.NewAuthClient(authCreds, cfg.AuthAddr,
		client.AuthConfig{
			RetryMaxAttempts:       cfg.User.GRPCAuthClient.RetryMaxAttempts,
			RetryInitialBackoff:    cfg.User.GRPCAuthClient.RetryInitialBackoff,
			RetryMaxBackoff:        cfg.User.GRPCAuthClient.RetryMaxBackoff,
			RetryBackoffMultiplier: cfg.User.GRPCAuthClient.RetryBackoffMultiplier,
			KeepaliveTime:          cfg.User.GRPCAuthClient.KeepaliveTime,
			KeepaliveTimeout:       cfg.User.GRPCAuthClient.KeepaliveTimeout,
			KeepalivePermitWithout: cfg.User.GRPCAuthClient.KeepalivePermitWithout,
		})
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
		cfg.UserAddr, cfg.User.ShutdownTimeout,
		delivery.OptionConfig{
			MaxReceivedSize:   cfg.User.GRPCMaxRecvMsgSize,
			MaxSendSize:       cfg.User.GRPCMaxSendMsgSize,
			ConnectionTimeout: cfg.User.GRPCConnTimeout,
			MaxConnectionIdle: cfg.User.GRPCMaxConnIdle,
			KeepAliveTime:     cfg.User.GRPCKeepaliveTime,
			KeepAliveTimeout:  cfg.User.GRPCKeepaliveTimeout,
		})

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
