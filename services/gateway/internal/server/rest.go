package rest

import (
	"context"
	"net/http"
	"os/signal"
	"restaurant/pkg/logger"
	"syscall"
	"time"
)

type RESTServer struct {
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

func New(handler http.Handler, cfg RESTServerConfig) *RESTServer {
	return &RESTServer{
		server: &http.Server{
			Addr:         cfg.Addr,
			Handler:      handler,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			IdleTimeout:  cfg.IdleTimeout,
		},
		gsTime: cfg.GSTime,
	}
}

func (r *RESTServer) Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	go func() {
		logger.Info.Printf("сервер слушает на %s", r.server.Addr)
		if err := r.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error.Printf("ошибка HTTP сервера: %v", err)
		}
	}()

	<-ctx.Done()

	logger.Info.Println("завершение HTTP сервера...")

	ctx, cancel = context.WithTimeout(context.Background(), r.gsTime)
	defer cancel()

	return r.Shutdown(ctx)
}

func (r *RESTServer) Shutdown(ctx context.Context) error {
	return r.server.Shutdown(ctx)
}
