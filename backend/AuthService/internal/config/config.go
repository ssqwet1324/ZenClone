package config

import (
	"AuthService/internal/client/usersclient"
	"AuthService/internal/repository/redis"
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

// Config - конфиг
type Config struct {
	JWTSecret    string `env:"JWT_SECRET"`
	RedisConfig  redis.Config
	ClientConfig usersclient.ConfigUsersServiceClient
	LogLevel     string `env:"LOG_LEVEL"`
}

// New - конструктор кфг
func New() (*Config, error) {
	var cfg Config

	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		return nil, fmt.Errorf("can't initialize config: %w", err)
	}

	return &cfg, nil
}
