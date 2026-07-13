package interceptors

import (
	"context"
	"log/slog"
	"restaurant/pkg/models"
	"runtime/debug"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func Logger() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		traceID := extractTraceIDFromMetadata(ctx)
		logger := slog.With(slog.String("trace_id", traceID))
		ctx = context.WithValue(ctx, models.TraceIDKey, traceID)
		ctx = context.WithValue(ctx, models.LoggerKey, logger)

		resp, err := handler(ctx, req)

		if err != nil {
			logger.Info("gRPC call failed",
				slog.String("method", info.FullMethod),
				slog.Duration("duration", time.Since(start)),
				slog.String("error", err.Error()))
		} else {
			logger.Info("gRPC call completed",
				slog.String("method", info.FullMethod),
				slog.Duration("duration", time.Since(start)))
		}
		return resp, err
	}
}

func Recoverer() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				logger := ExtractLoggerFromContext(ctx)
				logger.Warn("panic recovered",
					slog.Any("panic", r),
					slog.String("stack", string(debug.Stack())),
				)
				err = status.Errorf(codes.Internal, "internal error")
			}
		}()
		return handler(ctx, req)
	}
}

func TraceClient() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{},
		cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		traceID := ""
		if v := ctx.Value(models.TraceIDKey); v != nil {
			traceID = v.(string)
		}

		if traceID != "" {
			md := metadata.Pairs("x-trace-id", traceID)
			ctx = metadata.NewOutgoingContext(ctx, md)
		}

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func extractTraceIDFromMetadata(ctx context.Context) string {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if values := md.Get("x-trace-id"); len(values) >= 1 {
			return values[0]
		}
	}
	return ""
}

func ExtractLoggerFromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(models.LoggerKey).(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}
