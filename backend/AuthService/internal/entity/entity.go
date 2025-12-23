package entity

import (
	"fmt"
)

//..................User Requests & Responses...............

// LoginUserInfo - данные для входа пользователя
type LoginUserInfo struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// RegisterRequest - данные для регистрации нового пользователя
type RegisterRequest struct {
	Login     string `json:"login"`
	Password  string `json:"password"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Bio       string `json:"bio"`
}

// RegisterResponse - ответ сервиса на регистрацию
type RegisterResponse struct {
	ID           string `json:"id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// LoginResponse - ответ сервиса на авторизацию
type LoginResponse struct {
	ID           string `json:"id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// RefreshResponse - ответ сервиса при обновлении токенов
type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// TokenRequest - запрос на получение/проверку токена
type TokenRequest struct {
	ID           string `json:"id"`
	RefreshToken string `json:"refresh_token"`
}

//..................Errors from Service........................

// ErrorResponse - стандартная структура ошибки от UsersService
type ErrorResponse struct {
	ErrorDetail ErrorDetail `json:"error"`
}

// ErrorDetail - подробности ошибки
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Error реализует интерфейс error для ErrorResponse
func (e ErrorResponse) Error() string {
	return fmt.Sprintf("code: %s, message: %s", e.ErrorDetail.Code, e.ErrorDetail.Message)
}
