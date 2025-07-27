package handler

import (
	"UsersService/internal/entity"
	"UsersService/internal/usecase"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
)

type UsersHandler struct {
	service *usecase.UserService
	log     *zap.Logger
}

func New(service *usecase.UserService, log *zap.Logger) *UsersHandler {
	return &UsersHandler{
		service: service,
		log:     log.Named("UsersHandler"),
	}
}

func (h *UsersHandler) UpdateRefreshToken(c *gin.Context) {
	var req entity.UpdateRefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}
	if err := h.service.UpdateRefreshToken(c.Request.Context(), req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "refresh token updated"})
}

func (h *UsersHandler) GetRefreshToken(ctx *gin.Context) {
	var req entity.TokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.service.GetRefreshToken(ctx.Request.Context(), req.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, token)
}

func (h *UsersHandler) CompareAuthPassword(ctx *gin.Context) {
	var req entity.AuthRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	data, err := h.service.CompareAuthData(ctx, req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	ctx.JSON(http.StatusOK, gin.H{"id": data.ID.String()})
}

func (h *UsersHandler) AddUser(ctx *gin.Context) {
	var req entity.AddUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}
	if err := h.service.AddUser(ctx.Request.Context(), req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "user added"})
}

func (h *UsersHandler) GetProfile(ctx *gin.Context) {
	username := ctx.Param("username") // username из URL

	data, err := h.service.GetUserProfileByUsername(ctx.Request.Context(), username)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, data)
}
