package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type contextKey1 string

const (
	TraceIDKey contextKey1 = "trace_id"
	LoggerKey  contextKey1 = "logger"
)

type Logger struct {
}

func NewLogger() *Logger {
	return &Logger{}
}

func (l *Logger) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		traceID := r.Header.Get("X-Trace-Id")
		if traceID == "" {
			traceID = uuid.New().String()
		}

		requestID := uuid.New().String()

		logger := slog.With(
			slog.String("trace_id", traceID),
			slog.String("request_id", requestID),
			slog.String("path", r.URL.Path),
			slog.String("method", r.Method),
			slog.String("client_ip", r.RemoteAddr),
			slog.String("user_agent", r.UserAgent()),
		)

		ctx := context.WithValue(r.Context(), TraceIDKey, traceID)
		ctx = context.WithValue(ctx, LoggerKey, logger)

		w.Header().Set("X-Trace-Id", traceID)
		w.Header().Set("X-Request-Id", requestID)

		logger.Info("request started")

		wrapped := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(wrapped, r.WithContext(ctx))

		logger.Info("request completed",
			slog.Int("status", wrapped.status),
			slog.Duration("duration", time.Since(start)),
		)
	})
}

func (l *Logger) GetLogger(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(LoggerKey).(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}

type responseWriter struct {
	http.ResponseWriter
	status       int
	bytesWritten int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += n
	return n, err
}
