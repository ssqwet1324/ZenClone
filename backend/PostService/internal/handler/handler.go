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

// getUserUUID - достает userID из контекста и парсит в UUID
func getUserUUID(ctx *gin.Context) (uuid.UUID, *entity.ErrorResponse) {
	userIDRaw, exists := ctx.Get("userID")
	if !exists {
		return uuid.Nil, &entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "UNAUTHORIZED",
				Message: "userID not found in context",
			},
		}
	}

	userIDStr, ok := userIDRaw.(string)
	if !ok {
		return uuid.Nil, &entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "UNAUTHORIZED",
				Message: "userID has wrong type",
			},
		}
	}

	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, &entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: "invalid userID format",
			},
		}
	}

	return userUUID, nil
}

// CreatePost godoc
// @Summary Создание поста
// @Description Создает новый пост для текущего аутентифицированного пользователя
// @Tags posts
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param input body entity.CreatePostRequest true "Данные для создания поста"
// @Success 201 {object} entity.CreatePostSuccessResponse "Пост успешно создан"
// @Failure 400 {object} entity.ErrorResponse "Некорректное тело запроса или пустые поля"
// @Failure 401 {object} entity.ErrorResponse "Пользователь не авторизован"
// @Failure 500 {object} entity.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/posts/create [post]
func (h *PostHandler) CreatePost(ctx *gin.Context) {
	userUUID, errResp := getUserUUID(ctx)
	if errResp != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, errResp)
		return
	}

	var req entity.CreatePostRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.log.Warn("CreatePost: failed to bind JSON", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: "invalid request structure",
			},
		})
		return
	}

	post, err := h.uc.CreatePost(ctx.Request.Context(), userUUID, req)
	if err != nil {
		if errors.Is(err, entity.ErrEmptyTitle) {
			ctx.JSON(http.StatusBadRequest, entity.ErrorResponse{
				Error: entity.ErrorDetail{
					Code:    "EMPTY_TITLE",
					Message: entity.ErrEmptyTitle.Error(),
				},
			})
			return
		}

		if errors.Is(err, entity.ErrEmptyContent) {
			ctx.JSON(http.StatusBadRequest, entity.ErrorResponse{
				Error: entity.ErrorDetail{
					Code:    "EMPTY_CONTENT",
					Message: entity.ErrEmptyContent.Error(),
				},
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: entity.ErrInternalError.Error(),
			},
		})
		return
	}

	ctx.JSON(http.StatusCreated, entity.CreatePostSuccessResponse{
		Message: "Post created successfully",
		Data: entity.CreatePostResponseData{
			ID:        post.ID.String(),
			Title:     post.Title,
			Content:   post.Content,
			AuthorID:  post.AuthorID.String(),
			CreatedAt: post.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
	})
}

// UpdatePost godoc
// @Summary Обновление поста
// @Description Обновляет существующий пост, принадлежащий текущему пользователю
// @Tags posts
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param postID path string true "ID поста"
// @Param input body entity.UpdateUserPostRequest true "Данные для обновления поста"
// @Success 200 {object} entity.UpdatePostSuccessResponse "Пост успешно обновлён"
// @Failure 400 {object} entity.ErrorResponse "Некорректный postID или пустые поля"
// @Failure 401 {object} entity.ErrorResponse "Пользователь не авторизован"
// @Failure 403 {object} entity.ErrorResponse "Пост не принадлежит пользователю"
// @Failure 500 {object} entity.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/posts/update/{postID} [post]
func (h *PostHandler) UpdatePost(ctx *gin.Context) {
	postUUID, err := uuid.Parse(ctx.Param("postID"))
	if err != nil {
		h.log.Warn("UpdatePost: invalid postID format", zap.String("postID", ctx.Param("postID")), zap.Error(err))
		ctx.JSON(http.StatusBadRequest, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: "invalid postID format",
			},
		})
		return
	}

	userUUID, errResp := getUserUUID(ctx)
	if errResp != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, errResp)
		return
	}

	var req entity.UpdateUserPostRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.log.Warn("UpdatePost: failed to bind JSON", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: "invalid request structure",
			},
		})
		return
	}

	err = h.uc.UpdatePost(ctx.Request.Context(), postUUID, userUUID, req)
	if err != nil {
		if errors.Is(err, entity.ErrPostNotOwned) {
			ctx.JSON(http.StatusForbidden, entity.ErrorResponse{
				Error: entity.ErrorDetail{
					Code:    "POST_NOT_OWNED",
					Message: entity.ErrPostNotOwned.Error(),
				},
			})
			return
		}

		if errors.Is(err, entity.ErrEmptyTitle) {
			ctx.JSON(http.StatusBadRequest, entity.ErrorResponse{
				Error: entity.ErrorDetail{
					Code:    "EMPTY_TITLE",
					Message: entity.ErrEmptyTitle.Error(),
				},
			})
			return
		}

		if errors.Is(err, entity.ErrEmptyContent) {
			ctx.JSON(http.StatusBadRequest, entity.ErrorResponse{
				Error: entity.ErrorDetail{
					Code:    "EMPTY_CONTENT",
					Message: entity.ErrEmptyContent.Error(),
				},
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: entity.ErrInternalError.Error(),
			},
		})
		return
	}

	ctx.JSON(http.StatusOK, entity.UpdatePostSuccessResponse{
		Message: "Post updated successfully",
		Data: entity.UpdatePostResponseData{
			PostID: postUUID.String(),
		},
	})
}

