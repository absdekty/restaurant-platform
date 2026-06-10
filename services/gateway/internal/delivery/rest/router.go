package delivery

import (
	"github.com/go-chi/chi/v5"
)

func NewRouter(handler *Handler) *chi.Mux {
	r := chi.NewRouter()

	setupMiddleware(r, handler)

	r.Get("/health", handler.HealthCheck) // GET /health - возвращает StatusOK
	r.Get("/metrics", handler.GetMetrics) // GET /metrics - получить актуальные метрики

	return r
}
