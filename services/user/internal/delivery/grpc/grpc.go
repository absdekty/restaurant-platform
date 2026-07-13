package delivery

import (
	"fmt"
	"log/slog"
	"net"
	"time"

	userv2 "restaurant/api/proto/user/v2"
	"restaurant/pkg/interceptors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

type gRPCServer struct {
	addr        string
	creds       credentials.TransportCredentials
	userService UserService
	gsTime      time.Duration
	server      *grpc.Server
	config      OptionConfig
}

type OptionConfig struct {
	MaxReceivedSize   int
	MaxSendSize       int
	ConnectionTimeout time.Duration
	MaxConnectionIdle time.Duration
	KeepAliveTime     time.Duration
	KeepAliveTimeout  time.Duration
}

func NewGRPCServer(creds credentials.TransportCredentials, userService UserService, addr string, gsTime time.Duration, config OptionConfig) *gRPCServer {
	return &gRPCServer{
		addr:        addr,
		creds:       creds,
		userService: userService,
		gsTime:      gsTime,
		config:      config,
	}
}

func (s *gRPCServer) Run() error {
	handler := NewHandler(s.userService)

	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("ошибка создания слушателя: %v", err)
	}

	opts := []grpc.ServerOption{
		grpc.Creds(s.creds),
		grpc.MaxRecvMsgSize(s.config.MaxReceivedSize),
		grpc.MaxSendMsgSize(s.config.MaxSendSize),
		grpc.ConnectionTimeout(s.config.ConnectionTimeout),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: s.config.MaxConnectionIdle,
			Time:              s.config.KeepAliveTime,
			Timeout:           s.config.KeepAliveTimeout,
		}),
		grpc.ChainUnaryInterceptor(
			interceptors.Logger(),
			interceptors.Recoverer(),
		),
	}

	s.server = grpc.NewServer(opts...)
	userv2.RegisterUserServiceServer(s.server, handler)

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
