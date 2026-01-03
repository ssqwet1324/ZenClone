package handler

import (
	"AuthService/internal/client/usersclient"
	"AuthService/internal/config"
	"AuthService/internal/entity"
	"AuthService/internal/usecase"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuthHandler - обработчик аутентификации
type AuthHandler struct {
	uc     *usecase.UseCase
	log    *zap.Logger
	client usersclient.ClientProvider
	cfg    *config.Config
}

// New - конструктор ручек
func New(service *usecase.UseCase, logger *zap.Logger, cfg *config.Config, client usersclient.ClientProvider) *AuthHandler {
	return &AuthHandler{
		uc:     service,
		log:    logger,
		cfg:    cfg,
		client: client,
	}
}

// Register godoc
// @Summary Регистрация пользователя
// @Description Регистрирует нового пользователя, создаёт учётную запись и возвращает access и refresh JWT токены
// @Tags auth
// @Accept json
// @Produce json
// @Param input body entity.RegisterRequest true "Данные для регистрации пользователя"
// @Success 201 {object} entity.RegisterResponse "Пользователь успешно зарегистрирован"
// @Failure 400 {object} entity.ErrorResponse "Некорректное тело запроса"
// @Failure 409 {object} entity.ErrorResponse "Пользователь с таким логином уже существует"
// @Failure 500 {object} entity.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(ctx *gin.Context) {
	var user entity.RegisterRequest

	if err := ctx.ShouldBindJSON(&user); err != nil {
		h.log.Warn("Register: invalid request body", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, entity.ErrorResponse{
			ErrorDetail: entity.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: entity.ErrInternalServer.Error(),
			},
		})
		return
	}

	userID, accessToken, refreshToken, err := h.uc.RegisterUser(ctx, user)
	if err != nil {
		h.handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, entity.RegisterResponse{
		ID:           userID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Username:     user.Username,
	})
}

// Refresh godoc
// @Summary Обновление JWT токенов
// @Description Обновляет access и refresh токены по валидному refresh токену и access токену из заголовка Authorization
// @Tags auth
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer access token" example(Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...)
// @Param input body entity.TokenRequest true "Refresh токен"
// @Success 200 {object} entity.RefreshResponse "Токены успешно обновлены"
// @Failure 400 {object} entity.ErrorResponse "Отсутствует refresh token или заголовок Authorization"
// @Failure 401 {object} entity.ErrorResponse "Токен невалиден или истёк срок действия"
// @Failure 500 {object} entity.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) Refresh(ctx *gin.Context) {
	var req entity.TokenRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.log.Warn("Refresh: invalid request body", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, entity.ErrorResponse{
			ErrorDetail: entity.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: entity.ErrInternalServer.Error(),
			},
		})
		return
	}

	if req.RefreshToken == "" {
		ctx.JSON(http.StatusBadRequest, entity.ErrorResponse{
			ErrorDetail: entity.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: "Refresh token is required",
			},
		})
		return
	}

	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		ctx.JSON(http.StatusBadRequest, entity.ErrorResponse{
			ErrorDetail: entity.ErrorDetail{
				Code:    "MISSING_AUTH_HEADER",
				Message: "Authorization header is required",
			},
		})
		return
	}

	newRefreshToken, newAccessToken, err := h.uc.RefreshTokens(ctx, req.RefreshToken, authHeader)
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
// @Summary Аутентификация пользователя
// @Description Выполняет вход пользователя по логину и паролю и возвращает access и refresh JWT токены
// @Tags auth
// @Accept json
// @Produce json
// @Param input body entity.LoginUserInfo true "Логин и пароль пользователя"
// @Success 200 {object} entity.LoginResponse "Аутентификация прошла успешно"
// @Failure 400 {object} entity.ErrorResponse "Некорректное тело запроса"
// @Failure 401 {object} entity.ErrorResponse "Неверный логин или пароль"
// @Failure 500 {object} entity.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(ctx *gin.Context) {
	var user entity.LoginUserInfo
	if err := ctx.ShouldBindJSON(&user); err != nil {
		h.log.Warn("Login: invalid request body", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, entity.ErrorResponse{
			ErrorDetail: entity.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: entity.ErrInternalServer.Error(),
			},
		})
		return
	}

	login, err := h.uc.LoginAccount(ctx, user.Login, user.Password)
	if err != nil {
		h.handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, login)
}
