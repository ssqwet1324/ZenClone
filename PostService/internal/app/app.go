package app

import (
	"PostService/internal/config"
	"PostService/internal/handler"
	"PostService/internal/kafka"
	"PostService/internal/middleware"
	"PostService/internal/repository"
	"PostService/internal/usecase"
	"context"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Run() {
	server := gin.Default()

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic("Error creating zap logger: " + err.Error())
	}

	cfg, err := config.New()
	if err != nil {
		logger.Fatal("can't initialize config", zap.Error(err))
	}
	fmt.Println(cfg.CreateDsn())

	userMiddleware := middleware.JWTAuthMiddleware(cfg)

	repo, err := repository.Init(context.Background(), cfg, logger)
	if err != nil {
		logger.Fatal("failed to init postgres repo", zap.Error(err))
	}

	producer := kafka.New([]string{cfg.KafkaAddr}, cfg.KafkaTopic)

	postService := usecase.New(repo, cfg, producer)

	postHandler := handler.New(postService)

	apiV1 := server.Group("/api/v1/posts")
	{
		apiV1.POST("/create", userMiddleware, postHandler.CreatePost)
		apiV1.POST("/update/:postID", userMiddleware, postHandler.UpdatePost)
		apiV1.DELETE("/delete/:postID", userMiddleware, postHandler.DeletePost)
		apiV1.GET("/by-user/:userID", postHandler.GetPostsUser)
	}

	if err := server.Run(":8082"); err != nil {
		log.Fatal(err)
	}
}
