package config

import (
	"fmt"
	"strconv"

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
	LogLevel            string `env:"LOG_LEVEL"`
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

// CreateDsn - создание адреса подключения к бд
func (cfg *Config) CreateDsn() string {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DbUser,
		cfg.DbPassword,
		cfg.DbHost,
		strconv.Itoa(cfg.DbPort),
		cfg.DbName,
	)

	return dsn
}
