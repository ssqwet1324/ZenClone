package app

import (
	"AuthService/internal/client/UsersClient"
	"AuthService/internal/config"
	"AuthService/internal/handler"
	"AuthService/internal/repository/redis"
	"AuthService/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
	"log"
)

func Run() {
	server := gin.Default()

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}

	cfg, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	restyClient := resty.New()

	repo := redis.New(logger, cfg.RedisConfig)

	client := UsersClient.New(restyClient, logger, cfg.ClientConfig)

	authService := service.New(repo, logger, client, cfg)

	authHandler := handler.New(authService, logger, cfg, client)

	server.POST("/register", authHandler.Register)
	server.POST("/login", authHandler.Login)
	server.POST("/refresh", authHandler.Refresh)

	if err := server.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
