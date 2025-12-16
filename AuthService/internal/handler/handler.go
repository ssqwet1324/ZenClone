package handler

import (
	"AuthService/internal/client/UsersClient"
	"AuthService/internal/config"
	"AuthService/internal/entity"
	"AuthService/internal/usecase"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuthHandler - ручки
type AuthHandler struct {
	service *usecase.Service
	log     *zap.Logger
	client  UsersClient.ClientProvider
	cfg     *config.Config
}

// New - конструктор ручек
func New(service *usecase.Service, logger *zap.Logger, cfg *config.Config, client UsersClient.ClientProvider) *AuthHandler {
	return &AuthHandler{
		service: service,
		log:     logger,
		cfg:     cfg,
		client:  client,
	}
}

// Register - Ручка регистрации
func (handler *AuthHandler) Register(ctx *gin.Context) {
	var user entity.RegisterRequest

	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	handler.log.Info("Register attempt", zap.String("login", user.Login))

	userID, accessToken, refreshToken, err := handler.service.RegisterUser(ctx, user)
	if err != nil {
		handler.log.Error("Register: failed to register user", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "registration failed"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"id":            userID,
		"refresh_token": refreshToken,
		"access_token":  accessToken,
	})
}

// Refresh - получить новые токены
func (handler *AuthHandler) Refresh(ctx *gin.Context) {
	var req entity.TokenRequest

	if err := ctx.ShouldBindJSON(&req); err != nil || req.RefreshToken == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Refresh token required"})
		return
	}

	authHeader := ctx.GetHeader("Authorization")

	newRefreshToken, newAccessToken, err := handler.service.RefreshTokens(ctx, req.RefreshToken, authHeader)
	if err != nil {
		handler.log.Error("Refresh: failed", zap.Error(err))
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"refresh_token": newRefreshToken,
		"access_token":  newAccessToken,
	})
}

// Login - войти в профиль
func (handler *AuthHandler) Login(ctx *gin.Context) {
	var user entity.LoginUserInfo
	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	userID, accessToken, refreshToken, err := handler.service.LoginAccount(ctx, user.Login, user.Password)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"id":            userID,
		"refresh_token": refreshToken,
		"access_token":  accessToken,
	})
}
