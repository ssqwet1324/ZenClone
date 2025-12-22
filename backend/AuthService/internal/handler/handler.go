package handler

import (
	"AuthService/internal/client/UsersClient"
	"AuthService/internal/config"
	"AuthService/internal/entity"
	"AuthService/internal/usecase"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuthHandler - ручки
type AuthHandler struct {
	service *usecase.UseCase
	log     *zap.Logger
	client  UsersClient.ClientProvider
	cfg     *config.Config
}

// New - конструктор ручек
func New(service *usecase.UseCase, logger *zap.Logger, cfg *config.Config, client UsersClient.ClientProvider) *AuthHandler {
	return &AuthHandler{
		service: service,
		log:     logger,
		cfg:     cfg,
		client:  client,
	}
}

// handleError - обрабатывает ошибки и возвращает соответствующий HTTP статус и сообщение
func (h *AuthHandler) handleError(ctx *gin.Context, err error) {
	if err == nil {
		return
	}

	var statusCode int
	var message string
	var code string

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
		Error: entity.ErrorDetail{
			Code:    code,
			Message: message,
		},
	})
}

// Register godoc
// @Summary Регистрация пользователя
// @Description Создаёт нового пользователя и возвращает access и refresh токены
// @Tags auth
// @Accept json
// @Produce json
// @Param input body entity.RegisterRequest true "Данные для регистрации"
// @Success 201 {object} entity.RegisterResponse "Пользователь успешно зарегистрирован"
// @Failure 400 {object} entity.ErrorResponse "Некорректный запрос"
// @Failure 500 {object} entity.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(ctx *gin.Context) {
	var user entity.RegisterRequest

	if err := ctx.ShouldBindJSON(&user); err != nil {
		h.log.Warn("Register: invalid request body", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: entity.ErrInternalServer.Error(),
			},
		})
		return
	}

	userID, accessToken, refreshToken, err := h.service.RegisterUser(ctx, user)
	if err != nil {
		h.handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, entity.RegisterResponse{
		ID:           userID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

// Refresh godoc
// @Summary Обновление access и refresh токенов
// @Description Обновляет access и refresh токены по текущему access токену и refresh токену
// @Tags auth
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer токен (текущий access token)" default(Bearer )
// @Param input body entity.TokenRequest true "Данные для обновления токенов"
// @Success 200 {object} entity.RefreshResponse "Токены успешно обновлены"
// @Failure 400 {object} entity.ErrorResponse "Некорректный запрос или отсутствует заголовок Authorization"
// @Failure 401 {object} entity.ErrorResponse "Невалидный/просроченный токен или refresh token не найден"
// @Failure 500 {object} entity.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) Refresh(ctx *gin.Context) {
	var req entity.TokenRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.log.Warn("Refresh: invalid request body", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: entity.ErrInternalServer.Error(),
			},
		})
		return
	}

	if req.RefreshToken == "" {
		ctx.JSON(http.StatusBadRequest, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: "Refresh token is required",
			},
		})
		return
	}

	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		ctx.JSON(http.StatusBadRequest, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "MISSING_AUTH_HEADER",
				Message: "Authorization header is required",
			},
		})
		return
	}

	newRefreshToken, newAccessToken, err := h.service.RefreshTokens(ctx, req.RefreshToken, authHeader)
	if err != nil {
		h.handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, entity.RefreshResponse{
		RefreshToken: newRefreshToken,
		AccessToken:  newAccessToken,
	})
}

// Login godoc
// @Summary Вход в систему
// @Description Аутентификация пользователя по логину и паролю, возвращает access и refresh токены
// @Tags auth
// @Accept json
// @Produce json
// @Param input body entity.LoginUserInfo true "Данные для входа"
// @Success 200 {object} entity.LoginResponse "Успешный вход, выданы токены"
// @Failure 400 {object} entity.ErrorResponse "Некорректный запрос"
// @Failure 401 {object} entity.ErrorResponse "Неверный логин или пароль"
// @Failure 500 {object} entity.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(ctx *gin.Context) {
	var user entity.LoginUserInfo
	if err := ctx.ShouldBindJSON(&user); err != nil {
		h.log.Warn("Login: invalid request body", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: entity.ErrInternalServer.Error(),
			},
		})
		return
	}

	userID, accessToken, refreshToken, err := h.service.LoginAccount(ctx, user.Login, user.Password)
	if err != nil {
		h.handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, entity.LoginResponse{
		ID:           userID,
		RefreshToken: refreshToken,
		AccessToken:  accessToken,
	})
}
