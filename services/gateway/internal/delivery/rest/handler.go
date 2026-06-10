package delivery

import (
	"encoding/json"
	"net/http"
)

type Handler struct {
	rateLimiter *RateLimiter
	metrics     *Metrics
	auth        *Auth
}

func NewHandler(rateLimiter *RateLimiter, metrics *Metrics, auth *Auth) *Handler {
	return &Handler{
		rateLimiter: rateLimiter,
		metrics:     metrics,
		auth:        auth}
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
