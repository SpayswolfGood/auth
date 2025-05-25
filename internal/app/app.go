package app

import (
	"github.com/DrusGalkin/forum-auth-grpc/internal/delivery/gin"
	"github.com/DrusGalkin/forum-auth-grpc/internal/repository"
	"github.com/DrusGalkin/forum-auth-grpc/internal/usecase"
	"github.com/DrusGalkin/forum-auth-grpc/pkg/database"
	"github.com/DrusGalkin/forum-auth-grpc/pkg/logger"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"log"
)

func Run() {
	db, err := database.NewSQLiteConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	userUseCase := usecase.NewUserUseCase(userRepo)

	router := gin.SetupRouter(userUseCase)

	if err := router.Run(":8080"); err != nil {
		logger.Logger.Fatal("Ошибка запуска сервера на порту :8080",
			zap.Error(err),
			zap.String("app", "database"))
	}
	logger.Logger.Info("Микросервис стартует на порту :8080")
}
