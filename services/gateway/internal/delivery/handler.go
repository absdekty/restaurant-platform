package delivery

import (
	"context"
	"encoding/json"
	"net/http"
)

type AuthService interface {
	ValidateToken(ctx context.Context, token string) (string, error)
}

type Handler struct {
	authService AuthService
	rateLimiter *RateLimiter
	metrics     *Metrics
}

func NewHandler(authService AuthService, rateLimiter *RateLimiter, metrics *Metrics) *Handler {
	return &Handler{authService: authService, rateLimiter: rateLimiter, metrics: metrics}
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func (h *Handler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total_requests":   h.metrics.GetTotalRequests(),
		"active_requests":  h.metrics.GetActiveRequests(),
		"errors_total":     h.metrics.GetErrorsTotal(),
		"errors_by_status": h.metrics.GetErrorsByStatus()})
}
