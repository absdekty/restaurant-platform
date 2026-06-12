package delivery

import (
	"github.com/go-chi/chi/v5"
)

func NewRouter(handler *Handler) *chi.Mux {
	r := chi.NewRouter()

	setupMiddleware(r, handler)

	r.Get("/health", handler.HealthCheck) // GET /health - возвращает StatusOK
	r.Get("/metrics", handler.GetMetrics) // GET /metrics - получить актуальные метрики

	r.Post("/register", handler.RegisterUser) // POST /register - зарегистрировать пользователя
	r.Post("/login", handler.LoginUser)       // POST /login - залогинить пользователя
	r.Post("/refresh", handler.RefreshTokens) // POST /refresh - получение новой пары токенов
	r.Post("/logout", handler.LogoutUser)     // POST /logout - логаут пользователя

	return r
}
