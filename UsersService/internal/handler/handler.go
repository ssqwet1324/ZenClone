package handler

import (
	"UsersService/internal/entity"
	"UsersService/internal/usecase"
	"errors"
	"mime/multipart"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// максимальный размер аватарки
const maxAvatarSize = 50 * 1024 * 1024

type UsersHandler struct {
	uc  *usecase.UserService
	log *zap.Logger
}

func New(uc *usecase.UserService, log *zap.Logger) *UsersHandler {
	return &UsersHandler{
		uc:  uc,
		log: log.Named("UsersHandler"),
	}
}

func (h *UsersHandler) UpdateRefreshToken(c *gin.Context) {
	var req entity.UpdateRefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warn("UpdateRefreshToken: failed to bind JSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: entity.ErrInvalidRequest.Error(),
			},
		})
		return
	}

	if err := h.uc.UpdateRefreshToken(c.Request.Context(), req); err != nil {
		if errors.Is(err, entity.ErrFailedToUpdateRefreshToken) {
			c.JSON(http.StatusInternalServerError, entity.ErrorResponse{
				Error: entity.ErrorDetail{
					Code:    "FAILED_TO_UPDATE_REFRESH_TOKEN",
					Message: entity.ErrFailedToUpdateRefreshToken.Error(),
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: entity.ErrInternalServer.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "refresh token updated"})
}

func (h *UsersHandler) GetRefreshToken(ctx *gin.Context) {
	var req entity.TokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.log.Warn("GetRefreshToken: failed to bind JSON", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: entity.ErrInvalidRequest.Error(),
			},
		})
		return
	}

	token, err := h.uc.GetRefreshToken(ctx.Request.Context(), req.ID)
	if err != nil {
		if errors.Is(err, entity.ErrFailedToGetRefreshToken) {
			ctx.JSON(http.StatusInternalServerError, entity.ErrorResponse{
				Error: entity.ErrorDetail{
					Code:    "FAILED_TO_GET_REFRESH_TOKEN",
					Message: entity.ErrFailedToGetRefreshToken.Error(),
				},
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: entity.ErrInternalServer.Error(),
			},
		})
		return
	}

	ctx.JSON(http.StatusOK, token)
}

func (h *UsersHandler) CompareAuthPassword(ctx *gin.Context) {
	var req entity.AuthRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.log.Warn("CompareAuthPassword: failed to bind JSON", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: entity.ErrInvalidRequest.Error(),
			},
		})
		return
	}

	data, err := h.uc.CompareAuthData(ctx, req)
	if err != nil {
		if errors.Is(err, entity.ErrUserNotFound) {
			ctx.JSON(http.StatusNotFound, entity.ErrorResponse{
				Error: entity.ErrorDetail{
					Code:    "USER_NOT_FOUND",
					Message: entity.ErrUserNotFound.Error(),
				},
			})
			return
		}

		if errors.Is(err, entity.ErrIncorrectPassword) {
			ctx.JSON(http.StatusUnauthorized, entity.ErrorResponse{
				Error: entity.ErrorDetail{
					Code:    "INCORRECT_PASSWORD",
					Message: entity.ErrIncorrectPassword.Error(),
				},
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: entity.ErrInternalServer.Error(),
			},
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"id": data.ID.String()})
}

// AddUser - создание пользователя
func (h *UsersHandler) AddUser(ctx *gin.Context) {
	var req entity.AddUserRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.log.Warn("AddUser: failed to bind JSON", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: entity.ErrInvalidRequest.Error(),
			},
		})
		return
	}

	if err := h.uc.AddUser(ctx.Request.Context(), req); err != nil {
		if errors.Is(err, entity.ErrUserAlreadyExists) {
			ctx.JSON(http.StatusBadRequest, entity.ErrorResponse{
				Error: entity.ErrorDetail{
					Code:    "USER_ALREADY_EXISTS",
					Message: entity.ErrUserAlreadyExists.Error(),
				},
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: entity.ErrInternalServer.Error(),
			},
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "user added"})
}

func (h *UsersHandler) GetProfile(ctx *gin.Context) {
	username := ctx.Param("username")

	data, err := h.uc.GetUserProfileByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, entity.ErrFailedToGetUserInfo) {
			ctx.JSON(http.StatusInternalServerError, entity.ErrorResponse{
				Error: entity.ErrorDetail{
					Code:    "FAILED_TO_GET_USER_INFO",
					Message: entity.ErrFailedToGetUserInfo.Error(),
				},
			})
			return
		}

		if errors.Is(err, entity.ErrFailedToGetAvatarURL) {
			ctx.JSON(http.StatusInternalServerError, entity.ErrorResponse{
				Error: entity.ErrorDetail{
					Code:    "FAILED_TO_GET_AVATAR_URL",
					Message: entity.ErrFailedToGetAvatarURL.Error(),
				},
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: entity.ErrInternalServer.Error(),
			},
		})
		return
	}

	ctx.JSON(http.StatusOK, data)
}

