package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type limitResult struct {
	remaining int64
	exceeded  bool
}

type RateLimiterConfig struct {
	Client       *redis.Client
	RoutesAll    map[string]int
	RoutesIP     map[string]int
	RoutesAllExp map[string]time.Duration
	RoutesIPExp  map[string]time.Duration
}

type RateLimiter struct {
	client       *redis.Client
	routesAll    map[string]int
	routesIP     map[string]int
	RoutesAllExp map[string]time.Duration
	RoutesIPExp  map[string]time.Duration
}

func NewRateLimiter(cfg RateLimiterConfig) *RateLimiter {
	return &RateLimiter{
		client:       cfg.Client,
		routesAll:    cfg.RoutesAll,
		routesIP:     cfg.RoutesIP,
		RoutesAllExp: cfg.RoutesAllExp,
		RoutesIPExp:  cfg.RoutesIPExp,
	}
}

func (rl *RateLimiter) checkLimit(ctx context.Context, key string, limit int, expiry time.Duration, logger *slog.Logger) (*limitResult, error) {
	current, err := rl.client.Incr(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	remaining := int64(limit) - current
	if remaining < 0 {
		remaining = 0
	}

	if current == 1 {
		if err := rl.client.Expire(ctx, key, expiry).Err(); err != nil {
			logger.Error("expire not set",
				slog.String("error", err.Error()),
				slog.String("key", key))
			if delErr := rl.client.Del(ctx, key).Err(); delErr != nil {
				logger.Error("failed to delete bad key",
					slog.String("error", delErr.Error()),
					slog.String("key", key))
			}
			return nil, err
		}
	}

	return &limitResult{
		remaining: remaining,
		exceeded:  remaining == 0,
	}, nil
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		allLimit := rl.routesAll[path]
		ipLimit := rl.routesIP[path]
		if allLimit == 0 && ipLimit == 0 {
			next.ServeHTTP(w, r)
			return
		}

		logger := GetLogger(r.Context())

		var allRemaining int64 = -1
		var ipRemaining int64 = -1
		var allExceeded, ipExceeded bool

		if allLimit > 0 {
			key := "rl:" + path + ":all"
			result, err := rl.checkLimit(r.Context(), key, allLimit, rl.RoutesAllExp[path], logger)
			if err != nil {
				logger.Error("rate limit check failed",
					slog.String("type", "all"),
					slog.String("error", err.Error()))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			allRemaining = result.remaining
			allExceeded = result.exceeded
		}

		if ipLimit > 0 {
			key := "rl:" + path + ":" + r.RemoteAddr
			result, err := rl.checkLimit(r.Context(), key, ipLimit, rl.RoutesIPExp[path], logger)
			if err != nil {
				logger.Error("rate limit check failed",
					slog.String("type", "ip"),
					slog.String("error", err.Error()))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			ipRemaining = result.remaining
			ipExceeded = result.exceeded
		}

		if allLimit > 0 {
			w.Header().Set("X-RateLimit-Global-Limit", strconv.Itoa(allLimit))
			if allRemaining >= 0 {
				w.Header().Set("X-RateLimit-Global-Remaining", strconv.FormatInt(allRemaining, 10))
			}
		}

		if ipLimit > 0 {
			w.Header().Set("X-RateLimit-IP-Limit", strconv.Itoa(ipLimit))
			if ipRemaining >= 0 {
				w.Header().Set("X-RateLimit-IP-Remaining", strconv.FormatInt(ipRemaining, 10))
			}
		}

		if allExceeded || ipExceeded {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
