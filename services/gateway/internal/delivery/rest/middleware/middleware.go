package middleware

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
	"time"
)

type Middleware interface {
	Middleware(next http.Handler) http.Handler
}

func SetupMiddleware(r *chi.Mux, rateLimiter Middleware, metrics Middleware) {
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Timeout(time.Second * 15))
	r.Use(middleware.RealIP)
	r.Use(rateLimiter.Middleware)
	r.Use(metrics.Middleware)
}
