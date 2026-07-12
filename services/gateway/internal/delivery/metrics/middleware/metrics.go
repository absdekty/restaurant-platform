package middleware

import (
	"net/http"
	"restaurant/services/gateway/pkg/httputil"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Metrics — сборщик метрик для Prometheus
type Metrics struct {
	requestsTotal   *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
	activeRequests  prometheus.Gauge
	errorsTotal     *prometheus.CounterVec
}

// NewMetrics — конструктор, регистрирует метрики
func NewMetrics() *Metrics {
	m := &Metrics{
		requestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		requestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "Duration of HTTP requests in seconds",
				Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
			},
			[]string{"method", "path"},
		),
		activeRequests: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "http_active_requests",
				Help: "Current number of active requests",
			},
		),
		errorsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_errors_total",
				Help: "Total number of HTTP errors by status",
			},
			[]string{"status"},
		),
	}

	// Регистрируем все метрики
	prometheus.MustRegister(
		m.requestsTotal,
		m.requestDuration,
		m.activeRequests,
		m.errorsTotal,
	)

	return m
}

// Middleware — собирает метрики по всем запросам
func (m *Metrics) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		path := r.URL.Path

		// Активные запросы
		m.activeRequests.Inc()
		defer m.activeRequests.Dec()

		rw := httputil.NewSafeResponseWriter(w)

		// Выполняем запрос
		next.ServeHTTP(rw, r)

		status := rw.StatusCode
		statusText := strconv.Itoa(status)

		// Счётчик запросов
		m.requestsTotal.With(prometheus.Labels{
			"method": r.Method,
			"path":   path,
			"status": statusText,
		}).Inc()

		// Гистограмма длительности
		m.requestDuration.With(prometheus.Labels{
			"method": r.Method,
			"path":   path,
		}).Observe(time.Since(start).Seconds())

		// Счётчик ошибок (4xx и 5xx)
		if status >= 400 {
			m.errorsTotal.With(prometheus.Labels{
				"status": statusText,
			}).Inc()
		}
	})
}
