package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	refreshPrefix   = "refresh:"
	refreshTokenTTL = 7 * 24 * time.Hour
)

// Repository - кэш
type Repository struct {
	client *redis.Client
	cfg    *Config
}

// New - конструктор кэша
func New(cfg Config) *Repository {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPWD,
		DB:       int(cfg.RedisDB),
	})
	return &Repository{
		client: rdb,
	}
}

// SaveRefreshToken - сохранить refresh токен
func (r *Repository) SaveRefreshToken(ctx context.Context, userID, refreshToken string) error {
	key := refreshPrefix + userID

	return r.client.Set(ctx, key, refreshToken, refreshTokenTTL).Err()
}

// GetRefreshToken - получить refresh токен по id пользователя
func (r *Repository) GetRefreshToken(ctx context.Context, userID string) (string, error) {
	key := refreshPrefix + userID

	return r.client.Get(ctx, key).Result()
}
