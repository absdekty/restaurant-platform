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
	"restaurant/services/gateway/internal/client"
	delivery "restaurant/services/gateway/internal/delivery/rest"
	"restaurant/services/gateway/internal/delivery/rest/middleware"
	"syscall"
)

func main() {
	/* Конфиг */
	cfg := &config.GatewayConfig{}
	if err := config.Load("./configs/config.yaml", "ENV", cfg); err != nil {
		log.Fatalf("config load: %v", err)
	}

	/* Логгер */
	logger.SetupLogger(cfg.ENV, "gateway")

	slog.Info("Server data:",
		slog.String("ENV", cfg.ENV))

	/* Redis client (Rate limiter) */
	rlRedisClient, err := clients.NewRedis(&clients.RedisConfig{
		Addr:     cfg.Gateway.RateLimiter.RedisClient.Addr,
		Password: cfg.Gateway.RateLimiter.RedisClient.Password,
		DB:       cfg.Gateway.RateLimiter.RedisClient.DB,
		PoolSize: cfg.Gateway.RateLimiter.RedisClient.PoolSize,
	})
	if err != nil {
		slog.Error("failed to create redis client",
			slog.String("error", err.Error()),
			slog.String("type", "rate limiter"))
		os.Exit(1)
	}
	defer func() {
		if err := rlRedisClient.Close(); err != nil {
			slog.Warn("failed to close redis client",
				slog.String("error", err.Error()))
		}
	}()

	/* TLS Clients */
	clientAuthCreds, err := tls.ClientCreds(
		cfg.CACert,
		cfg.Gateway.Cert, cfg.Gateway.CertKey,
		"auth")
	if err != nil {
		slog.Error("failed to create mTLS",
			slog.String("error", err.Error()),
			slog.String("type", "auth_client"))
		os.Exit(1)

	}
	clientUserCreds, err := tls.ClientCreds(
		cfg.CACert,
		cfg.Gateway.Cert, cfg.Gateway.CertKey,
		"user")
	if err != nil {
		slog.Error("failed to create mTLS",
			slog.String("error", err.Error()),
			slog.String("type", "user_client"))
		os.Exit(1)
	}

	/* gRPC auth-client */
	authClient, err := client.NewAuthClient(clientAuthCreds, cfg.AuthAddr,
		client.AuthConfig{
			RetryMaxAttempts:       cfg.Gateway.GRPCAuthClient.RetryMaxAttempts,
			RetryInitialBackoff:    cfg.Gateway.GRPCAuthClient.RetryInitialBackoff,
			RetryMaxBackoff:        cfg.Gateway.GRPCAuthClient.RetryMaxBackoff,
			RetryBackoffMultiplier: cfg.Gateway.GRPCAuthClient.RetryBackoffMultiplier,
			KeepaliveTime:          cfg.Gateway.GRPCAuthClient.KeepaliveTime,
			KeepaliveTimeout:       cfg.Gateway.GRPCAuthClient.KeepaliveTimeout,
			KeepalivePermitWithout: cfg.Gateway.GRPCAuthClient.KeepalivePermitWithout,
		})
	if err != nil {
		slog.Error("failed to create gRPC client",
			slog.String("error", err.Error()),
			slog.String("type", "auth"))
		os.Exit(1)
	}
	defer authClient.Close()

	/* gRPC user-client */
	userClient, err := client.NewUserClient(clientUserCreds, cfg.UserAddr,
		client.UserConfig{
			RetryMaxAttempts:       cfg.Gateway.GRPCUserClient.RetryMaxAttempts,
			RetryInitialBackoff:    cfg.Gateway.GRPCUserClient.RetryInitialBackoff,
			RetryMaxBackoff:        cfg.Gateway.GRPCUserClient.RetryMaxBackoff,
			RetryBackoffMultiplier: cfg.Gateway.GRPCUserClient.RetryBackoffMultiplier,
			KeepaliveTime:          cfg.Gateway.GRPCUserClient.KeepaliveTime,
			KeepaliveTimeout:       cfg.Gateway.GRPCUserClient.KeepaliveTimeout,
			KeepalivePermitWithout: cfg.Gateway.GRPCUserClient.KeepalivePermitWithout,
		})
	if err != nil {
		slog.Error("failed to create gRPC client",
			slog.String("error", err.Error()),
			slog.String("type", "user"))
		os.Exit(1)
	}
	defer userClient.Close()

	/* REST Middlewares */
	metrics := middleware.NewMetrics()
	authMW := middleware.NewAuth(authClient)
	loggerMW := middleware.NewLogger()

	rateLimiter := middleware.NewRateLimiter(middleware.RateLimiterConfig{
		Client:       rlRedisClient.Client,
		RoutesAll:    cfg.Gateway.RateLimiter.All.Limits,
		RoutesIP:     cfg.Gateway.RateLimiter.IP.Limits,
		RoutesAllExp: cfg.Gateway.RateLimiter.All.Expires,
		RoutesIPExp:  cfg.Gateway.RateLimiter.IP.Expires,
	})

	/* REST сервер */
	restServer := delivery.NewREST(
		delivery.Dependencies{
			AuthClient:  authClient,
			UserClient:  userClient,
			Metrics:     metrics,
			RateLimiter: rateLimiter,
			AuthMW:      authMW,
			LoggerMW:    loggerMW},
		delivery.RESTServerConfig{
			Addr:         cfg.Gateway.Addr,
			ReadTimeout:  cfg.Gateway.TimeoutRead,
			WriteTimeout: cfg.Gateway.TimeoutWrite,
			IdleTimeout:  cfg.Gateway.TimeoutIdle,
			GSTime:       cfg.Gateway.ShutdownTimeout,
		})

	go func() {
		if err := restServer.Run(); err != nil {
			slog.Error("failed to run REST server",
				slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	<-ctx.Done()

	slog.Info("got signal-notify")

	if err = restServer.Stop(); err != nil {
		slog.Warn("failed to gracefully shutdown",
			slog.String("error", err.Error()))
		return
	}
	slog.Info("gracefully shutdown")
}
