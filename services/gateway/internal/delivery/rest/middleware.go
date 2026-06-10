package delivery

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"time"
)

func setupMiddleware(r *chi.Mux, handler *Handler) {
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Timeout(time.Second * 15))
	r.Use(middleware.RealIP)
	r.Use(handler.rateLimiter.Middleware)
	r.Use(handler.metrics.Middleware)
}
