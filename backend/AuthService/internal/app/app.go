package app

import (
	"AuthService/internal/client/usersclient"
	"AuthService/internal/config"
	"AuthService/internal/handler"
	"AuthService/internal/middleware"
	"AuthService/internal/repository/redis"
	"AuthService/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// Run - запуск сервиса
func Run() {
	server := gin.Default()

	server.Use(middleware.ServerMiddleware())
	server.Use(middleware.PrometheusMiddleware())

	cfg, err := config.New()
	if err != nil {
		panic("Error loading config" + err.Error())
	}

	logger, err := middleware.InitLogger(cfg.LogLevel)
	if err != nil {
		panic("ErrorDetail init logger: " + err.Error())
	}
	defer func(logger *zap.Logger) {
		_ = logger.Sync()
	}(logger)

	restyClient := resty.New()

	repo := redis.New(cfg.RedisConfig)

	client := usersclient.New(restyClient, logger, &cfg.ClientConfig)

	uc := usecase.New(repo, logger, client, cfg)

	authHandler := handler.New(uc, logger, cfg, client)

	apiV1 := server.Group("/api/v1/auth")
	{
		apiV1.POST("/register", authHandler.Register)
		apiV1.POST("/login", authHandler.Login)
		apiV1.POST("/refresh", authHandler.Refresh)
	}

	//Metrics
	server.GET("/metrics", gin.WrapH(promhttp.Handler()))

	if err := server.Run(":8080"); err != nil {
		logger.Fatal("ErrorDetail starting usecase", zap.Error(err))
	}
}