// DeletePost godoc
// @Summary Удаление поста
// @Description Удаляет пост, принадлежащий текущему пользователю
// @Tags posts
// @Security BearerAuth
// @Produce json
// @Param postID path string true "ID поста"
// @Success 200 {object} entity.DeletePostSuccessResponse "Пост успешно удалён"
// @Failure 400 {object} entity.ErrorResponse "Некорректный postID"
// @Failure 401 {object} entity.ErrorResponse "Пользователь не авторизован"
// @Failure 403 {object} entity.ErrorResponse "Пост не принадлежит пользователю"
// @Failure 500 {object} entity.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/posts/delete/{postID} [delete]
func (h *PostHandler) DeletePost(ctx *gin.Context) {
	postUUID, err := uuid.Parse(ctx.Param("postID"))
	if err != nil {
		h.log.Warn("DeletePost: invalid postID format", zap.String("postID", ctx.Param("postID")), zap.Error(err))
		ctx.JSON(http.StatusBadRequest, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: "invalid postID format",
			},
		})
		return
	}

	userUUID, errResp := getUserUUID(ctx)
	if errResp != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, errResp)
		return
	}

	err = h.uc.DeletePost(ctx.Request.Context(), postUUID, userUUID)
	if err != nil {
		if errors.Is(err, entity.ErrPostNotOwned) {
			ctx.JSON(http.StatusForbidden, entity.ErrorResponse{
				Error: entity.ErrorDetail{
					Code:    "POST_NOT_OWNED",
					Message: entity.ErrPostNotOwned.Error(),
				},
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: entity.ErrInternalError.Error(),
			},
		})
		return
	}

	ctx.JSON(http.StatusOK, entity.DeletePostSuccessResponse{
		Message: "Post deleted successfully",
		Data: entity.DeletePostResponseData{
			PostID: postUUID.String(),
		},
	})
}

// GetPostsUser godoc
// @Summary Получение постов пользователя
// @Description Возвращает список всех постов указанного пользователя
// @Tags posts
// @Produce json
// @Param userID path string true "ID пользователя"
// @Success 200 {object} entity.GetPostsUserSuccessResponse "Список постов получен"
// @Failure 400 {object} entity.ErrorResponse "Некорректный userID"
// @Failure 404 {object} entity.ErrorResponse "Посты пользователя не найдены"
// @Failure 500 {object} entity.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/posts/by-user/{userID} [get]
func (h *PostHandler) GetPostsUser(ctx *gin.Context) {
	userUUID, err := uuid.Parse(ctx.Param("userID"))
	if err != nil {
		h.log.Warn("GetPostsUser: invalid userID format", zap.String("userID", ctx.Param("userID")), zap.Error(err))
		ctx.JSON(http.StatusBadRequest, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: "invalid userID format",
			},
		})
		return
	}

	data, err := h.uc.GetPostsUser(ctx.Request.Context(), userUUID)
	if err != nil {
		if errors.Is(err, entity.ErrPostsNotFound) {
			ctx.JSON(http.StatusNotFound, entity.ErrorResponse{
				Error: entity.ErrorDetail{
					Code:    "POSTS_NOT_FOUND",
					Message: entity.ErrPostsNotFound.Error(),
				},
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: entity.ErrInternalError.Error(),
			},
		})
		return
	}

	ctx.JSON(http.StatusOK, entity.GetPostsUserSuccessResponse{
		Message: "Posts retrieved successfully",
		Data: entity.GetPostsUserResponseData{
			Posts: data.Posts,
			Count: len(data.Posts),
		},
	})
}
