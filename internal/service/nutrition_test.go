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

	mockNutritionRepo := mocks.NewMockDailyNutritionRepository(ctrl)
	mockActivityRepo := mocks.NewMockActivityRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockRecipeRepo := mocks.NewMockRecipeRepository(ctrl)

	mockMealPeriodRepo := mocks.NewMockMealPeriodRepository(ctrl)

	svc := NewNutritionService(mockNutritionRepo, mockActivityRepo, mockUserRepo, mockRecipeRepo, mockMealPeriodRepo)

	now := time.Now().Truncate(24 * time.Hour)

	mockNutritionRepo.EXPECT().
		FindByUserAndDate(gomock.Any(), uint(1), now).
		Return(nil, gorm.ErrRecordNotFound)

	mockUserRepo.EXPECT().
		FindByID(uint(1)).
		Return(&model.User{ID: 1, Name: "Test"}, nil)

	mockActivityRepo.EXPECT().
		FindByUserID(uint(1), gomock.Any(), nil, 200, 0).
		Return([]model.Activity{}, nil)

	mockMealPeriodRepo.EXPECT().
		FindByUserID(uint(1)).
		Return(nil, nil)

	mockNutritionRepo.EXPECT().
		Upsert(gomock.Any(), gomock.Any()).
		Return(nil)

	mockRecipeRepo.EXPECT().
		FindByMealTypeExcludeIDs(gomock.Any(), gomock.Any()).
		Return([]model.Recipe{}, nil).AnyTimes()

	mockRecipeRepo.EXPECT().
		FindByMealType(gomock.Any()).
		Return([]model.Recipe{}, nil).AnyTimes()

	resp, err := svc.GetToday(context.Background(), 1)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "baseline", resp.Status)
	assert.Equal(t, 2000.0, resp.CaloriesTarget)
	assert.Len(t, resp.Meals, 0)
}

func TestNutritionService_GetToday_WithWeight(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNutritionRepo := mocks.NewMockDailyNutritionRepository(ctrl)
	mockActivityRepo := mocks.NewMockActivityRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockRecipeRepo := mocks.NewMockRecipeRepository(ctrl)

	mockMealPeriodRepo := mocks.NewMockMealPeriodRepository(ctrl)

	svc := NewNutritionService(mockNutritionRepo, mockActivityRepo, mockUserRepo, mockRecipeRepo, mockMealPeriodRepo)

	now := time.Now().Truncate(24 * time.Hour)

	mockNutritionRepo.EXPECT().
		FindByUserAndDate(gomock.Any(), uint(1), now).
		Return(nil, gorm.ErrRecordNotFound)

	mockUserRepo.EXPECT().
		FindByID(uint(1)).
		Return(&model.User{ID: 1, Name: "Test", Weight: 80, Height: 180, Age: 30, Gender: "male"}, nil)

	mockActivityRepo.EXPECT().
		FindByUserID(uint(1), gomock.Any(), nil, 200, 0).
		Return([]model.Activity{}, nil)

	mockMealPeriodRepo.EXPECT().
		FindByUserID(uint(1)).
		Return(nil, nil)

	mockNutritionRepo.EXPECT().
		Upsert(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, n *model.DailyNutrition) error {
			assert.InDelta(t, 2136.0, n.CaloriesTarget, 1)
			assert.InDelta(t, 160.2, n.ProteinG, 1)
			return nil
		})

	mockRecipeRepo.EXPECT().
		FindByMealTypeExcludeIDs(gomock.Any(), gomock.Any()).
		Return(nil, nil).AnyTimes()

	mockRecipeRepo.EXPECT().
		FindByMealType(gomock.Any()).
		Return([]model.Recipe{{Title: "Default", MealType: model.MealBreakfast, Calories: 200, ProteinG: 10, FatG: 5, CarbsG: 30}}, nil).AnyTimes()

	resp, err := svc.GetToday(context.Background(), 1)
	require.NoError(t, err)
	assert.InDelta(t, 2136.0, resp.CaloriesTarget, 1)
	assert.InDelta(t, 160.2, resp.ProteinG, 1)
}

func TestNutritionService_GetMeal_Valid(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNutritionRepo := mocks.NewMockDailyNutritionRepository(ctrl)
	mockActivityRepo := mocks.NewMockActivityRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockRecipeRepo := mocks.NewMockRecipeRepository(ctrl)

	mockMealPeriodRepo := mocks.NewMockMealPeriodRepository(ctrl)

	svc := NewNutritionService(mockNutritionRepo, mockActivityRepo, mockUserRepo, mockRecipeRepo, mockMealPeriodRepo)

	mockMealPeriodRepo.EXPECT().
		FindByUserID(uint(1)).
		Return([]model.MealPeriod{
			{MealType: model.MealBreakfast, Name: "Завтрак", StartHour: 7, StartMinute: 0},
		}, nil)

	mockRecipeRepo.EXPECT().
		FindByMealTypeExcludeIDs("breakfast", gomock.Any()).
		Return([]model.Recipe{{Title: "Омлет", MealType: model.MealBreakfast, Calories: 250, ProteinG: 18, FatG: 15, CarbsG: 5}}, nil)

	meal, err := svc.GetMeal(context.Background(), 1, "breakfast")

	require.NoError(t, err)
	require.NotNil(t, meal)
	require.Len(t, meal.Dishes, 1)
	assert.Equal(t, "Омлет", meal.Dishes[0].Title)
	assert.Contains(t, meal.Time, ":")
}

func TestNutritionService_GetMeal_InvalidType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := NewNutritionService(nil, nil, nil, nil, nil)

	meal, err := svc.GetMeal(context.Background(), 1, "brunch")

	require.Error(t, err)
	assert.Nil(t, meal)
	assert.Contains(t, err.Error(), "invalid meal")
}
