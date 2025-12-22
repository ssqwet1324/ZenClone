package entity

import "errors"

// LoginUserInfo - данные для входа
type LoginUserInfo struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// RegisterResponse - ответ на регистрацию
type RegisterResponse struct {
	ID           string `json:"id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// LoginResponse - ответ на логин
type LoginResponse struct {
	ID           string `json:"id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// RefreshResponse - ответ на обновление токенов
type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// TokenRequest - получение токена
type TokenRequest struct {
	ID           string `json:"id"`
	RefreshToken string `json:"refresh_token"`
}

// RegisterRequest - информация для регистрации
type RegisterRequest struct {
	Login     string `json:"login"`
	Password  string `json:"password"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Bio       string `json:"bio"`
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

// Ошибки для usecase
var (
	ErrSaveRefreshToken        = errors.New("save refresh token")
	ErrGetRefreshToken         = errors.New("get refresh token: not found in Redis and UsersService")
	ErrSignToken               = errors.New("failed to sign token")
	ErrUnexpectedSigningMethod = errors.New("unexpected signing method")
	ErrInvalidToken            = errors.New("invalid token")
	ErrCannotParseClaims       = errors.New("cannot parse claims")
	ErrUserIDNotFound          = errors.New("userID not found in token claims")
	ErrHashPassword            = errors.New("failed to hash password")
	ErrRegisterUser            = errors.New("failed to register user")
	ErrGenerateAccessToken     = errors.New("failed to generate access token")
	ErrCompareAuthData         = errors.New("failed to compare auth data")
	ErrUpdateRefreshToken      = errors.New("failed to update refresh token")
	ErrInvalidAuthHeader       = errors.New("invalid authorization header")
	ErrRefreshTokenMismatch    = errors.New("refresh token mismatch")
	ErrInternalServer          = errors.New("internal server error")
)
