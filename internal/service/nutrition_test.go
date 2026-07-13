package service

import (
	"context"
	"testing"
	"time"

	"github.com/ikukhar/refuel-backend/internal/model"
	"github.com/ikukhar/refuel-backend/internal/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

func TestNutritionService_GetToday_CreatesBaseline(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNutritionRepo := mocks.NewMockNutritionRepository(ctrl)
	mockActivityRepo := mocks.NewMockActivityRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockRecipeRepo := mocks.NewMockRecipeRepository(ctrl)

	svc := NewNutritionService(mockNutritionRepo, mockActivityRepo, mockUserRepo, mockRecipeRepo)

	now := time.Now().Truncate(24 * time.Hour)

	mockNutritionRepo.EXPECT().
		FindByUserAndDate(uint(1), now).
		Return(nil, gorm.ErrRecordNotFound)

	mockUserRepo.EXPECT().
		FindByID(uint(1)).
		Return(&model.User{ID: 1, Name: "Test"}, nil)

	mockActivityRepo.EXPECT().
		FindByUserID(uint(1), &now, nil, 50, 0).
		Return([]model.Activity{}, nil)

	mockNutritionRepo.EXPECT().
		Upsert(gomock.Any()).
		Return(nil)

	mockRecipeRepo.EXPECT().
		FindByMealType("breakfast").
		Return([]model.Recipe{{Title: "Каша", MealType: model.MealBreakfast, Calories: 300, ProteinG: 10, FatG: 5, CarbsG: 50}}, nil)

	mockRecipeRepo.EXPECT().
		FindByMealType("lunch").
		Return([]model.Recipe{{Title: "Суп", MealType: model.MealLunch, Calories: 400, ProteinG: 20, FatG: 10, CarbsG: 40}}, nil)

	mockRecipeRepo.EXPECT().
		FindByMealType("dinner").
		Return(nil, nil)

	resp, err := svc.GetToday(context.Background(), 1)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "baseline", resp.Status)
	assert.Equal(t, 2000.0, resp.CaloriesTarget)
	assert.NotNil(t, resp.Breakfast)
	assert.Equal(t, "Каша", resp.Breakfast.Dish)
	assert.NotNil(t, resp.Lunch)
	assert.Equal(t, "Суп", resp.Lunch.Dish)
	assert.Nil(t, resp.Dinner)
}

func TestNutritionService_GetToday_WithWeight(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNutritionRepo := mocks.NewMockNutritionRepository(ctrl)
	mockActivityRepo := mocks.NewMockActivityRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockRecipeRepo := mocks.NewMockRecipeRepository(ctrl)

	svc := NewNutritionService(mockNutritionRepo, mockActivityRepo, mockUserRepo, mockRecipeRepo)

	now := time.Now().Truncate(24 * time.Hour)
	w := 80.0

	mockNutritionRepo.EXPECT().
		FindByUserAndDate(uint(1), now).
		Return(nil, gorm.ErrRecordNotFound)

	mockUserRepo.EXPECT().
		FindByID(uint(1)).
		Return(&model.User{ID: 1, Name: "Test", Weight: &w}, nil)

	mockActivityRepo.EXPECT().
		FindByUserID(uint(1), &now, nil, 50, 0).
		Return([]model.Activity{}, nil)

	mockNutritionRepo.EXPECT().
		Upsert(gomock.Any()).
		DoAndReturn(func(n *model.DailyNutrition) error {
			assert.Equal(t, 2400.0, n.CaloriesTarget)
			assert.Equal(t, 128.0, n.ProteinG)
			return nil
		})

	mockRecipeRepo.EXPECT().
		FindByMealType(gomock.Any()).
		Return(nil, nil).Times(3)

	resp, err := svc.GetToday(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, 2400.0, resp.CaloriesTarget)
	assert.Equal(t, 128.0, resp.ProteinG)
}

func TestNutritionService_GetMeal_Valid(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNutritionRepo := mocks.NewMockNutritionRepository(ctrl)
	mockActivityRepo := mocks.NewMockActivityRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockRecipeRepo := mocks.NewMockRecipeRepository(ctrl)

	svc := NewNutritionService(mockNutritionRepo, mockActivityRepo, mockUserRepo, mockRecipeRepo)

	mockRecipeRepo.EXPECT().
		FindByMealType("breakfast").
		Return([]model.Recipe{{Title: "Омлет", MealType: model.MealBreakfast, Calories: 250, ProteinG: 18, FatG: 15, CarbsG: 5}}, nil)

	meal, err := svc.GetMeal(context.Background(), 1, "breakfast")

	require.NoError(t, err)
	require.NotNil(t, meal)
	assert.Equal(t, "Омлет", meal.Dish)
	assert.Contains(t, meal.Time, ":")
}

func TestNutritionService_GetMeal_InvalidType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := NewNutritionService(nil, nil, nil, nil)

	meal, err := svc.GetMeal(context.Background(), 1, "brunch")

	require.Error(t, err)
	assert.Nil(t, meal)
	assert.Contains(t, err.Error(), "invalid meal")
}
