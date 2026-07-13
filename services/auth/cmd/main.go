package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"restaurant/pkg/clients"
	"restaurant/pkg/config"
	"restaurant/pkg/logger"
	"restaurant/pkg/tls"
	delivery "restaurant/services/auth/internal/delivery/grpc"
	"restaurant/services/auth/internal/service"
	"restaurant/services/auth/internal/storage/redis"
	"syscall"
)

func main() {
	/* Конфиг */
	cfg := &config.AuthConfig{}
	if err := config.Load("./configs/config.yaml", "ENV", cfg); err != nil {
		log.Fatalf("config load: %v", err)
	}

	/* Логгер */
	logger.SetupLogger(cfg.ENV, "auth")

	slog.Info("Server data:",
		slog.String("ENV", cfg.ENV))

	/* Хранилище refresh-token */
	redisClient, err := clients.NewRedis(&clients.RedisConfig{
		Addr:     cfg.Auth.RedisClient.Addr,
		Password: cfg.Auth.RedisClient.Password,
		DB:       cfg.Auth.RedisClient.DB,
		PoolSize: cfg.Auth.RedisClient.PoolSize,
	})
	if err != nil {
		slog.Error("failed to create storage",
			slog.String("error", err.Error()))
		os.Exit(1)
	}
	storage := redis.New(redisClient.Client, "auth")
	defer func() {
		if err := storage.Close(); err != nil {
			slog.Warn("failed to close storage",
				slog.String("error", err.Error()),
				slog.String("storage_type", "rediss"))
		}
	}()

	/* JWT-Service */
	jwtService := service.NewJWT(
		cfg.Auth.SecretKey,
		cfg.Auth.AccessTTL, cfg.Auth.RefreshTTL,
		storage)

	/* TLS Server */
	creds, err := tls.ServerCreds(
		cfg.CACert,
		cfg.Auth.CertsServer.Cert, cfg.Auth.CertsServer.CertKey)
	if err != nil {
		slog.Error("failed to create mTLS",
			slog.String("error", err.Error()),
			slog.String("type", "auth_server"))
		os.Exit(1)
	}

	/* gRPC Server */
	srv := delivery.NewGRPCServer(creds, jwtService,
		cfg.AuthAddr, cfg.Auth.ShutdownTimeout,
		delivery.OptionConfig{
			MaxReceivedSize:   cfg.Auth.GRPCMaxRecvMsgSize,
			MaxSendSize:       cfg.Auth.GRPCMaxSendMsgSize,
			ConnectionTimeout: cfg.Auth.GRPCConnTimeout,
			MaxConnectionIdle: cfg.Auth.GRPCMaxConnIdle,
			KeepAliveTime:     cfg.Auth.GRPCKeepaliveTime,
			KeepAliveTimeout:  cfg.Auth.GRPCKeepaliveTimeout,
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
