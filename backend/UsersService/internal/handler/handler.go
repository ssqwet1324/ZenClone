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

// Максимальный размер аватарки
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

// getUserIDFromJWTToken - получить из jwt токена id пользователя
func getUserIDFromJWTToken(ctx *gin.Context) (uuid.UUID, error) {
	userIDRaw, exists := ctx.Get("userID")
	if !exists {
		return uuid.Nil, errors.New("userID not found in context")
	}

	userIDStr, ok := userIDRaw.(string)
	if !ok {
		return uuid.Nil, errors.New("userID has wrong type")
	}

	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, errors.New("invalid userID format")
	}

	return userUUID, nil
}

// UpdateRefreshToken godoc
// @Summary Обновление refresh токена
// @Description Обновляет refresh токен для пользователя
// @Tags internal
// @Accept json
// @Produce json
// @Param input body entity.UpdateRefreshTokenRequest true "Данные для обновления токена"
// @Success 200 {object} entity.UpdateRefreshTokenResponse "Токен успешно обновлен"
// @Failure 400 {object} entity.ErrorResponse "Некорректный запрос"
// @Failure 500 {object} entity.ErrorResponse "Внутренняя ошибка сервера"
// @Router /internal/update-refresh-token [post]
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

	c.JSON(http.StatusOK, entity.UpdateRefreshTokenResponse{
		Message: "refresh token updated",
	})
}

// GetRefreshToken godoc
// @Summary Получение refresh токена
// @Description Получает refresh токен по ID пользователя
// @Tags internal
// @Accept json
// @Produce json
// @Param input body entity.TokenRequest true "ID пользователя"
// @Success 200 {object} entity.TokenResponse "Токен успешно получен"
// @Failure 400 {object} entity.ErrorResponse "Некорректный запрос"
// @Failure 500 {object} entity.ErrorResponse "Внутренняя ошибка сервера"
// @Router /internal/get-refresh-token [post]
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

// CompareAuthPassword godoc
// @Summary Сравнение данных авторизации
// @Description Проверяет логин и пароль пользователя
// @Tags internal
// @Accept json
// @Produce json
// @Param input body entity.AuthRequest true "Данные для авторизации"
// @Success 200 {object} entity.CompareAuthPasswordResponse "Успешная проверка"
// @Failure 400 {object} entity.ErrorResponse "Некорректный запрос"
// @Failure 401 {object} entity.ErrorResponse "Неверный пароль"
// @Failure 404 {object} entity.ErrorResponse "Пользователь не найден"
// @Failure 500 {object} entity.ErrorResponse "Внутренняя ошибка сервера"
// @Router /internal/compare-auth-data [post]
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
					Code:    "INCORRECT_PASSWORD_OR_LOGIN",
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

	ctx.JSON(http.StatusOK, entity.CompareAuthPasswordResponse{
		ID:       data.ID.String(),
		Username: data.Username,
	})
}

