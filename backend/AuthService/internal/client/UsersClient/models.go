package UsersClient

// TokenRequest - запрос для получения токена
type TokenRequest struct {
	ID string `json:"id"`
}

// TokenResponse - получения токена
type TokenResponse struct {
	RefreshToken string `json:"refresh_token"`
}

// AuthRequest - запрос для регистрации
type AuthRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// AuthResponse - ответ сервиса
type AuthResponse struct {
	ID string `json:"id"`
}

// RegisterRequest - информация о пользователе
type RegisterRequest struct {
	ID            string `json:"id"`
	Login         string `json:"login"`
	Password      string `json:"password"`
	Username      string `json:"username"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	Bio           string `json:"bio"`
	UserAvatarUrl string `json:"user_avatar_url"`
}

// UpdateRefreshTokenRequest - обновление токенов
type UpdateRefreshTokenRequest struct {
	ID           string `json:"id"`
	RefreshToken string `json:"refresh_token"`
}
