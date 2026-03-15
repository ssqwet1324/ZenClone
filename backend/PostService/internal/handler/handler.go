package handler

import (
	"PostService/internal/entity"
	"PostService/internal/usecase"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// PostHandler - ручки для постов
type PostHandler struct {
	uc  usecase.UseCaseInterface
	log *zap.Logger
}

// New - конструктор
func New(uc usecase.UseCaseInterface, log *zap.Logger) *PostHandler {
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
			ErrorDetail: entity.ErrorDetail{
				Code:    "UNAUTHORIZED",
				Message: "userID not found in context",
			},
		}
	}

	userIDStr, ok := userIDRaw.(string)
	if !ok {
		return uuid.Nil, &entity.ErrorResponse{
			ErrorDetail: entity.ErrorDetail{
				Code:    "UNAUTHORIZED",
				Message: "userID has wrong type",
			},
		}
	}

	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, &entity.ErrorResponse{
			ErrorDetail: entity.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: "invalid userID format",
			},
		}
	}

	return userUUID, nil
}

// getJwtToken - получить jwt
func getJwtToken(ctx *gin.Context) (string, error) {
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header missing")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", errors.New("invalid authorization format")
	}

	return parts[1], nil
}

// parseLimit - парсит limit из query параметра
func parseLimit(ctx *gin.Context, log *zap.Logger, funcName string) (int, *entity.ErrorResponse) {
	limit := ctx.Query("limit")
	constLimit := 20
	if limit != "" {
		parsedLimit, err := strconv.Atoi(limit)
		if err != nil {
			log.Warn(funcName+": invalid limit format", zap.String("limit", limit))
			return 0, &entity.ErrorResponse{
				ErrorDetail: entity.ErrorDetail{
					Code:    "INVALID_REQUEST",
					Message: "invalid limit format",
				},
			}
		}
		constLimit = parsedLimit
		// Ограничиваем максимальный лимит
		if constLimit > 50 {
			constLimit = 50
		}
	}

	return constLimit, nil
}

// parseCursor - парсит cursor из query параметра
func parseCursor(ctx *gin.Context) *entity.PostCursor {
	var cursor *entity.PostCursor
	if cur := ctx.Query("cursor"); cur != "" {
		parts := strings.Split(cur, "|")
		if len(parts) == 2 {
			t, err := time.Parse(time.RFC3339, parts[0])
			if err == nil {
				id, err := uuid.Parse(parts[1])
				if err == nil {
					cursor = &entity.PostCursor{
						CreatedAt: t,
						ID:        id,
					}
				}
			}
		}
	}

	return cursor
}