// AddUser godoc
// @Summary Создание пользователя
// @Description Создаёт нового пользователя в системе
// @Tags internal
// @Accept json
// @Produce json
// @Param input body entity.AddUserRequest true "Данные пользователя"
// @Success 200 {object} entity.AddUserResponse "Пользователь успешно создан"
// @Failure 400 {object} entity.ErrorResponse "Некорректный запрос или пользователь уже существует"
// @Failure 500 {object} entity.ErrorResponse "Внутренняя ошибка сервера"
// @Router /internal/add-user [post]
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
			ctx.JSON(http.StatusConflict, entity.ErrorResponse{
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

	ctx.JSON(http.StatusOK, entity.AddUserResponse{
		Message: "user added",
	})
}

// GetProfile godoc
// @Summary Получение профиля пользователя
// @Description Получает информацию о профиле пользователя по username
// @Tags users
// @Security BearerAuth
// @Produce json
// @Param username path string true "Username пользователя"
// @Success 200 {object} entity.ProfileUserInfoResponse "Профиль пользователя"
// @Failure 401 {object} entity.ErrorResponse "Неавторизован"
// @Failure 404 {object} entity.ErrorResponse "Пользователь не найден"
// @Failure 500 {object} entity.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/get-user-profile/{username} [get]
func (h *UsersHandler) GetProfile(ctx *gin.Context) {
	yourUserJWTID, err := getUserIDFromJWTToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "UNAUTHORIZED",
				Message: err.Error(),
			},
		})
		return
	}

	username := ctx.Param("username")

	userID, err := h.uc.GetUserIDByUsername(ctx, username)
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

	otherValidUserID, err := uuid.Parse(userID)
	if err != nil {
		h.log.Warn("GetProfile: invalid userID format", zap.String("userID", userID), zap.Error(err))
		ctx.JSON(http.StatusBadRequest, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "USER_NOT_FOUND",
				Message: "invalid userID format",
			},
		})
		return
	}

	data, err := h.uc.GetUserProfileByID(ctx, yourUserJWTID, otherValidUserID)
	if err != nil {
		if errors.Is(err, entity.ErrFailedToGetUserInfo) {
			ctx.JSON(http.StatusNotFound, entity.ErrorResponse{
				Error: entity.ErrorDetail{
					Code:    "USER_NOT_FOUND",
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

// UpdateProfile godoc
// @Summary Обновление профиля пользователя
// @Description Обновляет информацию профиля текущего пользователя
// @Tags users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param input body entity.UpdateUserProfileInfoRequest true "Данные для обновления профиля"
// @Success 200 {object} entity.UpdateUserProfileInfoResponse "Профиль успешно обновлён"
// @Failure 400 {object} entity.ErrorResponse "Некорректный запрос"
// @Failure 401 {object} entity.ErrorResponse "Неверный пароль или неавторизован"
// @Failure 500 {object} entity.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/update-user-info [post]
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

// Subscribe godoc
// @Summary Подписка на пользователя
// @Description Подписывает текущего пользователя на другого пользователя
// @Tags users
// @Security BearerAuth
// @Produce json
// @Param username path string true "Username пользователя"
// @Success 200 {object} entity.SubscribeResponse "Успешная подписка"
// @Failure 401 {object} entity.ErrorResponse "Неавторизован"
// @Failure 404 {object} entity.ErrorResponse "Пользователь не найден"
// @Failure 500 {object} entity.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/user/subscribe/{username} [post]
func (h *UsersHandler) Subscribe(ctx *gin.Context) {
	followerID, err := getUserIDFromJWTToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "UNAUTHORIZED",
				Message: err.Error(),
			},
		})
		return
	}

	username := ctx.Param("username")
	userID, err := h.uc.GetUserIDByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, entity.ErrUserNotFound) {
			ctx.JSON(http.StatusUnauthorized, entity.ErrorResponse{
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

	followingID, err := uuid.Parse(userID)
	if err != nil {
		h.log.Warn("Subscribe: invalid userID format", zap.String("userID", followingID.String()), zap.Error(err))
		ctx.JSON(http.StatusBadRequest, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: "invalid userID format",
			},
		})
		return
	}

	err = h.uc.SubscribeToUser(ctx, followerID, followingID)
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

	ctx.JSON(http.StatusOK, entity.SubscribeResponse{
		Message: "user subscribed",
	})
}

// GetSubsUser godoc
// @Summary Получение списка подписок пользователя
// @Description Получает список пользователей, на которых подписан пользователь
// @Tags users
// @Security BearerAuth
// @Produce json
// @Param username path string true "Username пользователя"
// @Success 200 {object} entity.SubsList "Список подписок"
// @Failure 401 {object} entity.ErrorResponse "Неавторизован"
// @Failure 404 {object} entity.ErrorResponse "Подписки не найдены"
// @Failure 500 {object} entity.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/user/subs/{username} [get]
func (h *UsersHandler) GetSubsUser(ctx *gin.Context) {
	if _, err := getUserIDFromJWTToken(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "UNAUTHORIZED",
				Message: err.Error(),
			},
		})
		return
	}

	username := ctx.Param("username")

	userID, err := h.uc.GetUserIDByUsername(ctx, username)
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

	validUserID, err := uuid.Parse(userID)
	if err != nil {
		h.log.Warn("GetSubsUser: invalid userID format", zap.String("userID", userID), zap.Error(err))
		ctx.JSON(http.StatusBadRequest, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: "invalid userID format",
			},
		})
		return
	}

	data, err := h.uc.GetSubsUser(ctx, validUserID)
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

// UnsubscribeFromUser godoc
// @Summary Отписка от пользователя
// @Description Отписывает текущего пользователя от другого пользователя
// @Tags users
// @Security BearerAuth
// @Produce json
// @Param username path string true "Username пользователя"
// @Success 200 {object} entity.UnsubscribeResponse "Успешная отписка"
// @Failure 401 {object} entity.ErrorResponse "Неавторизован"
// @Failure 404 {object} entity.ErrorResponse "Пользователь не найден"
// @Failure 500 {object} entity.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/user/unsubscribe/{username} [post]
func (h *UsersHandler) UnsubscribeFromUser(ctx *gin.Context) {
	followerID, err := getUserIDFromJWTToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "UNAUTHORIZED",
				Message: err.Error(),
			},
		})
		return
	}

	username := ctx.Param("username")
	userID, err := h.uc.GetUserIDByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, entity.ErrUserNotFound) {
			ctx.JSON(http.StatusNotFound, entity.ErrorResponse{
				Error: entity.ErrorDetail{
					Code:    "USER_NOT_FOUND",
					Message: entity.ErrUserNotFound.Error(),
				},
			})
		}

		ctx.JSON(http.StatusInternalServerError, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: entity.ErrInternalServer.Error(),
			},
		})
		return
	}

	followingID, err := uuid.Parse(userID)
	if err != nil {
		h.log.Warn("UnsubscribeFromUser: invalid userID format", zap.String("userID", followingID.String()), zap.Error(err))
		ctx.JSON(http.StatusBadRequest, entity.ErrorResponse{
			Error: entity.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: "invalid userID format",
			},
		})
		return
	}

	err = h.uc.UnsubscribeFromUser(ctx, followerID, followingID)
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

	ctx.JSON(http.StatusOK, entity.UnsubscribeResponse{
		Message: "user unsubscribed",
	})
}

// UploadAvatar godoc
// @Summary Загрузка аватара пользователя
// @Description Загружает аватар для текущего пользователя
// @Tags users
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param avatar formData file true "Файл аватара (до 50MB)"
// @Success 200 {object} entity.UploadAvatarResponse "Аватар успешно загружен"
// @Failure 400 {object} entity.ErrorResponse "Некорректный файл"
// @Failure 401 {object} entity.ErrorResponse "Неавторизован"
// @Failure 500 {object} entity.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/user/upload-avatar [post]
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

	ctx.JSON(http.StatusOK, entity.UploadAvatarResponse{
		Message: "avatar uploaded successfully",
	})
}
