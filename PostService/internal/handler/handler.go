package handler

import (
	"PostService/internal/entity"
	"PostService/internal/usecase"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type PostHandler struct {
	uc  *usecase.PostUseCase
	log *zap.Logger
}

func New(uc *usecase.PostUseCase, log *zap.Logger) *PostHandler {
	return &PostHandler{
		uc:  uc,
		log: log.Named("Handler"),
	}
}

// SuccessResponse - единый формат успешного ответа
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// getUserUUID - достает userID из контекста и парсит в UUID
func getUserUUID(ctx *gin.Context) (uuid.UUID, *entity.ErrorResponse) {
	userIDRaw, exists := ctx.Get("userID")
	if !exists {
		return uuid.Nil, &entity.ErrorResponse{
			Error: entity.ErrorDetail{Code: "UNAUTHORIZED",
				Message: "userID not found in context",
			},
		}
	}

	userIDStr, ok := userIDRaw.(string)
	if !ok {
		return uuid.Nil, &entity.ErrorResponse{
			Error: entity.ErrorDetail{Code: "UNAUTHORIZED",
				Message: "userID has wrong type",
			},
		}
	}

	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, &entity.ErrorResponse{
			Error: entity.ErrorDetail{Code: "INVALID_USER_ID",
				Message: "invalid userID format",
			},
		}
	}

	return userUUID, nil
}

// handlePostError - унифицированная обработка ошибок
func handlePostError(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, entity.ErrEmptyTitle):
		ctx.JSON(http.StatusBadRequest, entity.ErrorResponse{Error: entity.ErrorDetail{Code: "EMPTY_TITLE", Message: "Post title cannot be empty"}})
	case errors.Is(err, entity.ErrEmptyContent):
		ctx.JSON(http.StatusBadRequest, entity.ErrorResponse{Error: entity.ErrorDetail{Code: "EMPTY_CONTENT", Message: "Post content cannot be empty"}})
	case errors.Is(err, entity.ErrPostNotOwned):
		ctx.JSON(http.StatusForbidden, entity.ErrorResponse{Error: entity.ErrorDetail{Code: "POST_NOT_OWNED", Message: "Post does not belong to user"}})
	case errors.Is(err, entity.ErrPostsNotFound):
		ctx.JSON(http.StatusNotFound, entity.ErrorResponse{Error: entity.ErrorDetail{Code: "POSTS_NOT_FOUND", Message: "User posts not found"}})
	case errors.Is(err, entity.ErrInternalError):
		ctx.JSON(http.StatusInternalServerError, entity.ErrorResponse{Error: entity.ErrorDetail{Code: "INTERNAL_ERROR", Message: "Internal server error"}})
	default:
		ctx.JSON(http.StatusInternalServerError, entity.ErrorResponse{Error: entity.ErrorDetail{Code: "INTERNAL_ERROR", Message: "An unexpected error occurred"}})
	}
}

// CreatePost - создание поста
func (h *PostHandler) CreatePost(ctx *gin.Context) {
	userUUID, errResp := getUserUUID(ctx)
	if errResp != nil {
		ctx.JSON(http.StatusUnauthorized, errResp)
		return
	}

	var req entity.CreatePostRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.log.Error("CreatePost: error binding JSON request", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: entity.ErrInternalError.Error(),
			},
		})
		return
	}

	post, err := h.uc.CreatePost(ctx.Request.Context(), userUUID, req)
	if err != nil {
		handlePostError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, SuccessResponse{
		Message: "Post created successfully",
		Data: gin.H{
			"id":         post.ID.String(),
			"title":      post.Title,
			"content":    post.Content,
			"author_id":  post.AuthorID.String(),
			"created_at": post.CreatedAt,
		},
	})
}

// UpdatePost - обновление поста
func (h *PostHandler) UpdatePost(ctx *gin.Context) {
	postUUID, err := uuid.Parse(ctx.Param("postID"))
	if err != nil {
		h.log.Error("UpdatePost: error parsing postID", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, entity.ErrorResponse{
			Error: entity.ErrorDetail{Code: "INVALID_POST_ID",
				Message: entity.ErrInternalError.Error(),
			},
		})
		return
	}

	userUUID, errResp := getUserUUID(ctx)
	if errResp != nil {
		ctx.JSON(http.StatusUnauthorized, errResp)
		return
	}

	var req entity.UpdateUserPostRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.log.Error("UpdatePost: error binding JSON request", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: entity.ErrInternalError.Error(),
			},
		})
		return
	}

	err = h.uc.UpdatePost(ctx.Request.Context(), postUUID, userUUID, req)
	if err != nil {
		handlePostError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, SuccessResponse{
		Message: "Post updated successfully",
		Data:    gin.H{"post_id": postUUID.String()}},
	)
}

// DeletePost - удаление поста
func (h *PostHandler) DeletePost(ctx *gin.Context) {
	postUUID, err := uuid.Parse(ctx.Param("postID"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "INVALID_POST_ID",
				Message: "Invalid post ID format",
			},
		})
		return
	}

	userUUID, errResp := getUserUUID(ctx)
	if errResp != nil {
		ctx.JSON(http.StatusUnauthorized, errResp)
		return
	}

	err = h.uc.DeletePost(ctx.Request.Context(), postUUID, userUUID)
	if err != nil {
		handlePostError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, SuccessResponse{
		Message: "Post deleted successfully",
		Data:    gin.H{"post_id": postUUID.String()},
	})
}

// GetPostsUser - получение всех постов пользователя
func (h *PostHandler) GetPostsUser(ctx *gin.Context) {
	userUUID, err := uuid.Parse(ctx.Param("userID"))
	if err != nil {
		h.log.Error("GetPostsUser: error parsing user UUID", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "INVALID_USER_ID",
				Message: "Invalid user ID format",
			},
		})
		return
	}

	data, err := h.uc.GetPostsUser(ctx.Request.Context(), userUUID)
	if err != nil {
		handlePostError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, SuccessResponse{
		Message: "Posts retrieved successfully",
		Data: gin.H{
			"posts": data.Posts,
			"count": len(data.Posts),
		},
	})
}
