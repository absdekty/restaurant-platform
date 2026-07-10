package delivery

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"restaurant/services/gateway/internal/model"
)

type MetricsHandler interface {
	GetTotalRequests() uint64
	GetActiveRequests() int64
	GetErrorsTotal() uint64
	GetErrorsByStatus() map[string]uint64
}

type AuthHandler interface {
	RefreshTokens(ctx context.Context, refreshtoken string) (string, string, int32, error)
	RevokeRefreshToken(ctx context.Context, refreshtoken string) error
}

type UserHandler interface {
	RegisterUser(ctx context.Context, name, password string) (string, error)
	LoginUser(ctx context.Context, name, password string) (string, string, int32, error)
}

type LogGetter interface {
	GetLogger(ctx context.Context) *slog.Logger
}

type Handler struct {
	metrics MetricsHandler
	auth    AuthHandler
	user    UserHandler
	logger  LogGetter
}

func NewHandler(metrics MetricsHandler, auth AuthHandler, user UserHandler, logger LogGetter) *Handler {
	return &Handler{
		metrics: metrics,
		auth:    auth,
		user:    user,
		logger:  logger}
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
	logger := h.logger.GetLogger(r.Context())

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("client request",
			slog.String("error", err.Error()),
			slog.String("type", "decoder"))
		http.Error(w, "invalid JSON format", http.StatusBadRequest)
		return
	}

	userID, err := h.user.RegisterUser(r.Context(), req.Name, req.Password)
	if err != nil {
		if errors.Is(err, model.ErrUserInvalidRegisterDetails) {
			logger.Warn("client request",
				slog.String("error", err.Error()),
				slog.String("type", "bad register details"))
			http.Error(w, "invalid register details", http.StatusBadRequest)
			return
		}

		if errors.Is(err, model.ErrUserAlreadyExists) {
			logger.Warn("client request",
				slog.String("error", err.Error()),
				slog.String("type", "user already exist"))
			http.Error(w, "user already exists", http.StatusConflict)
			return
		}

		logger.Error("internal server error",
			slog.String("error", err.Error()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(RegisterResponse{UserID: userID})
}

/* Авторизация пользователя */
func (h *Handler) LoginUser(w http.ResponseWriter, r *http.Request) {
	logger := h.logger.GetLogger(r.Context())

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("client request",
			slog.String("error", err.Error()),
			slog.String("type", "decoder"))
		http.Error(w, "invalid JSON format", http.StatusBadRequest)
		return
	}

	accessToken, refreshToken, refreshTTL, err := h.user.LoginUser(r.Context(), req.Name, req.Password)
	if err != nil {
		if errors.Is(err, model.ErrUserInvalidCredentials) {
			logger.Warn("client request",
				slog.String("error", err.Error()),
				slog.String("type", "bad credintials"))
			http.Error(w, "invalid credentials", http.StatusBadRequest)
			return
		}

		if errors.Is(err, model.ErrUserNotFound) {
			logger.Warn("client request",
				slog.String("error", err.Error()),
				slog.String("type", "user not found"))
			http.Error(w, "user already exists", http.StatusNotFound)
			return
		}

		logger.Error("internal server error",
			slog.String("error", err.Error()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	setCookie(w, refreshToken, refreshTTL)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken})
}

/* Рефреш пары токенов */
func (h *Handler) RefreshTokens(w http.ResponseWriter, r *http.Request) {
	logger := h.logger.GetLogger(r.Context())

	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		logger.Warn("client request",
			slog.String("error", err.Error()),
			slog.String("type", "cookies are missing"))
		http.Error(w, "refresh token required", http.StatusUnauthorized)
		return
	}

	cookieRefreshToken := cookie.Value

	accessToken, refreshToken, refreshTTL, err := h.auth.RefreshTokens(r.Context(), cookieRefreshToken)
	if err != nil {
		if errors.Is(err, model.ErrInvalidToken) {
			logger.Warn("client request",
				slog.String("error", err.Error()),
				slog.String("type", "invalid token"))
			http.Error(w, "token revoked or expired or not found", http.StatusForbidden)
			return
		}

		logger.Error("internal server error",
			slog.String("error", err.Error()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	setCookie(w, refreshToken, refreshTTL)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(RefreshResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken})
}

/* Логаут пользователя */
func (h *Handler) LogoutUser(w http.ResponseWriter, r *http.Request) {
	logger := h.logger.GetLogger(r.Context())

	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		logger.Warn("client request",
			slog.String("error", err.Error()),
			slog.String("type", "cookies are missing"))
		http.Error(w, "refresh token required", http.StatusUnauthorized)
		return
	}

	cookieRefreshToken := cookie.Value

	if err := h.auth.RevokeRefreshToken(r.Context(), cookieRefreshToken); err != nil {
		logger.Warn("client request",
			slog.String("error", err.Error()),
			slog.String("type", "revoke token"))
	}

	clearCookie(w)

	w.WriteHeader(http.StatusOK)
}

/* Установить куки */
func setCookie(w http.ResponseWriter, refreshToken string, refreshTTL int32) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   int(refreshTTL),
	})
}

/* Очистить куки */
func clearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:   "refresh_token",
		MaxAge: -1,
	})
}
