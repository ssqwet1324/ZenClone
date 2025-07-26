package entity

import "github.com/google/uuid"

type UserInfoAuth struct {
	ID           uuid.UUID `json:"id"`
	Login        string    `json:"login"`
	Password     string    `json:"password"`
	RefreshToken string    `json:"refresh_token"`
}

type AuthRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type CompareDataResponse struct {
	ID uuid.UUID `json:"id"`
}

type TokenRequest struct {
	ID uuid.UUID `json:"id"`
}

type TokenResponse struct {
	RefreshToken string `json:"refresh_token"`
}

type UpdateRefreshTokenRequest struct {
	ID           uuid.UUID `json:"id"`
	RefreshToken string    `json:"refresh_token"`
}

type AddUserRequest struct {
	ID       uuid.UUID `json:"id"`
	Login    string    `json:"login"`
	Password string    `json:"password"`
}

// LoginResponse - - для ручки /compare-auth-data
type LoginResponse struct {
	ID       uuid.UUID `json:"id"`
	Password string    `json:"password"`
}

type RefreshTokenResponse struct {
	RefreshToken string `json:"refresh_token"`
}

// ProfileUserInfoResponse - информация о профиле пользователя
type ProfileUserInfoResponse struct {
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Bio       string `json:"bio"`
}

type UpdateUserProfileInfoRequest struct {
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Bio       string `json:"bio"`
}