func (h *UsersHandler) UpdateProfile(ctx *gin.Context) {
	// берем userID из jwt токена
	userIDRaw, exists := ctx.Get("userID")
	if !exists {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "UNAUTHORIZED",
				Message: "userID not found in context",
			},
		})
		return
	}

	userIDStr, ok := userIDRaw.(string)
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "UNAUTHORIZED",
				Message: "userID has wrong type",
			},
		})
		return
	}

	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: "invalid userID format",
			},
		})
		return
	}

	var req entity.UpdateUserProfileInfoRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.log.Warn("UpdateProfile: failed to bind JSON", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: entity.ErrInvalidRequest.Error(),
			},
		})
		return
	}

	data, err := h.uc.UpdateUserProfile(ctx, userUUID, req)
	if err != nil {
		if errors.Is(err, entity.ErrIncorrectPassword) {
			ctx.JSON(http.StatusUnauthorized, entity.ErrorResponse{
				Error: entity.ErrorDetail{
					Code:    "INCORRECT_PASSWORD",
					Message: entity.ErrIncorrectPassword.Error(),
				},
			})
			return
		}

		if errors.Is(err, entity.ErrFailedToGetUserInfo) {
			ctx.JSON(http.StatusInternalServerError, entity.ErrorResponse{
				Error: entity.ErrorDetail{
					Code:    "FAILED_TO_GET_USER_INFO",
					Message: entity.ErrFailedToGetUserInfo.Error(),
				},
			})
			return
		}

		if errors.Is(err, entity.ErrFailedToUpdateProfile) {
			ctx.JSON(http.StatusInternalServerError, entity.ErrorResponse{
				Error: entity.ErrorDetail{
					Code:    "FAILED_TO_UPDATE_PROFILE",
					Message: entity.ErrFailedToUpdateProfile.Error(),
				},
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: entity.ErrInternalServer.Error(),
			},
		})
		return
	}

	ctx.JSON(http.StatusOK, data)
}

func (h *UsersHandler) GetUserIDByUsername(ctx *gin.Context) {
	username := ctx.Param("username")

	data, err := h.uc.GetUserIDByUsername(ctx.Request.Context(), username)
	if err != nil {
		if errors.Is(err, entity.ErrUserNotFound) {
			ctx.JSON(http.StatusNotFound, entity.ErrorResponse{
				Error: entity.ErrorDetail{
					Code:    "USER_NOT_FOUND",
					Message: entity.ErrUserNotFound.Error(),
				},
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: entity.ErrInternalServer.Error(),
			},
		})
		return
	}

	h.log.Info("GetUserIDByUsername: success", zap.String("userID", data.ID.String()))
	ctx.JSON(http.StatusOK, data)
}

func (h *UsersHandler) Subscribe(ctx *gin.Context) {
	userIDRaw, exists := ctx.Get("userID")
	if !exists {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "UNAUTHORIZED",
				Message: "userID not found in context",
			},
		})
		return
	}

	userIDStr, ok := userIDRaw.(string)
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "UNAUTHORIZED",
				Message: "userID has wrong type",
			},
		})
		return
	}

	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.log.Warn("Subscribe: invalid userID format", zap.String("userID", userIDStr), zap.Error(err))
		ctx.AbortWithStatusJSON(http.StatusBadRequest, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: "invalid userID format",
			},
		})
		return
	}

	username := ctx.Param("username")

	err = h.uc.SubscribeToUser(ctx, userUUID, username)
	if err != nil {
		if errors.Is(err, entity.ErrUserNotFound) {
			ctx.JSON(http.StatusNotFound, entity.ErrorResponse{
				Error: entity.ErrorDetail{
					Code:    "USER_NOT_FOUND",
					Message: entity.ErrUserNotFound.Error(),
				},
			})
			return
		}

		if errors.Is(err, entity.ErrFailedToSubscribe) {
			ctx.JSON(http.StatusInternalServerError, entity.ErrorResponse{
				Error: entity.ErrorDetail{
					Code:    "FAILED_TO_SUBSCRIBE",
					Message: entity.ErrFailedToSubscribe.Error(),
				},
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: entity.ErrInternalServer.Error(),
			},
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "user subscribed"})
}

