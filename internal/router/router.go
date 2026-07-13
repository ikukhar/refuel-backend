package router

import (
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ikukhar/refuel-backend/internal/handler"
	adminHandler "github.com/ikukhar/refuel-backend/internal/handler/admin"
	"github.com/ikukhar/refuel-backend/internal/middleware"
	"github.com/ikukhar/refuel-backend/pkg/jwt"
	"github.com/rs/zerolog"
)

func Setup(
	logger zerolog.Logger,
	jwtManager *jwt.Manager,
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	activityHandler *handler.ActivityHandler,
	nutritionHandler *handler.NutritionHandler,
	recipeAdminHandler *adminHandler.RecipeAdminHandler,
	userMealPeriodsAdminHandler *adminHandler.UserMealPeriodAdminHandler,
) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger(logger))

	tmpl := template.Must(template.New("").ParseGlob("templates/admin/*.html"))
	template.Must(tmpl.ParseGlob("templates/admin/*/*.html"))
	r.SetHTMLTemplate(tmpl)

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
			protected.GET("/activities", activityHandler.List)
			protected.POST("/activities", activityHandler.Create)

			protected.GET("/nutrition/today", nutritionHandler.GetToday)

			user := protected.Group("/user")
			{
				user.GET("/profile", userHandler.GetProfile)
				user.PUT("/profile", userHandler.UpdateProfile)
			}
		}
	}

	admin := r.Group("/admin")
	{
		admin.GET("/", dashboard)
		admin.GET("/recipes", recipeAdminHandler.List)
		admin.GET("/recipes/new", recipeAdminHandler.NewForm)
		admin.POST("/recipes", recipeAdminHandler.Create)
		admin.GET("/recipes/:id/edit", recipeAdminHandler.EditForm)
		admin.POST("/recipes/:id", recipeAdminHandler.Update)
		admin.DELETE("/recipes/:id", recipeAdminHandler.Delete)

		admin.GET("/user-meal-periods", userMealPeriodsAdminHandler.List)
		admin.GET("/user-meal-periods/new", userMealPeriodsAdminHandler.NewForm)
		admin.POST("/user-meal-periods", userMealPeriodsAdminHandler.Create)
		admin.GET("/user-meal-periods/:id/edit", userMealPeriodsAdminHandler.EditForm)
		admin.POST("/user-meal-periods/:id", userMealPeriodsAdminHandler.Update)
		admin.DELETE("/user-meal-periods/:id", userMealPeriodsAdminHandler.Delete)
	}

	return r
}

func dashboard(c *gin.Context) {
	c.HTML(http.StatusOK, "dashboard.html", nil)
}
