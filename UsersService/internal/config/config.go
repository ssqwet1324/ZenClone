package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"strconv"
)

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

func New() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("config UserService: error loading .env file")
	}

	var conf Config

	conf.JWTSecret = os.Getenv("JWT_SECRET")
	conf.DbName = os.Getenv("DB_NAME")
	conf.DbUser = os.Getenv("DB_USER")
	conf.DbPassword = os.Getenv("DB_PASSWORD")
	conf.DbHost = os.Getenv("DB_HOST")

	conf.DbPort, err = strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		return nil, fmt.Errorf("config UserService: error converting DB_PORT to int: %w", err)
	}

	conf.MinioEndpoint = os.Getenv("MINIO_ENDPOINT")
	conf.MinioAccessKey = os.Getenv("MINIO_ACCESS_KEY")
	conf.MinioSecretKey = os.Getenv("MINIO_SECRET_KEY")
	conf.BucketName = os.Getenv("BUCKET_NAME")
	conf.MinIoPublicEndpoint = os.Getenv("MINIO_PUBLIC_ENDPOINT")

	useSSLStr := os.Getenv("MINIO_USE_SSL")
	conf.MinioUseSSl, err = strconv.ParseBool(useSSLStr)
	if err != nil {
		conf.MinioUseSSl = false
	}

	return &conf, nil
}
