package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	if err := db.AutoMigrate(&model.User{}, &model.Activity{}, &model.DailyNutrition{}, &model.Recipe{}, &model.MealPeriod{}); err != nil {
		logger.Fatal().Err(err).Msg("failed to run migrations")
	}

	// ensure unique constraint on (user_id, date) for ON CONFLICT DO UPDATE
	db.Exec(`DELETE FROM daily_nutritions WHERE id NOT IN (SELECT DISTINCT ON (user_id, date) id FROM daily_nutritions)`)
	db.Exec(`DROP INDEX IF EXISTS idx_nutrition_user_date`)
	db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_nutrition_user_date ON daily_nutritions (user_id, date)`)

	recipeRepo := repository.NewRecipeRepository(db)
	if err := recipeRepo.SeedRecipes(); err != nil {
		logger.Fatal().Err(err).Msg("failed to seed recipes")
	}

	jwtManager := jwt.NewManager(cfg.JWTSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL)

	userRepo := repository.NewUserRepository(db)
	activityRepo := repository.NewActivityRepository(db)
	nutritionRepo := repository.NewDailyNutritionRepository(db)
	userMealPeriodsRepo := repository.NewMealPeriodRepository(db)

	authService := service.NewAuthService(userRepo, jwtManager, logger)
	userService := service.NewUserService(userRepo, userMealPeriodsRepo)
	activityService := service.NewActivityService(activityRepo)
	nutritionService := service.NewNutritionService(nutritionRepo, activityRepo, userRepo, recipeRepo, userMealPeriodsRepo)
	mealPeriodService := service.NewMealPeriodService(userMealPeriodsRepo)

	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	activityHandler := handler.NewActivityHandler(activityService)
	nutritionHandler := handler.NewNutritionHandler(nutritionService)
	mealPeriodHandler := handler.NewMealPeriodHandler(mealPeriodService)
	recipeAdminHandler := adminHandler.NewRecipeAdminHandler(recipeRepo)
	userMealPeriodsAdminHandler := adminHandler.NewMealPeriodAdminHandler(userMealPeriodsRepo)

	r := router.Setup(cfg, logger, jwtManager, authHandler, userHandler, activityHandler, nutritionHandler, mealPeriodHandler, recipeAdminHandler, userMealPeriodsAdminHandler)

	addr := fmt.Sprintf(":%d", cfg.AppPort)
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Info().Str("addr", addr).Msg("starting server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("server failed")
		}
	}()

	<-quit
	logger.Info().Msg("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal().Err(err).Msg("server forced to shutdown")
	}

	logger.Info().Msg("server exited")
}
