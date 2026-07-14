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
	delivery "restaurant/services/user/internal/delivery/grpc"
	"restaurant/services/user/internal/service"
	"restaurant/services/user/internal/storage/postgres"
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
	storage, err := postgres.New(postgres.Config{
		Addr:     cfg.User.PostgreSQL.Addr,
		User:     cfg.User.PostgreSQL.User,
		Password: cfg.User.PostgreSQL.Password,
		Name:     cfg.User.PostgreSQL.Name,
		SSLMode:  cfg.User.PostgreSQL.SSLMode,
	})
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

	/* Hasher */
	hasher := hasher.New()

	/* User service */
	userService := service.NewUserService(storage, hasher)

	/* gRPC Server */
	srv := delivery.NewGRPCServer(serverCreds, userService,
		cfg.UserAddr, cfg.User.ShutdownTimeout,
		delivery.OptionConfig{
			MaxReceivedSize:   cfg.User.GRPCServer.MaxRecvMsgSize,
			MaxSendSize:       cfg.User.GRPCServer.MaxSendMsgSize,
			ConnectionTimeout: cfg.User.GRPCServer.ConnTimeout,
			MaxConnectionIdle: cfg.User.GRPCServer.MaxConnIdle,
			KeepAliveTime:     cfg.User.GRPCServer.KeepaliveTime,
			KeepAliveTimeout:  cfg.User.GRPCServer.KeepaliveTimeout,
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
