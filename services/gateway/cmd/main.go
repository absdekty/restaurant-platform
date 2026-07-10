package main

import (
	"context"
	"os/signal"
	"restaurant/pkg/config"
	"restaurant/pkg/logger"
	"restaurant/pkg/tls"
	"restaurant/services/gateway/internal/client"
	"restaurant/services/gateway/internal/delivery/rest"
	"restaurant/services/gateway/internal/delivery/rest/middleware"
	"syscall"
	"time"
)

func main() {
	/* Конфиг */
	config.Load("gateway")

	/* Логгер */
	logger.Init("Gateway")

	/* TLS Clients */
	clientAuthCreds, err := tls.ClientCreds(
		config.Get[string]("CA_CERT", "certs/ca/ca-cert.pem"),
		config.Get[string]("GATEWAY_CERT", "certs/gateway/server-cert.pem"),
		config.Get[string]("GATEWAY_KEY", "certs/gateway/server-key.pem"),
		"auth")
	if err != nil {
		logger.Error.Printf("ошибка создания mTLS[auth]: %v", err)
	}

	clientUserCreds, err := tls.ClientCreds(
		config.Get[string]("CA_CERT", "certs/ca/ca-cert.pem"),
		config.Get[string]("GATEWAY_CERT", "certs/gateway/server-cert.pem"),
		config.Get[string]("GATEWAY_KEY", "certs/gateway/server-key.pem"),
		"user")
	if err != nil {
		logger.Error.Printf("ошибка создания mTLS: %v[user]", err)
	}

	/* gRPC auth-client */
	authClient, err := client.NewAuthClient(clientAuthCreds, config.Get[string]("AUTH_GRPC_LISTENER", "localhost:50051"))
	if err != nil {
		logger.Error.Printf("ошибка gRPC клиента: %v", err)
	}
	defer authClient.Close()

	/* gRPC user-client */
	userClient, err := client.NewUserClient(clientUserCreds, config.Get[string]("USER_GRPC_LISTENER", "localhost:50052"))
	if err != nil {
		logger.Error.Printf("ошибка gRPC клиента: %v", err)
	}
	defer userClient.Close()

	/* REST Middlewares */
	metrics := middleware.NewMetrics()
	rateLimiter := middleware.NewRateLimiter(
		float64(config.Get[int]("RATE_LIMITER_RPS_TOTAL", 100)),
		config.Get[int]("RATE_LIMITER_RPS_BURST", 200))
	authMW := middleware.NewAuth(authClient)

	/* REST сервер */
	restServer := delivery.NewREST(
		delivery.Dependencies{
			AuthClient:  authClient,
			UserClient:  userClient,
			Metrics:     metrics,
			RateLimiter: rateLimiter,
			AuthMW:      authMW},
		delivery.RESTServerConfig{
			Addr:         config.Get[string]("GATEWAY_ADDR", "localhost:8080"),
			ReadTimeout:  time.Duration(config.Get[int]("GATEWAY_TIMEOUT_READ", 10)) * time.Second,
			WriteTimeout: time.Duration(config.Get[int]("GATEWAY_TIMEOUT_WRITE", 10)) * time.Second,
			IdleTimeout:  time.Duration(config.Get[int]("GATEWAY_TIMEOUT_IDLE", 30)) * time.Second,
			GSTime:       time.Duration(config.Get[int]("GATEWAY_SHUTDOWN", 30)) * time.Second,
		})

	go func() {
		if err := restServer.Run(); err != nil {
			logger.Error.Printf("ошибка остановки сервера: %v", err)
		}
	}()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	<-ctx.Done()

	logger.Info.Println("получен сигнал завершение сервера, останавливаем..")

	if err = restServer.Stop(); err != nil {
		logger.Warn.Printf("программа завершилась не gracefully: %v", err)
		return
	}
	logger.Info.Println("программа завершилась gracefully")
}
