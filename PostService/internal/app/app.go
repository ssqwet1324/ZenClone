package app

import (
	"PostService/internal/config"
	"PostService/internal/handler"
	"PostService/internal/kafka"
	"PostService/internal/middleware"
	"PostService/internal/migrations"
	"PostService/internal/repository"
	"PostService/internal/usecase"
	"context"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"log"
)

func Run() {
	server := gin.Default()

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}

	cfg, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	userMiddleware := middleware.JWTAuthMiddleware(cfg)

	repo, err := repository.New(cfg, logger)
	if err != nil {
		log.Fatal(err)
	}

	producer := kafka.New([]string{"kafka:9092"}, "posts")

	migration := migrations.New(repo, logger)

	ctx := context.Background()

	err = migration.InitTables(ctx)
	if err != nil {
		log.Fatalf("can't initialize tables: %v", err)
	}

	postService := usecase.New(repo, cfg, producer)

	postHandler := handler.New(postService)

	server.POST("/create-post", postHandler.CreatePost)
	server.POST("/update-post/:postID", userMiddleware, postHandler.UpdatePost)
	server.DELETE("/delete-post/:postID", userMiddleware, postHandler.DeletePost)

	if err := server.Run(":8082"); err != nil {
		log.Fatal(err)
	}
}
