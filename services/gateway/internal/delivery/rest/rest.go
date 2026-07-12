package delivery

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	mw "restaurant/services/gateway/internal/delivery/rest/middleware"
	"time"
)

type MetricsREST interface {
	MetricsHandler
	mw.Middleware
}

type RateLimiterREST interface {
	mw.Middleware
}

type AuthMiddlewareREST interface {
	mw.Middleware
}

type LoggerMiddlewareREST interface {
	mw.Middleware
}

type Dependencies struct {
	AuthClient  AuthHandler
	UserClient  UserHandler
	Metrics     MetricsREST
	RateLimiter RateLimiterREST
	AuthMW      AuthMiddlewareREST
	LoggerMW    LoggerMiddlewareREST
}

type RESTServer struct {
	deps   Dependencies
	server *http.Server
	gsTime time.Duration
}

type RESTServerConfig struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	GSTime       time.Duration
}

func NewREST(deps Dependencies, cfg RESTServerConfig) *RESTServer {
	return &RESTServer{
		deps: deps,
		server: &http.Server{
			Addr:         cfg.Addr,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			IdleTimeout:  cfg.IdleTimeout,
		},
		gsTime: cfg.GSTime,
	}
}

func (r *RESTServer) Run() error {
	restHandler := NewHandler(r.deps.Metrics, r.deps.AuthClient, r.deps.UserClient)
	r.server.Handler = NewRouter(restHandler, r.deps.LoggerMW, r.deps.RateLimiter, r.deps.Metrics, r.deps.AuthMW)

	slog.Info("gRPC server started",
		"address", r.server.Addr)

	if err := r.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("ошибка HTTP сервера: %v", err)
	}
	return nil
}

func (r *RESTServer) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), r.gsTime)
	defer cancel()

	return r.server.Shutdown(ctx)
}
