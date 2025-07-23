package UsersClient

type TokenRequest struct {
	ID string `json:"id"`
}

type TokenResponse struct {
	RefreshToken string `json:"refresh_token"`
}

type AuthRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type AuthResponse struct {
	ID string `json:"id"`
}

type RegisterRequest struct {
	ID       string `json:"id"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

type UpdateRefreshTokenRequest struct {
	ID           string `json:"id"`
	RefreshToken string `json:"refresh_token"`
}
