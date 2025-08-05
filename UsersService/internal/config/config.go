package config

import (
	"UsersService/internal/client/PostClient"
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"strconv"
	"time"
)

type Config struct {
	JWTSecret    string `env:"JWT_SECRET"`
	DbName       string `env:"DB_NAME"`
	DbUser       string `env:"DB_USER"`
	DbPassword   string `env:"DB_PASSWORD"`
	DbHost       string `env:"DB_HOST"`
	DbPort       int    `env:"DB_PORT"`
	ClientConfig *PostClient.PostServiceClient
}

func New() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("config UserService: Error loading .env file")
	}

	var conf Config

	conf.ClientConfig = &PostClient.PostServiceClient{}

	conf.JWTSecret = os.Getenv("JWT_SECRET")
	conf.DbName = os.Getenv("DB_NAME")
	conf.DbUser = os.Getenv("DB_USER")
	conf.DbPassword = os.Getenv("DB_PASSWORD")
	conf.DbHost = os.Getenv("DB_HOST")
	conf.DbPort, err = strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		return nil, fmt.Errorf("config UserService: Error converting DB_PORT to int")
	}
	conf.ClientConfig.BaseURL = os.Getenv("BASE_URL")
	conf.ClientConfig.RetryCount, err = strconv.ParseInt(os.Getenv("RETRY_COUNT"), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("config UserService: Error converting RETRY_COUNT to int")
	}
	conf.ClientConfig.RetryDelay, err = time.ParseDuration(os.Getenv("RETRY_DELAY"))
	if err != nil {
		return nil, fmt.Errorf("config UserService: Error converting RETRY_DELAY to int")
	}

	return &conf, nil
}
