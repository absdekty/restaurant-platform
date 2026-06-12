package delivery

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"restaurant/pkg/logger"
	"time"
)

func loggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	logger.Info.Printf("method: %s, duration: %v, err: %v", info.FullMethod, time.Since(start), err)
	return resp, err
}

func recoveryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			logger.Warn.Printf("поймана паника: %v", r)
			err = status.Errorf(codes.Internal, "internal error")
		}
	}()
	return handler(ctx, req)
}
