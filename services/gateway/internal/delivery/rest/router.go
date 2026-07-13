package delivery

import (
	mw "restaurant/services/gateway/internal/delivery/rest/middleware"

	"github.com/go-chi/chi/v5"
)

func NewRouter(handler *Handler, logger mw.Middleware, rateLimiter mw.Middleware, metrics mw.Middleware, auth mw.Middleware) *chi.Mux {
	r := chi.NewRouter()

	mw.SetupMiddleware(r, logger, rateLimiter, metrics)

	r.Get("/health", handler.HealthCheck) // GET /health - возвращает StatusOK

	r.Post("/register", handler.RegisterUser) // POST /register - зарегистрировать пользователя
	r.Post("/login", handler.LoginUser)       // POST /login - залогинить пользователя
	r.Post("/refresh", handler.RefreshTokens) // POST /refresh - получение новой пары токенов
	r.Post("/logout", handler.LogoutUser)     // POST /logout - логаут пользователя

	return r
}
