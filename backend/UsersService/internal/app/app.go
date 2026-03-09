package app

import (
	"UsersService/internal/config"
	"UsersService/internal/handler"
	"UsersService/internal/middleware"
	"UsersService/internal/repository"
	"UsersService/internal/usecase"
	"context"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Run - запуск сервиса
func Run() {
	server := gin.Default()

	server.Use(middleware.ServerMiddleware())

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

	repo, err := repository.Init(context.Background(), cfg)
	if err != nil {
		logger.Fatal("can't initialize repository", zap.Error(err))
	}
	defer repo.Close()

	usersService := usecase.New(repo, cfg, logger)

	userHandler := handler.New(usersService, logger)

	apiV1 := server.Group("/api/v1")
	{
		apiV1.GET("/get-user-profile/:username", userMiddleware, userHandler.GetProfile)
		apiV1.GET("/user/subs/:username", userMiddleware, userHandler.GetSubsUser)

		apiV1.POST("/update-user-info", userMiddleware, userHandler.UpdateProfile)
		apiV1.POST("/user/subscribe/:username", userMiddleware, userHandler.Subscribe)
		apiV1.POST("/user/unsubscribe/:username", userMiddleware, userHandler.UnsubscribeFromUser)
		apiV1.POST("/user/upload-avatar", userMiddleware, userHandler.UploadAvatar)
		apiV1.GET("/user/search", userMiddleware, userHandler.GlobalSearchUser)
	}

	internalAPI := server.Group("/internal")
	{
		internalAPI.POST("/add-user", userHandler.AddUser)
		internalAPI.POST("/compare-auth-data", userHandler.CompareAuthPassword)
		internalAPI.POST("/get-refresh-token", userHandler.GetRefreshToken)
		internalAPI.POST("/update-refresh-token", userHandler.UpdateRefreshToken)
	}

	if err := server.Run(":8081"); err != nil {
		logger.Fatal("failed to run server", zap.Error(err))
	}
}