// formatNextCursor - форматирует nextCursor для ответа
func formatNextCursor(cursor *entity.PostCursor) string {
	if cursor == nil {
		return ""
	}
	return fmt.Sprintf("%s|%s",
		cursor.CreatedAt.Format(time.RFC3339),
		cursor.ID.String(),
	)
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
			ErrorDetail: entity.ErrorDetail{
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
				ErrorDetail: entity.ErrorDetail{
					Code:    "EMPTY_TITLE",
					Message: entity.ErrEmptyTitle.Error(),
				},
			})
			return
		}

		if errors.Is(err, entity.ErrEmptyContent) {
			ctx.JSON(http.StatusBadRequest, entity.ErrorResponse{
				ErrorDetail: entity.ErrorDetail{
					Code:    "EMPTY_CONTENT",
					Message: entity.ErrEmptyContent.Error(),
				},
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, entity.ErrorResponse{
			ErrorDetail: entity.ErrorDetail{
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
			ErrorDetail: entity.ErrorDetail{
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
			ErrorDetail: entity.ErrorDetail{
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
				ErrorDetail: entity.ErrorDetail{
					Code:    "POST_NOT_OWNED",
					Message: entity.ErrPostNotOwned.Error(),
				},
			})
			return
		}

		if errors.Is(err, entity.ErrEmptyTitle) {
			ctx.JSON(http.StatusBadRequest, entity.ErrorResponse{
				ErrorDetail: entity.ErrorDetail{
					Code:    "EMPTY_TITLE",
					Message: entity.ErrEmptyTitle.Error(),
				},
			})
			return
		}

		if errors.Is(err, entity.ErrEmptyContent) {
			ctx.JSON(http.StatusBadRequest, entity.ErrorResponse{
				ErrorDetail: entity.ErrorDetail{
					Code:    "EMPTY_CONTENT",
					Message: entity.ErrEmptyContent.Error(),
				},
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, entity.ErrorResponse{
			ErrorDetail: entity.ErrorDetail{
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
			ErrorDetail: entity.ErrorDetail{
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
				ErrorDetail: entity.ErrorDetail{
					Code:    "POST_NOT_OWNED",
					Message: entity.ErrPostNotOwned.Error(),
				},
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, entity.ErrorResponse{
			ErrorDetail: entity.ErrorDetail{
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
// @Summary Получение постов пользователя с пагинацией
// @Description Возвращает список постов указанного пользователя с поддержкой cursor-based pagination.
// @Tags posts
// @Produce json
// @Param userID path string true "ID пользователя"
// @Param limit query int false "Количество постов за один запрос (по умолчанию 20, максимум 50)"
// @Param cursor query string false "Cursor для следующей страницы, формат: created_at|post_id"
// @Success 200 {object} entity.GetPostsUserSuccessResponse "Список постов успешно получен"
// @Failure 400 {object} entity.ErrorResponse "Некорректный userID или limit"
// @Failure 404 {object} entity.ErrorResponse "Посты пользователя не найдены"
// @Failure 500 {object} entity.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/posts/by-user/{userID} [get]
func (h *PostHandler) GetPostsUser(ctx *gin.Context) {
	userUUID, err := uuid.Parse(ctx.Param("userID"))
	if err != nil {
		h.log.Warn("GetPostsUser: invalid userID format", zap.String("userID", ctx.Param("userID")), zap.Error(err))
		ctx.JSON(http.StatusBadRequest, entity.ErrorResponse{
			ErrorDetail: entity.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: "invalid userID format",
			},
		})
		return
	}

	constLimit, errResp := parseLimit(ctx, h.log, "GetPostsUser")
	if errResp != nil {
		ctx.JSON(http.StatusBadRequest, errResp)
		return
	}

	cursor := parseCursor(ctx)

	data, err := h.uc.GetPostsUser(ctx.Request.Context(), userUUID, constLimit, cursor)
	if err != nil {
		if errors.Is(err, entity.ErrPostsNotFound) {
			ctx.JSON(http.StatusNotFound, entity.ErrorResponse{
				ErrorDetail: entity.ErrorDetail{
					Code:    "POSTS_NOT_FOUND",
					Message: entity.ErrPostsNotFound.Error(),
				},
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, entity.ErrorResponse{
			ErrorDetail: entity.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: entity.ErrInternalError.Error(),
			},
		})
		return
	}

	nextCursor := formatNextCursor(data.NextCursor)

	ctx.JSON(http.StatusOK, entity.GetPostsUserSuccessResponse{
		Message: "Posts retrieved successfully",
		Data: entity.GetPostsUserResponseData{
			Posts:      data.Posts,
			Count:      len(data.Posts),
			NextCursor: nextCursor,
		},
	})
}

// GetPostsFeedFromUser godoc
// @Summary Получение ленты постов подписок
// @Description Возвращает список постов от пользователей, на которых подписан текущий пользователь, с поддержкой cursor-based pagination. Посты отсортированы по времени создания (от новых к старым).
// @Tags posts
// @Security BearerAuth
// @Produce json
// @Param username query string true "Username текущего пользователя"
// @Param limit query int false "Количество постов за один запрос (по умолчанию 20, максимум 50)"
// @Param cursor query string false "Cursor для следующей страницы, формат: created_at|post_id"
// @Success 200 {object} entity.GetPostsUserSuccessResponseFromFeed "Лента постов успешно получена"
// @Failure 400 {object} entity.ErrorResponse "Некорректный username или limit"
// @Failure 401 {object} entity.ErrorResponse "Пользователь не авторизован"
// @Failure 404 {object} entity.ErrorResponse "Посты в ленте не найдены (нет подписок или постов)"
// @Failure 500 {object} entity.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/posts/feed [get]
func (h *PostHandler) GetPostsFeedFromUser(ctx *gin.Context) {
	// id пользователя из JWT
	userID, errResp := getUserUUID(ctx)
	if errResp != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, errResp)
		return
	}

	accessToken, err := getJwtToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, entity.ErrorResponse{
			ErrorDetail: entity.ErrorDetail{
				Code:    "UNAUTHORIZED",
				Message: err.Error(),
			},
		})
		return
	}

	// получаем username
	username := ctx.Query("username")
	if username == "" {
		ctx.JSON(http.StatusBadRequest, entity.ErrorResponse{
			ErrorDetail: entity.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: "username is required",
			},
		})
		return
	}

	constLimit, errResp := parseLimit(ctx, h.log, "GetPostsFeedFromUser")
	if errResp != nil {
		ctx.JSON(http.StatusBadRequest, errResp)
		return
	}

	cursor := parseCursor(ctx)

	data, err := h.uc.GetPostsFromFeed(ctx.Request.Context(), username, userID, accessToken, constLimit, cursor)
	if err != nil {
		if errors.Is(err, entity.ErrPostsNotFound) {
			ctx.JSON(http.StatusNotFound, entity.ErrorResponse{
				ErrorDetail: entity.ErrorDetail{
					Code:    "POSTS_NOT_FOUND",
					Message: "You haven't followed anyone yet",
				},
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, entity.ErrorResponse{
			ErrorDetail: entity.ErrorDetail{
				Code:    "ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	nextCursor := formatNextCursor(data.NextCursor)

	ctx.JSON(http.StatusOK, entity.GetPostsUserSuccessResponseFromFeed{
		Data: entity.GetPostsUserResponseDataFromFeed{
			Posts:      data.Posts,
			Count:      len(data.Posts),
			NextCursor: nextCursor,
		},
	})
}
