package entity

import (
	"io"

	"github.com/google/uuid"
)

// AuthRequest - данные о регистрации
type AuthRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// CompareDataResponse - отдаем id в сервис
type CompareDataResponse struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
}

// TokenRequest - получение токена по id
type TokenRequest struct {
	ID uuid.UUID `json:"id"`
}

// TokenResponse - берем токен
type TokenResponse struct {
	RefreshToken string `json:"refresh_token"`
}

// UpdateRefreshTokenRequest - обновление токена
type UpdateRefreshTokenRequest struct {
	ID           uuid.UUID `json:"id"`
	RefreshToken string    `json:"refresh_token"`
}

// AddUserRequest - добавление нового пользователя
type AddUserRequest struct {
	ID        uuid.UUID `json:"id"`
	Login     string    `json:"login"`
	Password  string    `json:"password"`
	Username  string    `json:"username"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Bio       string    `json:"bio"`
}

// LoginResponse - для ручки /compare-auth-data
type LoginResponse struct {
	ID       uuid.UUID `json:"id"`
	Password string    `json:"password"`
	Username string    `json:"username"`
}

// RefreshTokenResponse - запрос для получения токена
type RefreshTokenResponse struct {
	RefreshToken string `json:"refresh_token"`
}

// ProfileUserInfoResponse - информация о профиле пользователя
type ProfileUserInfoResponse struct {
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	Bio           string `json:"bio"`
	UserAvatarUrl string `json:"user_avatar_url"`
	IsSubscribed  bool   `json:"is_subscribed"`
}

// UpdateUserProfileInfoRequest - обновление данных в профиле
type UpdateUserProfileInfoRequest struct {
	Username    *string `json:"username"`
	PasswordOld *string `json:"password_old"`
	PasswordNew *string `json:"password_new"`
	FirstName   *string `json:"first_name"`
	LastName    *string `json:"last_name"`
	Bio         *string `json:"bio"`
}

// UpdateUserProfileInfoResponse - ответ на обновление профиля
type UpdateUserProfileInfoResponse struct {
	Username    *string `json:"username"`
	PasswordNew *string `json:"password_new"`
	FirstName   *string `json:"first_name"`
	LastName    *string `json:"last_name"`
	Bio         *string `json:"bio"`
}

// SubUserInfo - информация о подписчике
type SubUserInfo struct {
	ID            uuid.UUID `json:"id"`
	Username      string    `json:"username"`
	FirstName     string    `json:"first_name"`
	LastName      string    `json:"last_name"`
	UserAvatarUrl string    `json:"user_avatar_url"`
}

// SubsList - подписчики
type SubsList struct {
	Subs []SubUserInfo `json:"subs"`
}

// AvatarRequest -
type AvatarRequest struct {
	Name   string `json:"name"`
	Size   int64  `json:"size"`
	Reader io.Reader
}

// ErrorResponse - ответ ошибки
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail - информация об ошибке
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// UpdateRefreshTokenResponse - ответ на обновление refresh токена
type UpdateRefreshTokenResponse struct {
	Message string `json:"message"`
}

// CompareAuthPasswordResponse - ответ на сравнение данных авторизации
type CompareAuthPasswordResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

// AddUserResponse - ответ на создание пользователя
type AddUserResponse struct {
	Message string `json:"message"`
}

// SubscribeResponse - ответ на подписку
type SubscribeResponse struct {
	Message string `json:"message"`
}

// UnsubscribeResponse - ответ на отписку
type UnsubscribeResponse struct {
	Message string `json:"message"`
}

// UploadAvatarResponse - ответ на загрузку аватара
type UploadAvatarResponse struct {
	Message string `json:"message"`
}

// PersonDate - информация о пользвателях которых ищем
type PersonDate struct {
	Name          string `json:"name"`
	LastName      string `json:"last_name"`
	Username      string `json:"username"`
	UserAvatarUrl string `json:"user_avatar_url"`
}

// PersonDateList - список пользователей с похожими именем фамилией
type PersonDateList struct {
	Persons []PersonDate `json:"persons"`
}
