package handler

import (
	"PostService/internal/entity"
	"PostService/internal/usecase"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PostHandler struct {
	uc *usecase.PostUseCase
}

func New(uc *usecase.PostUseCase) *PostHandler {
	return &PostHandler{uc: uc}
}

func (h *PostHandler) CreatePost(ctx *gin.Context) {
	userIDRaw, exists := ctx.Get("userID")
	if !exists {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "userID not found in context"})
		return
	}

	userIDStr, ok := userIDRaw.(string)
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "userID has wrong type"})
		return
	}

	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid userID format"})
		return
	}

	var req entity.CreatePostRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	postID, err := h.uc.CreatePost(ctx, userUUID, req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"id": postID.ID.String()})
}

func (h *PostHandler) UpdatePost(ctx *gin.Context) {
	postID := ctx.Param("postID")
	postIdUUID, err := uuid.Parse(postID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDRaw, exists := ctx.Get("userID")
	if !exists {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "userID not found in context"})
		return
	}

	userIDStr, ok := userIDRaw.(string)
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "userID has wrong type"})
		return
	}

	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid userID format"})
		return
	}

	var req entity.UpdateUserPostRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.uc.UpdatePost(ctx, postIdUUID, userUUID, req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"post updated": postID})
}

func (h *PostHandler) DeletePost(ctx *gin.Context) {
	postID := ctx.Param("postID")
	postIdUUID, err := uuid.Parse(postID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDRaw, exists := ctx.Get("userID")
	if !exists {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "userID not found in context"})
		return
	}

	userIDStr, ok := userIDRaw.(string)
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "userID has wrong type"})
		return
	}

	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid userID format"})
		return
	}

	err = h.uc.DeletePost(ctx, postIdUUID, userUUID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"post deleted": postID})
}

func (h *PostHandler) GetPostsUser(ctx *gin.Context) {
	userID := ctx.Param("userID")
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	data, err := h.uc.GetPostsUser(ctx, userUUID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"posts": data.Posts})
}
