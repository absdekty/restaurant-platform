package delivery

import (
	"encoding/json"
	"errors"
	"net/http"
	"restaurant/pkg/logger"
	"restaurant/services/gateway/internal/model"
)

type Handler struct {
	rateLimiter *RateLimiter
	metrics     *Metrics
	auth        *Auth
	user        UserService
}

func NewHandler(rateLimiter *RateLimiter, metrics *Metrics, auth *Auth, user UserService) *Handler {
	return &Handler{
		rateLimiter: rateLimiter,
		metrics:     metrics,
		auth:        auth,
		user:        user}
}

/* Сервис доступен? */
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

/* Метрики */
func (h *Handler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total_requests":   h.metrics.GetTotalRequests(),
		"active_requests":  h.metrics.GetActiveRequests(),
		"errors_total":     h.metrics.GetErrorsTotal(),
		"errors_by_status": h.metrics.GetErrorsByStatus()})
}

/* Регистрация пользователя */
func (h *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Info.Printf("decoder: %v", err)
		http.Error(w, "invalid JSON format", http.StatusBadRequest)
		return
	}

	userID, err := h.user.RegisterUser(r.Context(), req.Name, req.Password)
	if err != nil {
		if errors.Is(err, model.ErrUserInvalidRegisterDetails) {
			logger.Info.Printf("register[reg details]: %v", err)
			http.Error(w, "invalid register details", http.StatusBadRequest)
			return
		}

		if errors.Is(err, model.ErrUserAlreadyExists) {
			logger.Info.Printf("register[exist]: %v", err)
			http.Error(w, "user already exists", http.StatusConflict)
			return
		}

		logger.Error.Printf("register[internal]: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(RegisterResponse{UserID: userID})
}

/* Авторизация пользователя */
func (h *Handler) LoginUser(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Info.Printf("decoder: %v", err)
		http.Error(w, "invalid JSON format", http.StatusBadRequest)
		return
	}

	accessToken, refreshToken, refreshTTL, err := h.user.LoginUser(r.Context(), req.Name, req.Password)
	if err != nil {
		if errors.Is(err, model.ErrUserInvalidCredentials) {
			logger.Info.Printf("login[credentials]: %v", err)
			http.Error(w, "invalid credentials", http.StatusBadRequest)
			return
		}

		if errors.Is(err, model.ErrUserNotFound) {
			logger.Info.Printf("login[not exists]: %v", err)
			http.Error(w, "user already exists", http.StatusNotFound)
			return
		}

		logger.Error.Printf("login[internal]: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   int(refreshTTL),
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken})
}

/* Рефреш пары токенов */
func (h *Handler) RefreshTokens(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		logger.Info.Printf("cookie: %v", err)
		http.Error(w, "refresh token required", http.StatusUnauthorized)
		return
	}

	cookieRefreshToken := cookie.Value

	accessToken, refreshToken, refreshTTL, err := h.auth.RefreshTokens(r.Context(), cookieRefreshToken)
	if err != nil {
		if errors.Is(err, model.ErrInvalidToken) {
			logger.Info.Printf("refresh[invalid token]: %v", err)
			http.Error(w, "token revoked or expired or not found", http.StatusForbidden)
			return
		}

		logger.Error.Printf("refresh[internal]: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   int(refreshTTL),
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(RefreshResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken})
}

/* Логаут пользователя */
func (h *Handler) LogoutUser(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		logger.Info.Printf("cookie: %v", err)
		http.Error(w, "refresh token required", http.StatusUnauthorized)
		return
	}

	cookieRefreshToken := cookie.Value

	if err := h.auth.RevokeRefreshToken(r.Context(), cookieRefreshToken); err != nil {
		logger.Error.Printf("revoke: %v", err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:   "refresh_token",
		MaxAge: -1,
	})

	w.WriteHeader(http.StatusOK)
}