func (h *UsersHandler) GetSubsUser(ctx *gin.Context) {
	username := ctx.Param("username")

	data, err := h.uc.GetSubsUser(ctx, username)
	if err != nil {
		if errors.Is(err, entity.ErrUserNotFound) {
			ctx.JSON(http.StatusNotFound, entity.ErrorResponse{
				Error: entity.ErrorDetail{
					Code:    "USER_NOT_FOUND",
					Message: entity.ErrUserNotFound.Error(),
				},
			})
			return
		}

		if errors.Is(err, entity.ErrFailedToGetUserInfo) {
			ctx.JSON(http.StatusInternalServerError, entity.ErrorResponse{
				Error: entity.ErrorDetail{
					Code:    "FAILED_TO_GET_USER_INFO",
					Message: entity.ErrFailedToGetUserInfo.Error(),
				},
			})
			return
		}

		if errors.Is(err, entity.ErrNoSubscriptions) {
			ctx.JSON(http.StatusNotFound, entity.ErrorResponse{
				Error: entity.ErrorDetail{
					Code:    "NO_SUBSCRIPTIONS",
					Message: entity.ErrNoSubscriptions.Error(),
				},
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: entity.ErrInternalServer.Error(),
			},
		})
		return
	}

	ctx.JSON(http.StatusOK, data)
}

func (h *UsersHandler) UnsubscribeFromUser(ctx *gin.Context) {
	userIDRaw, exists := ctx.Get("userID")
	if !exists {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "UNAUTHORIZED",
				Message: "userID not found in context",
			},
		})
		return
	}

	userIDStr, ok := userIDRaw.(string)
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "UNAUTHORIZED",
				Message: "userID has wrong type",
			},
		})
		return
	}

	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: "invalid userID format",
			},
		})
		return
	}

	username := ctx.Param("username")

	err = h.uc.UnsubscribeFromUser(ctx, userUUID, username)
	if err != nil {
		if errors.Is(err, entity.ErrUserNotFound) {
			ctx.JSON(http.StatusNotFound, entity.ErrorResponse{
				Error: entity.ErrorDetail{
					Code:    "USER_NOT_FOUND",
					Message: entity.ErrUserNotFound.Error(),
				},
			})
			return
		}

		if errors.Is(err, entity.ErrFailedToUnsubscribe) {
			ctx.JSON(http.StatusInternalServerError, entity.ErrorResponse{
				Error: entity.ErrorDetail{
					Code:    "FAILED_TO_UNSUBSCRIBE",
					Message: entity.ErrFailedToUnsubscribe.Error(),
				},
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: entity.ErrInternalServer.Error(),
			},
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "user unsubscribed"})
}

func (h *UsersHandler) UploadAvatar(ctx *gin.Context) {
	userIDRaw, exists := ctx.Get("userID")
	if !exists {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "UNAUTHORIZED",
				Message: "userID not found in context",
			},
		})
		return
	}

	userIDStr, ok := userIDRaw.(string)
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "UNAUTHORIZED",
				Message: "userID has wrong type",
			},
		})
		return
	}

	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.log.Warn("UploadAvatar: invalid userID format", zap.String("userID", userIDStr), zap.Error(err))
		ctx.AbortWithStatusJSON(http.StatusBadRequest, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: "invalid userID format",
			},
		})
		return
	}

	file, err := ctx.FormFile("avatar")
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: "avatar file is required",
			},
		})
		return
	}

	// Открываем файл
	srcFile, err := file.Open()
	if err != nil {
		h.log.Warn("UploadAvatar: error opening uploaded file", zap.Error(err))
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "cannot open uploaded file",
			},
		})
		return
	}
	defer func(srcFile multipart.File) {
		err := srcFile.Close()
		if err != nil {
			h.log.Warn("UploadAvatar: error closing uploaded file", zap.Error(err))
			return
		}
	}(srcFile)

	if file.Size > maxAvatarSize {
		h.log.Warn("UploadAvatar: file too large", zap.Int64("size", file.Size))
		ctx.AbortWithStatusJSON(http.StatusBadRequest, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: "avatar size must be <= 50MB",
			},
		})
		return
	}

	// данные аватарки
	avatarReq := entity.AvatarRequest{
		Name:   file.Filename,
		Size:   file.Size,
		Reader: srcFile,
	}

	err = h.uc.UploadAvatar(ctx, userUUID, avatarReq)
	if err != nil {
		if errors.Is(err, entity.ErrFailedToUploadAvatar) {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, entity.ErrorResponse{
				Error: entity.ErrorDetail{
					Code:    "FAILED_TO_UPLOAD_AVATAR",
					Message: entity.ErrFailedToUploadAvatar.Error(),
				},
			})
			return
		}

		ctx.AbortWithStatusJSON(http.StatusInternalServerError, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: entity.ErrInternalServer.Error(),
			},
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "avatar uploaded successfully"})
}
