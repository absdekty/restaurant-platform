package main

import (
	"restaurant/pkg/config"
	"restaurant/pkg/logger"
	"restaurant/services/gateway/internal/client"
	"restaurant/services/gateway/internal/delivery"
	"restaurant/services/gateway/internal/server"
	"time"
)

func main() {
	/* Конфиг */
	config.Load("gateway")

	/* Логгер */
	logger.Init("Gateway")

	/* gRPC Client */
	authClient, err := client.NewAuthClient(config.Get[string]("GATEWAY_GRPC_CLIENT", "localhost:50051"))
	if err != nil {
		logger.Error.Printf("ошибка gRPC клиента: %v", err)
	}

	/* REST */
	rateLimiter := delivery.NewRateLimiter(100, 200)
	metrics := delivery.NewMetrics()

	restHandler := delivery.NewHandler(authClient, rateLimiter, metrics)
	restRouter := delivery.NewRouter(restHandler)

	/* Создание, запуск сервера */
	restServer := rest.New(restRouter, rest.RESTServerConfig{
		Addr:         config.Get[string]("GATEWAY_HOST", "localhost") + ":" + config.Get[string]("GATEWAY_PORT", "8080"),
		ReadTimeout:  time.Duration(config.Get[int]("GATEWAY_TIMEOUT_READ", 10)) * time.Second,
		WriteTimeout: time.Duration(config.Get[int]("GATEWAY_TIMEOUT_WRITE", 10)) * time.Second,
		IdleTimeout:  time.Duration(config.Get[int]("GATEWAY_TIMEOUT_IDLE", 30)) * time.Second,
		GSTime:       time.Duration(config.Get[int]("GATEWAY_SHUTDOWN", 30)) * time.Second,
	})

	if err := restServer.Run(); err != nil {
		logger.Error.Printf("ошибка остановки сервера: %v", err)
	}
}
