package delivery

import (
	"context"
	"fmt"
	"net/http"
	"restaurant/pkg/logger"
	"time"
)

type RESTServer struct {
	authClient AuthService
	server     *http.Server
	gsTime     time.Duration
}

type RESTServerConfig struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	GSTime       time.Duration
}

func NewREST(authClient AuthService, cfg RESTServerConfig) *RESTServer {
	return &RESTServer{
		authClient: authClient,
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
	rateLimiter := NewRateLimiter(100, 200) // Rate Limiter
	metrics := NewMetrics()                 // Metrics
	auth := NewAuth(r.authClient)           // Auth middleware

	restHandler := NewHandler(rateLimiter, metrics, auth)
	r.server.Handler = NewRouter(restHandler)

	logger.Info.Printf("сервер слушает на: %s", r.server.Addr)

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
