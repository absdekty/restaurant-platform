package delivery

import (
	"context"
	"net/http"
	"restaurant/services/gateway/internal/model"
	"strings"
)

type contextKey string

const UserIDKey contextKey = "userID"

type AuthService interface {
	ValidateToken(ctx context.Context, token string) (string, error)
}

type Auth struct {
	authService AuthService
}

func NewAuth(authService AuthService) *Auth {
	return &Auth{authService: authService}
}

func (a *Auth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, "Invalid authorization format. Use Bearer <token>", http.StatusUnauthorized)
			return
		}

		userID, err := a.authService.ValidateToken(r.Context(), parts[1])
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserID(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok || userID == "" {
		return "", model.ErrUnauthorized
	}
	return userID, nil
}
