package entity

import (
	"fmt"
)

//..................User Requests & Responses...............

// LoginUserInfo - данные для входа пользователя
type LoginUserInfo struct {
	Login    string `json:"login" validate:"required,min=3,max=16,alphanum"`
	Password string `json:"password" validate:"required,min=8,passwordregex8,has1letters"`
}

// RegisterRequest - данные для регистрации нового пользователя
type RegisterRequest struct {
	Login     string `json:"login" validate:"required,min=3,max=16,alphanum"`
	Password  string `json:"password" validate:"required,min=8,passwordregex8,has1letters"`
	Username  string `json:"username" validate:"required,min=4,max=16,alphanum,has4enletters"`
	FirstName string `json:"first_name" validate:"required,min=2,max=32,has2letters"`
	LastName  string `json:"last_name" validate:"required,min=2,max=32,has2letters"`
	Bio       string `json:"bio" validate:"omitempty,max=500"`
}

// RegisterResponse - ответ сервиса на регистрацию
type RegisterResponse struct {
	ID           string `json:"id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Username     string `json:"username"`
}

// LoginResponse - ответ сервиса на авторизацию
type LoginResponse struct {
	ID           string `json:"id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Username     string `json:"username"`
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

// ErrorResponseValidation - ответ ошибки при валидации
type ErrorResponseValidation struct {
	ErrorDetail []ErrorDetail `json:"error"`
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
