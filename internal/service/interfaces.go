package service

import (
	"context"
	"time"

	"github.com/ikukhar/refuel-backend/internal/model"
)

type UserRepository interface {
	Create(user *model.User) error
	FindByID(id uint) (*model.User, error)
	FindByEmail(email string) (*model.User, error)
	Update(user *model.User) error
}

type ActivityRepository interface {
	Create(activity *model.Activity) error
	FindByUserID(userID uint, from, to *time.Time, limit, offset int) ([]model.Activity, error)
	FindBySourceID(sourceID string) (*model.Activity, error)
}

type DailyNutritionRepository interface {
	Upsert(ctx context.Context, n *model.DailyNutrition) error
	FindByUserAndDate(ctx context.Context, userID uint, date time.Time) (*model.DailyNutrition, error)
}

type RecipeRepository interface {
	Create(recipe *model.Recipe) error
	FindByID(id uint) (*model.Recipe, error)
	FindAll() ([]model.Recipe, error)
	FindByMealType(mealType string) ([]model.Recipe, error)
	FindByMealTypeExcludeIDs(mealType string, excludeIDs []uint) ([]model.Recipe, error)
	Update(recipe *model.Recipe) error
	Delete(id uint) error
}

type MealPeriodRepository interface {
	FindByUserID(userID uint) ([]model.MealPeriod, error)
	Create(p *model.MealPeriod) error
	Update(p *model.MealPeriod) error
	Delete(id uint) error
	DeleteByUserID(userID uint) error
}
