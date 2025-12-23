package handler

import (
	"AuthService/internal/entity"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// handleError - обрабатывает ошибки и возвращает соответствующий HTTP статус и сообщение
func (h *AuthHandler) handleError(ctx *gin.Context, err error) {
	if err == nil {
		return
	}

	var statusCode int
	var message string
	var code string

	// Проверяем, является ли ошибка ErrorResponse от клиента
	var errorResp entity.ErrorResponse
	if errors.As(err, &errorResp) {
		// Если это ErrorResponse от клиента, возвращаем его напрямую
		code = errorResp.ErrorDetail.Code
		message = errorResp.ErrorDetail.Message

		// Определяем статус код в зависимости от типа ошибки
		switch code {
		case "USER_ALREADY_EXISTS":
			statusCode = http.StatusConflict
		case "USER_NOT_FOUND", "INVALID_CREDENTIALS", "INCORRECT_PASSWORD_OR_LOGIN":
			statusCode = http.StatusUnauthorized
		case "INVALID_REQUEST":
			statusCode = http.StatusBadRequest
		default:
			statusCode = http.StatusInternalServerError
		}

		ctx.JSON(statusCode, entity.ErrorResponse{
			ErrorDetail: entity.ErrorDetail{
				Code:    code,
				Message: message,
			},
		})
		return
	}

	switch {
	case errors.Is(err, entity.ErrInvalidAuthHeader):
		statusCode = http.StatusBadRequest
		message = "Invalid authorization header. Expected format: Bearer <token>"
		code = "INVALID_AUTH_HEADER"

	case errors.Is(err, entity.ErrInvalidToken):
		statusCode = http.StatusUnauthorized
		message = "Invalid or expired access token"
		code = "INVALID_TOKEN"

	case errors.Is(err, entity.ErrUnexpectedSigningMethod):
		statusCode = http.StatusUnauthorized
		message = "Token signing method is not supported"
		code = "UNSUPPORTED_TOKEN_METHOD"

	case errors.Is(err, entity.ErrCannotParseClaims):
		statusCode = http.StatusUnauthorized
		message = "Failed to parse token claims"
		code = "INVALID_TOKEN_CLAIMS"

	case errors.Is(err, entity.ErrUserIDNotFound):
		statusCode = http.StatusUnauthorized
		message = "User ID not found in token"
		code = "USER_ID_NOT_FOUND"

	case errors.Is(err, entity.ErrRefreshTokenMismatch):
		statusCode = http.StatusUnauthorized
		message = "Refresh token mismatch. Please login again"
		code = "REFRESH_TOKEN_MISMATCH"

	case errors.Is(err, entity.ErrGetRefreshToken):
		statusCode = http.StatusUnauthorized
		message = "Refresh token not found. Please login again"
		code = "REFRESH_TOKEN_NOT_FOUND"

	case errors.Is(err, entity.ErrCompareAuthData):
		statusCode = http.StatusUnauthorized
		message = "Invalid login or password"
		code = "INVALID_CREDENTIALS"

	case errors.Is(err, entity.ErrHashPassword):
		statusCode = http.StatusInternalServerError
		message = "Failed to process password"
		code = "PASSWORD_HASH_ERROR"

	case errors.Is(err, entity.ErrRegisterUser):
		statusCode = http.StatusInternalServerError
		message = "Failed to register user. Please try again later"
		code = "REGISTRATION_FAILED"

	case errors.Is(err, entity.ErrGenerateAccessToken):
		statusCode = http.StatusInternalServerError
		message = "Failed to generate access token"
		code = "TOKEN_GENERATION_ERROR"

	case errors.Is(err, entity.ErrSignToken):
		statusCode = http.StatusInternalServerError
		message = "Failed to sign token"
		code = "TOKEN_SIGNING_ERROR"

	case errors.Is(err, entity.ErrSaveRefreshToken):
		statusCode = http.StatusInternalServerError
		message = "Failed to save refresh token"
		code = "REFRESH_TOKEN_SAVE_ERROR"

	case errors.Is(err, entity.ErrUpdateRefreshToken):
		statusCode = http.StatusInternalServerError
		message = "Failed to update refresh token"
		code = "REFRESH_TOKEN_UPDATE_ERROR"

	default:
		statusCode = http.StatusInternalServerError
		message = "An unexpected error occurred"
		code = "INTERNAL_ERROR"
	}

	ctx.JSON(statusCode, entity.ErrorResponse{
		ErrorDetail: entity.ErrorDetail{
			Code:    code,
			Message: message,
		},
	})
}
