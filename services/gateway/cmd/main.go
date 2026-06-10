package main

import (
	"context"
	"os/signal"
	"restaurant/pkg/config"
	"restaurant/pkg/logger"
	"restaurant/services/gateway/internal/client"
	"restaurant/services/gateway/internal/delivery/rest"
	"syscall"
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
	defer authClient.Close()

	/* REST сервер */
	restServer := delivery.NewREST(authClient, delivery.RESTServerConfig{
		Addr:         config.Get[string]("GATEWAY_HOST", "localhost") + ":" + config.Get[string]("GATEWAY_PORT", "8080"),
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
