package subsclient

import "github.com/google/uuid"

// SubUserInfo - информация о подписчике
type SubUserInfo struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
}

// SubsList - подписчики
type SubsList struct {
	Subs []SubUserInfo `json:"subs"`
}

// UserProfile - профиль пользователя
type UserProfile struct {
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	Bio           string `json:"bio"`
	UserAvatarURL string `json:"user_avatar_url"`
	IsSubscribed  bool   `json:"is_subscribed"`
}
