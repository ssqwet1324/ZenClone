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

	apiV1 := server.Group("/api/v1/auth")
	{
		apiV1.POST("/register", authHandler.Register)
		apiV1.POST("/login", authHandler.Login)
		apiV1.POST("/refresh", authHandler.Refresh)
	}

	if err := server.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
