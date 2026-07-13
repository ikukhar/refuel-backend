package main

import (
	"fmt"
	"os"

	"github.com/ikukhar/refuel-backend/internal/config"
	"github.com/ikukhar/refuel-backend/internal/handler"
	adminHandler "github.com/ikukhar/refuel-backend/internal/handler/admin"
	"github.com/ikukhar/refuel-backend/internal/model"
	"github.com/ikukhar/refuel-backend/internal/repository"
	"github.com/ikukhar/refuel-backend/internal/router"
	"github.com/ikukhar/refuel-backend/internal/service"
	"github.com/ikukhar/refuel-backend/pkg/jwt"
	"github.com/rs/zerolog"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

func main() {
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()

	cfg, err := config.Load(".env")
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to load config")
	}

	logLevel, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		logLevel = zerolog.DebugLevel
	}
	zerolog.SetGlobalLevel(logLevel)

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Warn),
	})
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to connect to database")
	}

	if err := db.AutoMigrate(&model.User{}, &model.Activity{}, &model.DailyNutrition{}, &model.Recipe{}, &model.UserMealPeriod{}); err != nil {
		logger.Fatal().Err(err).Msg("failed to run migrations")
	}

	recipeRepo := repository.NewRecipeRepository(db)
	if err := recipeRepo.SeedRecipes(); err != nil {
		logger.Fatal().Err(err).Msg("failed to seed recipes")
	}

	jwtManager := jwt.NewManager(cfg.JWTSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL)

	userRepo := repository.NewUserRepository(db)
	activityRepo := repository.NewActivityRepository(db)
	nutritionRepo := repository.NewNutritionRepository(db)
	userMealPeriodsRepo := repository.NewUserMealPeriodRepository(db)

	authService := service.NewAuthService(userRepo, jwtManager, logger)
	userService := service.NewUserService(userRepo)
	activityService := service.NewActivityService(activityRepo)
	nutritionService := service.NewNutritionService(nutritionRepo, activityRepo, userRepo, recipeRepo)

	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	activityHandler := handler.NewActivityHandler(activityService)
	nutritionHandler := handler.NewNutritionHandler(nutritionService)
	recipeAdminHandler := adminHandler.NewRecipeAdminHandler(recipeRepo)
	userMealPeriodsAdminHandler := adminHandler.NewUserMealPeriodAdminHandler(userMealPeriodsRepo)

	r := router.Setup(logger, jwtManager, authHandler, userHandler, activityHandler, nutritionHandler, recipeAdminHandler, userMealPeriodsAdminHandler)

	addr := fmt.Sprintf(":%d", cfg.AppPort)
	logger.Info().Str("addr", addr).Msg("starting server")
	if err := r.Run(addr); err != nil {
		logger.Fatal().Err(err).Msg("server failed")
	}
}
