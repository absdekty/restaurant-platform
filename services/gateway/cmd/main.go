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
	authClient, err := client.NewAuthClient(clientAuthCreds, cfg.AuthAddr)
	if err != nil {
		slog.Error("failed to create gRPC client",
			slog.String("error", err.Error()),
			slog.String("type", "auth"))
		os.Exit(1)
	}
	defer authClient.Close()

	/* gRPC user-client */
	userClient, err := client.NewUserClient(clientUserCreds, cfg.UserAddr)
	if err != nil {
		slog.Error("failed to create gRPC client",
			slog.String("error", err.Error()),
			slog.String("type", "user"))
		os.Exit(1)
	}
	defer userClient.Close()

	/* REST Middlewares */
	metrics := middleware.NewMetrics()
	rateLimiter := middleware.NewRateLimiter(
		float64(cfg.Gateway.RPS),
		cfg.Gateway.Burst)
	authMW := middleware.NewAuth(authClient)
	loggerMW := middleware.NewLogger()

	/* REST сервер */
	restServer := delivery.NewREST(
		delivery.Dependencies{
			AuthClient:  authClient,
			UserClient:  userClient,
			Metrics:     metrics,
			RateLimiter: rateLimiter,
			AuthMW:      authMW,
			LoggerMW:    loggerMW,
			Logger:      loggerMW},
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
