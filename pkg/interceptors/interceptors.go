package interceptor

import (
	"context"
	"log/slog"
	"runtime/debug"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Logger() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		if err != nil {
			slog.Info("gRPC call failed",
				slog.String("method", info.FullMethod),
				slog.Duration("duration_ms", time.Since(start)),
				slog.String("error", err.Error()))
		} else {
			slog.Info("gRPC call completed",
				slog.String("method", info.FullMethod),
				slog.Duration("duration_ms", time.Since(start)))
		}
		return resp, err
	}
}

func Recoverer() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				slog.Warn("panic recovered",
					slog.Any("panic", r),
					slog.String("stack", string(debug.Stack())),
				)
				err = status.Errorf(codes.Internal, "internal error")
			}
		}()
		return handler(ctx, req)
	}
}
