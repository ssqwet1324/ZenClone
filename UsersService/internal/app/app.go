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

	userMiddleware := middleware.JWTAuthMiddleware(cfg)

	repo, err := repository.New(cfg)
	if err != nil {
		logger.Fatal("can't initialize repository", zap.Error(err))
	}

	migration := migrations.New(repo, logger)

	ctx := context.Background()

	err = migration.InitTables(ctx)
	if err != nil {
		logger.Fatal("can't initialize migration", zap.Error(err))
	}

	usersService := usecase.New(repo, cfg, logger)

	userHandler := handler.New(usersService, logger)

	apiV1 := server.Group("/api/v1")
	{
		apiV1.GET("/get-user-profile/:username", userHandler.GetProfile)
		apiV1.GET("/user/:username", userHandler.GetUserIDByUsername)
		apiV1.GET("/user/subs/:username", userHandler.GetSubsUser)

		apiV1.POST("/update-user-info", userMiddleware, userHandler.UpdateProfile)
		apiV1.POST("/user/subscribe/:username", userMiddleware, userHandler.Subscribe)
		apiV1.POST("/user/unsubscribe/:username", userMiddleware, userHandler.UnsubscribeFromUser)
		apiV1.POST("/user/upload-avatar", userMiddleware, userHandler.UploadAvatar)
	}

	internalApi := server.Group("/internal")
	{
		internalApi.POST("/add-user", userHandler.AddUser)
		internalApi.POST("/compare-auth-data", userHandler.CompareAuthPassword)
		internalApi.POST("/get-refresh-token", userHandler.GetRefreshToken)
		internalApi.POST("/update-refresh-token", userHandler.UpdateRefreshToken)
	}

	if err := server.Run(":8081"); err != nil {
		logger.Fatal("failed to run server", zap.Error(err))
	}
}
