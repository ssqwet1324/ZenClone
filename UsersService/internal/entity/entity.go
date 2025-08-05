package entity

import (
	"UsersService/internal/client/PostClient"
	"github.com/google/uuid"
)

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
	ID        uuid.UUID `json:"id"`
	Login     string    `json:"login"`
	Password  string    `json:"password"`
	Username  string    `json:"username"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Bio       string    `json:"bio"`
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
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Bio       string `json:"bio"`
}

type UpdateUserProfileInfoRequest struct {
	Username    *string `json:"username"`
	PasswordOld *string `json:"password_old"`
	PasswordNew *string `json:"password_new"`
	FirstName   *string `json:"first_name"`
	LastName    *string `json:"last_name"`
	Bio         *string `json:"bio"`
}

type UpdateUserProfileInfoResponse struct {
	Username    *string `json:"username"`
	PasswordNew *string `json:"password_new"`
	FirstName   *string `json:"first_name"`
	LastName    *string `json:"last_name"`
	Bio         *string `json:"bio"`
}

type UserPosts struct {
	Posts []PostClient.PostResponse `json:"posts"` // слайс постов
}
