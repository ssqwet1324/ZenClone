package handler

import (
	"AuthService/internal/client/UsersClient"
	"AuthService/internal/config"
	"AuthService/internal/entity"
	"AuthService/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
)

type AuthHandler struct {
	service *service.Service
	log     *zap.Logger
	client  UsersClient.ClientProvider
	cfg     *config.Config
}

func New(service *service.Service, logger *zap.Logger, cfg *config.Config, client UsersClient.ClientProvider) *AuthHandler {
	return &AuthHandler{
		service: service,
		log:     logger,
		cfg:     cfg,
		client:  client,
	}
}

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
		"user_id":       userID,
		"refresh_token": refreshToken,
		"access_token":  accessToken,
	})
}

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

func (handler *AuthHandler) Login(ctx *gin.Context) {
	var user entity.LoginUserInfo
	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})

		return
	}

	userID, accessToken, refreshToken, err := handler.service.LoginAccount(ctx, user.Login, user.Password)
	if err != nil {
		handler.log.Error("LoginAccount failed", zap.Error(err))
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})

		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"user_id":       userID,
		"refresh_token": refreshToken,
		"access_token":  accessToken,
	})
}
