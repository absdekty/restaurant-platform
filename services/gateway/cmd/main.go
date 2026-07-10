package main

import (
	"context"
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
	"time"
)

func main() {
	/* Конфиг */
	config.Load("gateway")

	/* Логгер */
	logger.SetupLogger("dev", "gateway")

	/* TLS Clients */
	clientAuthCreds, err := tls.ClientCreds(
		config.Get[string]("CA_CERT", "certs/ca/ca-cert.pem"),
		config.Get[string]("GATEWAY_CERT", "certs/gateway/server-cert.pem"),
		config.Get[string]("GATEWAY_KEY", "certs/gateway/server-key.pem"),
		"auth")
	if err != nil {
		slog.Error("failed to create mTLS",
			slog.String("error", err.Error()),
			slog.String("type", "auth_client"))
		os.Exit(1)

	}
	clientUserCreds, err := tls.ClientCreds(
		config.Get[string]("CA_CERT", "certs/ca/ca-cert.pem"),
		config.Get[string]("GATEWAY_CERT", "certs/gateway/server-cert.pem"),
		config.Get[string]("GATEWAY_KEY", "certs/gateway/server-key.pem"),
		"user")
	if err != nil {
		slog.Error("failed to create mTLS",
			slog.String("error", err.Error()),
			slog.String("type", "user_client"))
		os.Exit(1)
	}

	/* gRPC auth-client */
	authClient, err := client.NewAuthClient(clientAuthCreds, config.Get[string]("AUTH_GRPC_LISTENER", "localhost:50051"))
	if err != nil {
		slog.Error("failed to create gRPC client",
			slog.String("error", err.Error()),
			slog.String("type", "auth"))
		os.Exit(1)
	}
	defer authClient.Close()

	/* gRPC user-client */
	userClient, err := client.NewUserClient(clientUserCreds, config.Get[string]("USER_GRPC_LISTENER", "localhost:50052"))
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
		float64(config.Get[int]("RATE_LIMITER_RPS_TOTAL", 100)),
		config.Get[int]("RATE_LIMITER_RPS_BURST", 200))
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
			Addr:         config.Get[string]("GATEWAY_ADDR", "localhost:8080"),
			ReadTimeout:  time.Duration(config.Get[int]("GATEWAY_TIMEOUT_READ", 10)) * time.Second,
			WriteTimeout: time.Duration(config.Get[int]("GATEWAY_TIMEOUT_WRITE", 10)) * time.Second,
			IdleTimeout:  time.Duration(config.Get[int]("GATEWAY_TIMEOUT_IDLE", 30)) * time.Second,
			GSTime:       time.Duration(config.Get[int]("GATEWAY_SHUTDOWN", 30)) * time.Second,
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
