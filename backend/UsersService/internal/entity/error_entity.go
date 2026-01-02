package entity

import "errors"

var (
	// Ошибки авторизации
	ErrUserNotFound      = errors.New("user not found")
	ErrIncorrectPassword = errors.New("incorrect password or login")

	// Ошибки при работе с пользователем
	ErrUserAlreadyExists     = errors.New("user with this login already exists")
	ErrFailedToAddUser       = errors.New("failed to add user")
	ErrFailedToUpdateProfile = errors.New("failed to update user profile")
	ErrFailedToGetUserInfo   = errors.New("failed to get user info")

	// Ошибки токенов
	ErrFailedToGetRefreshToken    = errors.New("failed to get refresh token")
	ErrFailedToUpdateRefreshToken = errors.New("failed to update refresh token")

	// Ошибки подписок
	ErrFailedToSubscribe   = errors.New("failed to subscribe to user")
	ErrFailedToUnsubscribe = errors.New("failed to unsubscribe from user")
	ErrNoSubscriptions     = errors.New("no subscriptions found")

	// Ошибки аватаров
	ErrFailedToUploadAvatar = errors.New("failed to upload avatar")
	ErrFailedToGetAvatarURL = errors.New("failed to get avatar URL")

	// Общие
	ErrInternalServer = errors.New("internal server error")
	ErrInvalidRequest = errors.New("invalid request structure")
)
