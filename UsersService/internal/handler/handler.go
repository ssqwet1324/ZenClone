package handler

import (
	"UsersService/internal/entity"
	"UsersService/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
		h.log.Warn("UpdateRefreshToken: failed to bind JSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	if err := h.service.UpdateRefreshToken(c.Request.Context(), req); err != nil {
		h.log.Error("UpdateRefreshToken: service error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	h.log.Info("UpdateRefreshToken: success")
	c.JSON(http.StatusOK, gin.H{"message": "refresh token updated"})
}

func (h *UsersHandler) GetRefreshToken(ctx *gin.Context) {
	var req entity.TokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.log.Warn("GetRefreshToken: failed to bind JSON", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	token, err := h.service.GetRefreshToken(ctx.Request.Context(), req.ID)
	if err != nil {
		h.log.Error("GetRefreshToken: service error", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	h.log.Info("GetRefreshToken: success", zap.String("userID", req.ID.String()))
	ctx.JSON(http.StatusOK, token)
}

func (h *UsersHandler) CompareAuthPassword(ctx *gin.Context) {
	var req entity.AuthRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.log.Warn("CompareAuthPassword: failed to bind JSON", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	data, err := h.service.CompareAuthData(ctx, req)
	if err != nil {
		h.log.Error("CompareAuthPassword: service error", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	h.log.Info("CompareAuthPassword: success", zap.String("userID", data.ID.String()))
	ctx.JSON(http.StatusOK, gin.H{"id": data.ID.String()})
}

func (h *UsersHandler) AddUser(ctx *gin.Context) {
	var req entity.AddUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.log.Warn("AddUser: failed to bind JSON", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	if err := h.service.AddUser(ctx.Request.Context(), req); err != nil {
		h.log.Error("AddUser: service error", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	h.log.Info("AddUser: success", zap.String("username", req.Username))
	ctx.JSON(http.StatusOK, gin.H{"message": "user added"})
}

func (h *UsersHandler) GetProfile(ctx *gin.Context) {
	username := ctx.Param("username")
	h.log.Info("GetProfile: start", zap.String("username", username))

	data, err := h.service.GetUserProfileByUsername(ctx.Request.Context(), username)
	if err != nil {
		h.log.Error("GetProfile: service error", zap.String("username", username), zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	h.log.Info("GetProfile: success", zap.String("username", username))
	ctx.JSON(http.StatusOK, data)
}

func (h *UsersHandler) UpdateProfile(ctx *gin.Context) {
	userIDRaw, exists := ctx.Get("userID")
	if !exists {
		h.log.Warn("UpdateProfile: userID not found in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "userID not found in context"})

		return
	}

	userIDStr, ok := userIDRaw.(string)
	if !ok {
		h.log.Warn("UpdateProfile: userID has wrong type")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "userID has wrong type"})

		return
	}

	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.log.Warn("UpdateProfile: invalid UUID", zap.String("userID", userIDStr), zap.Error(err))
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid userID format"})

		return
	}

	var req entity.UpdateUserProfileInfoRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.log.Warn("UpdateProfile: failed to bind JSON", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	data, err := h.service.UpdateUserProfile(ctx, userUUID, req)
	if err != nil {
		h.log.Error("UpdateProfile: service error", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	h.log.Info("UpdateProfile: success", zap.String("userID", userUUID.String()))
	ctx.JSON(http.StatusOK, data)
}

func (h *UsersHandler) GetUserIDByUsername(ctx *gin.Context) {
	username := ctx.Param("username")
	h.log.Info("GetUserIDByUsername: start", zap.String("username", username))

	data, err := h.service.GetUserIDByUsername(ctx.Request.Context(), username)
	if err != nil {
		h.log.Error("GetUserIDByUsername: service error", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	h.log.Info("GetUserIDByUsername: success", zap.String("userID", data.ID.String()))
	ctx.JSON(http.StatusOK, data)
}

func (h *UsersHandler) Subscribe(ctx *gin.Context) {
	userIDRaw, exists := ctx.Get("userID")
	if !exists {
		h.log.Warn("Subscribe: userID not found in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "userID not found in context"})

		return
	}

	userIDStr, ok := userIDRaw.(string)
	if !ok {
		h.log.Warn("Subscribe: userID has wrong type")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "userID has wrong type"})

		return
	}

	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.log.Warn("Subscribe: invalid userID format", zap.String("userID", userIDStr), zap.Error(err))
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid userID format"})

		return
	}

	username := ctx.Param("username")
	h.log.Info("Subscribe attempt", zap.String("follower_id", userUUID.String()), zap.String("target_username", username))

	err = h.service.SubscribeToUser(ctx, userUUID, username)
	if err != nil {
		h.log.Error("Subscribe failed", zap.String("follower_id", userUUID.String()), zap.String("target_username", username), zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	h.log.Info("Subscribe successful", zap.String("follower_id", userUUID.String()), zap.String("target_username", username))
	ctx.JSON(http.StatusOK, gin.H{"message": "user subscribed"})
}

func (h *UsersHandler) GetSubsUser(ctx *gin.Context) {
	username := ctx.Param("username")
	h.log.Info("GetSubsUser: start", zap.String("username", username))

	data, err := h.service.GetSubsUser(ctx, username)
	if err != nil {
		h.log.Error("GetSubsUser: failed to retrieve subscriptions", zap.String("username", username), zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve subscriptions"})

		return
	}

	h.log.Info("GetSubsUser: success", zap.String("username", username), zap.Int("subscriptions_count", len(data.Subs)))
	ctx.JSON(http.StatusOK, data)
}

func (h *UsersHandler) UnsubscribeFromUser(ctx *gin.Context) {
	userIDRaw, exists := ctx.Get("userID")
	if !exists {
		h.log.Warn("UnsubscribeFromUser: userID not found in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "userID not found in context"})

		return
	}

	userIDStr, ok := userIDRaw.(string)
	if !ok {
		h.log.Warn("UnsubscribeFromUser: userID has wrong type")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "userID has wrong type"})

		return
	}

	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.log.Warn("UnsubscribeFromUser: invalid UUID format", zap.String("userID", userIDStr), zap.Error(err))
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid userID format"})

		return
	}

	username := ctx.Param("username")
	h.log.Info("UnsubscribeFromUser: start", zap.String("follower_id", userUUID.String()), zap.String("target_username", username))

	err = h.service.UnsubscribeFromUser(ctx, userUUID, username)
	if err != nil {
		h.log.Error("UnsubscribeFromUser: service error", zap.String("follower_id", userUUID.String()), zap.String("target_username", username), zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	h.log.Info("UnsubscribeFromUser: success", zap.String("follower_id", userUUID.String()), zap.String("target_username", username))
	ctx.JSON(http.StatusOK, gin.H{"message": "user unsubscribed"})
}
