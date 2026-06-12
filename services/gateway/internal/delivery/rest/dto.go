package delivery

type RegisterRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	UserID string `json:"userid"`
}

type LoginRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
