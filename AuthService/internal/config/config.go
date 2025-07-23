package config

import (
	"AuthService/internal/client/UsersClient"
	"AuthService/internal/repository/redis"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	JWTSecret    string `json:"JWT_SECRET"`
	RedisConfig  *redis.Config
	ClientConfig *UsersClient.ConfigUsersServiceClient
}

func New() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	var conf Config
	conf.RedisConfig = &redis.Config{}
	conf.ClientConfig = &UsersClient.ConfigUsersServiceClient{}

	conf.JWTSecret = os.Getenv("JWT_SECRET")
	conf.RedisConfig.RedisAddr = os.Getenv("REDIS_ADDR")
	conf.RedisConfig.RedisDB, err = strconv.ParseInt(os.Getenv("REDIS_DB"), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing REDIS_DB: %w", err)
	}
	conf.RedisConfig.RedisPWD = os.Getenv("REDIS_PWD")
	conf.ClientConfig.BaseURL = os.Getenv("BASE_URL")
	conf.ClientConfig.RetryDelay, err = time.ParseDuration(os.Getenv("BASE_RETRY_DELAY"))
	if err != nil {
		return nil, fmt.Errorf("Config: error parsing BASE_RETRY_DELAY: %w", err)
	}
	conf.ClientConfig.RetryCount, err = strconv.ParseInt(os.Getenv("BASE_RETRY_COUNT"), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("Config: error parsing BASE_RETRY_COUNT: %w", err)
	}

	return &conf, nil
}
