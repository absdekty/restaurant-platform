package metrics

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MetricsConfig struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	Shutdown     time.Duration
}

type Metrics struct {
	server *http.Server
	gsTime time.Duration
}

func New(cfg MetricsConfig) *Metrics {
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:         cfg.Addr,
		Handler:      metricsMux,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	return &Metrics{server: server, gsTime: cfg.Shutdown}
}

func (m *Metrics) Run() error {
	slog.Info("Metrics server started",
		"address", m.server.Addr)

	if err := m.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("ошибка HTTP сервера: %v", err)
	}
	return nil
}

func (m *Metrics) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), m.gsTime)
	defer cancel()

	return m.server.Shutdown(ctx)
}
