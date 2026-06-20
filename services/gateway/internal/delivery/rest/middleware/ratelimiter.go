package middleware

import (
	"golang.org/x/time/rate"
	"net/http"
)

type RateLimiter struct {
	limiter *rate.Limiter
}

func NewRateLimiter(rps float64, burst int) *RateLimiter {
	return &RateLimiter{
		limiter: rate.NewLimiter(rate.Limit(rps), burst),
	}
}

func (r *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if !r.limiter.Allow() {
			http.Error(w, `{"error": "too many requests"}`, http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, req)
	})
}
