package config

import (
	"PostService/internal/client/subsclient"
	"fmt"
	"strconv"

	"github.com/ilyakaznacheev/cleanenv"
)

// Config - кфг
type Config struct {
	JWTSecret    string `env:"JWT_SECRET"`
	DbName       string `env:"DB_NAME"`
	DbUser       string `env:"DB_USER"`
	DbPassword   string `env:"DB_PASSWORD"`
	DbHost       string `env:"DB_HOST"`
	DbPort       int    `env:"DB_PORT"`
	KafkaAddr    string `env:"KAFKA_ADDR"`
	KafkaTopic   string `env:"KAFKA_TOPIC"`
	ClientConfig subsclient.ConfigUsersServiceClient
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
