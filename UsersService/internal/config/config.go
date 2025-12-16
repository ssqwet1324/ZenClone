package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

// Config - структура кфг
type Config struct {
	JWTSecret           string `env:"JWT_SECRET"`
	DbName              string `env:"DB_NAME"`
	DbUser              string `env:"DB_USER"`
	DbPassword          string `env:"DB_PASSWORD"`
	DbHost              string `env:"DB_HOST"`
	DbPort              int    `env:"DB_PORT"`
	MinioEndpoint       string `env:"MINIO_ENDPOINT"`
	MinioAccessKey      string `env:"MINIO_ACCESS_KEY"`
	MinIoPublicEndpoint string `env:"MINIO_PUBLIC_ENDPOINT"`
	MinioSecretKey      string `env:"MINIO_SECRET_KEY"`
	MinioUseSSl         bool   `env:"MINIO_USE_SSL"`
	BucketName          string `env:"BUCKET_NAME"`
}

// New - конструктор кфг
func New() (*Config, error) {
	var cfg Config

	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	return &cfg, nil
}
