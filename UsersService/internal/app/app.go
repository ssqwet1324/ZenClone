package app

import (
	"UsersService/internal/config"
	"UsersService/internal/handler"
	"UsersService/internal/middleware"
	"UsersService/internal/migrations"
	"UsersService/internal/repository"
	"UsersService/internal/usecase"
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

	repo, err := repository.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	migration := migrations.New(repo, logger)

	ctx := context.Background()

	err = migration.InitTables(ctx)
	if err != nil {
		log.Fatalf("can't initialize tables: %v", err)
	}

	usersService := usecase.New(logger, repo, cfg)

	userHandler := handler.New(usersService, logger)

	server.POST("/add-user", userHandler.AddUser)
	server.POST("/compare-auth-data", userHandler.CompareAuthPassword)
	server.POST("/get-refresh-token", userHandler.GetRefreshToken)
	server.POST("/update-refresh-token", userHandler.UpdateRefreshToken)

	server.GET("/get-user-profile/:username", userHandler.GetProfile)
	server.POST("/update-user-info", userMiddleware, userHandler.UpdateProfile)

	server.GET("/users/api/user/:username", userHandler.GetUserIDByUsername)

	if err := server.Run(":8081"); err != nil {
		log.Fatal(err)
	}
}
