package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Middleware interface {
	Middleware(next http.Handler) http.Handler
}

func SetupMiddleware(r *chi.Mux, Logger Middleware, rateLimiter Middleware, metrics Middleware) {
	r.Use(middleware.Recoverer)
	r.Use(Logger.Middleware)
	r.Use(middleware.Timeout(time.Second * 15))
	r.Use(metrics.Middleware)
	r.Use(rateLimiter.Middleware)
}
