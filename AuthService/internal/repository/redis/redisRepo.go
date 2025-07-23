package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"time"
)

const (
	refreshPrefix   = "refresh:"
	refreshTokenTTL = 7 * 24 * time.Hour
)

type Repo struct {
	client *redis.Client
	log    *zap.Logger
	cfg    *Config
}

func New(log *zap.Logger, cfg *Config) *Repo {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPWD,
		DB:       int(cfg.RedisDB),
	})
	return &Repo{
		client: rdb,
		log:    log.Named("RedisRepository"),
	}
}

// SaveRefreshToken - сохранить refresh токен
func (r *Repo) SaveRefreshToken(ctx context.Context, userID, refreshToken string) error {
	key := refreshPrefix + userID
	r.log.Info("Saving refresh token")

	return r.client.Set(ctx, key, refreshToken, refreshTokenTTL).Err()
}

// GetRefreshToken - получить refresh токен по id пользователя
func (r *Repo) GetRefreshToken(ctx context.Context, userID string) (string, error) {
	key := refreshPrefix + userID
	r.log.Info("Getting refresh token")
	return r.client.Get(ctx, key).Result()
}
