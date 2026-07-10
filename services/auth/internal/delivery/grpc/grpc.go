package delivery

import (
	"fmt"
	"log/slog"
	"net"
	"time"

	authv3 "restaurant/api/proto/auth/v3"
	"restaurant/pkg/interceptors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

type gRPCServer struct {
	addr         string
	creds        credentials.TransportCredentials
	tokenService HandlerToken
	gsTime       time.Duration
	server       *grpc.Server
}

func NewGRPCServer(creds credentials.TransportCredentials, tokenService HandlerToken, addr string, gsTime time.Duration) *gRPCServer {
	return &gRPCServer{
		addr:         addr,
		creds:        creds,
		tokenService: tokenService,
		gsTime:       gsTime,
	}
}

func (s *gRPCServer) Run() error {
	handler := NewHandler(s.tokenService)

	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("ошибка создания слушателя: %v", err)
	}

	opts := []grpc.ServerOption{
		grpc.Creds(s.creds),
		grpc.MaxRecvMsgSize(1024 * 1024),
		grpc.MaxSendMsgSize(1024 * 1024),
		grpc.ConnectionTimeout(10 * time.Second),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: 5 * time.Minute,
			Time:              1 * time.Minute,
			Timeout:           20 * time.Second,
		}),
		grpc.ChainUnaryInterceptor(
			interceptors.Logger(),
			interceptors.Recoverer(),
		),
	}

	s.server = grpc.NewServer(opts...)
	authv3.RegisterAuthServiceServer(s.server, handler)

	slog.Info("gRPC server started",
		"address", s.addr)
	if err := s.server.Serve(lis); err != nil {
		return fmt.Errorf("ошибка gRPC сервера: %v", err)
	}
	return nil
}

func (s *gRPCServer) Stop() error {
	done := make(chan struct{})

	go func() {
		s.server.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-time.After(s.gsTime):
		s.server.Stop()
		return fmt.Errorf("таймаут graceful shutdown, принудительная остановка")
	}
}
