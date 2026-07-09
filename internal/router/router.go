package router

import (
	"github.com/gin-gonic/gin"
	"github.com/ikukhar/refuel-backend/internal/handler"
	"github.com/ikukhar/refuel-backend/internal/middleware"
	"github.com/ikukhar/refuel-backend/pkg/jwt"
	"github.com/rs/zerolog"
)

func Setup(
	logger zerolog.Logger,
	jwtManager *jwt.Manager,
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger(logger))

	api := r.Group("/api/v1")
	{
		api.GET("/health", handler.HealthCheck)

		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.Refresh)
		}

		protected := api.Group("")
		protected.Use(middleware.AuthRequired(jwtManager))
		{
			user := protected.Group("/user")
			{
				user.GET("/profile", userHandler.GetProfile)
				user.PUT("/profile", userHandler.UpdateProfile)
			}
		}
	}

	return r
}
