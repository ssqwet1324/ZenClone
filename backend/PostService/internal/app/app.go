package app

import (
	"PostService/internal/client/subsclient"
	"PostService/internal/config"
	"PostService/internal/handler"
	"PostService/internal/kafka"
	"PostService/internal/middleware"
	"PostService/internal/repository"
	"PostService/internal/usecase"
	"context"

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

	userMiddleware := middleware.JWTAuthMiddleware(cfg)

	repo, err := repository.Init(context.Background(), cfg, logger)
	if err != nil {
		logger.Fatal("failed to init postgres repo", zap.Error(err))
	}
	defer repo.Close()

	producer := kafka.New([]string{cfg.KafkaAddr}, cfg.KafkaTopic)

	// Инициализируем клиента для UsersService (подписок)
	restyClient := resty.New()

	client := subsclient.New(restyClient, logger, &cfg.ClientConfig)

	postService := usecase.New(repo, cfg, producer, logger, client)

	postHandler := handler.New(postService, logger)

	apiV1 := server.Group("/api/v1/posts")
	{
		apiV1.POST("/create", userMiddleware, postHandler.CreatePost)
		apiV1.POST("/update/:postID", userMiddleware, postHandler.UpdatePost)
		apiV1.DELETE("/delete/:postID", userMiddleware, postHandler.DeletePost)
		apiV1.GET("/by-user/:userID", postHandler.GetPostsUser)
		apiV1.GET("/feed", userMiddleware, postHandler.GetPostsFeedFromUser)
	}

	//Metrics
	server.GET("/metrics", gin.WrapH(promhttp.Handler()))

	if err := server.Run(":8082"); err != nil {
		logger.Fatal("failed to run http server", zap.Error(err))
	}
}
