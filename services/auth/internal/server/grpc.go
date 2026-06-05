package server

import (
	"context"
	"net"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"restaurant/pkg/logger"
)

type GRPCServer struct {
	server   *grpc.Server
	listener net.Listener
	gsTime   time.Duration
}

func NewGRPCServer(grpcServer *grpc.Server, listener net.Listener, gsTime time.Duration) *GRPCServer {
	return &GRPCServer{
		server:   grpcServer,
		listener: listener,
		gsTime:   gsTime,
	}
}

func (s *GRPCServer) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info.Printf("gRPC сервер слушает на %s", s.listener.Addr())
		if err := s.server.Serve(s.listener); err != nil {
			logger.Error.Printf("ошибка gRPC сервера: %v", err)
		}
	}()

	<-ctx.Done()
	logger.Info.Println("завершение gRPC сервера...")

	return s.GracefulStop()
}

func (s *GRPCServer) GracefulStop() error {
	done := make(chan struct{})

	go func() {
		s.server.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		logger.Info.Println("gRPC сервер остановлен gracefully")
		return nil
	case <-time.After(s.gsTime):
		logger.Warn.Printf("таймаут graceful shutdown, принудительная остановка")
		s.server.Stop()
		return nil
	}
}
