package main

import (
	"google.golang.org/grpc"
	"net"
	authv1 "restaurant/api/proto/auth/v1"
	"restaurant/pkg/config"
	"restaurant/pkg/logger"
	"restaurant/services/auth/internal/delivery/grpc"
	"restaurant/services/auth/internal/server"
	"restaurant/services/auth/internal/service"
	"restaurant/services/auth/internal/storage/sqlite3"
	"time"
)

func main() {
	/* Конфиг */
	config.Load("auth")

	/* Логгер */
	logger.Init("Auth")

	/* Хранилище refresh-token */
	storage, err := sqlite3.New("tokens.db")
	if err != nil {
		logger.Error.Printf("failed to create storage: %v", err)
	}
	defer storage.Close()

	/* JWT-Service */
	jwtService := service.NewJWT(
		config.Get[string]("AUTH_SECRET_KEY", "default-secret-key-min-32-chars"),
		time.Duration(config.Get[int]("AUTH_ACCESS_TTL", 15))*time.Minute,
		time.Duration(config.Get[int]("AUTH_REFRESH_TTL", 7))*time.Hour*24,
		storage)

	/* gRPC Handler */
	handler := delivery.NewHandler(jwtService)

	/* gRPC Listener */
	lis, err := net.Listen("tcp", config.Get[string]("AUTH_GRPC_LISTENER", ":50051"))

	/* gRPC Server */
	gRPCServer := grpc.NewServer()
	authv1.RegisterAuthServiceServer(gRPCServer, handler)
	srv := server.NewGRPCServer(
		gRPCServer, lis,
		time.Duration(config.Get[int]("AUTH_SHUTDOWN", 30))*time.Second)

	if err := srv.Run(); err != nil {
		logger.Error.Printf("ошибка остановки сервера: %v", err)
	}
}
