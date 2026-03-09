package entity

import "errors"

var (
	// ErrUserNotFound возвращается когда пользователь не найден.
	ErrUserNotFound = errors.New("user not found")

	// ErrIncorrectPassword возвращается при неверном пароле или логине.
	ErrIncorrectPassword = errors.New("incorrect password or login")

	// ErrUserAlreadyExists возвращается когда пользователь с таким логином уже существует.
	ErrUserAlreadyExists = errors.New("user with this login already exists")

	// ErrFailedToAddUser возвращается при ошибке добавления пользователя.
	ErrFailedToAddUser = errors.New("failed to add user")

	// ErrFailedToUpdateProfile возвращается при ошибке обновления профиля.
	ErrFailedToUpdateProfile = errors.New("failed to update user profile")

	// ErrFailedToGetUserInfo возвращается при ошибке получения информации о пользователе.
	ErrFailedToGetUserInfo = errors.New("failed to get user info")

	// ErrFailedToGetRefreshToken возвращается при ошибке получения refresh токена.
	ErrFailedToGetRefreshToken = errors.New("failed to get refresh token")

	// ErrFailedToUpdateRefreshToken возвращается при ошибке обновления refresh токена.
	ErrFailedToUpdateRefreshToken = errors.New("failed to update refresh token")

	// ErrFailedToSubscribe возвращается при ошибке подписки на пользователя.
	ErrFailedToSubscribe = errors.New("failed to subscribe to user")

	// ErrFailedToUnsubscribe возвращается при ошибке отписки от пользователя.
	ErrFailedToUnsubscribe = errors.New("failed to unsubscribe from user")

	// ErrNoSubscriptions возвращается когда подписки не найдены.
	ErrNoSubscriptions = errors.New("no subscriptions found")

	// ErrFailedToUploadAvatar возвращается при ошибке загрузки аватара.
	ErrFailedToUploadAvatar = errors.New("failed to upload avatar")

	// ErrFailedToGetAvatarURL возвращается при ошибке получения URL аватара.
	ErrFailedToGetAvatarURL = errors.New("failed to get avatar URL")

	// ErrInternalServer возвращается при внутренней ошибке сервера.
	ErrInternalServer = errors.New("internal server error")

	// ErrInvalidRequest возвращается при неверной структуре запроса.
	ErrInvalidRequest = errors.New("invalid request structure")
)
