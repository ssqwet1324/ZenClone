package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"strconv"
)

type Config struct {
	JWTSecret  string `env:"JWT_SECRET"`
	DbName     string `env:"DB_NAME"`
	DbUser     string `env:"DB_USER"`
	DbPassword string `env:"DB_PASSWORD"`
	DbHost     string `env:"DB_HOST"`
	DbPort     int    `env:"DB_PORT"`
	RedisAddr  string `env:"REDIS_ADDR"`
	RedisPwd   string `env:"REDIS_PWD"`
	RedisDB    int64  `env:"REDIS_DB"`
}

func New() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("config UserService: Error loading .env file")
	}

	var conf Config

	conf.JWTSecret = os.Getenv("JWT_SECRET")
	conf.DbName = os.Getenv("DB_NAME")
	conf.DbUser = os.Getenv("DB_USER")
	conf.DbPassword = os.Getenv("DB_PASSWORD")
	conf.DbHost = os.Getenv("DB_HOST")
	conf.DbPort, err = strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		return nil, fmt.Errorf("config UserService: Error converting DB_PORT to int")
	}

	return &conf, nil
}
