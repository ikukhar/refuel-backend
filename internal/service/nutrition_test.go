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

func TestNutritionService_BuildFromPeriods_DistributesCalories(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRecipe := mocks.NewMockRecipeRepository(ctrl)
	svc := NewNutritionService(nil, nil, nil, mockRecipe, nil)

	nutrition := &model.DailyNutrition{
		UserID:         1,
		CaloriesTarget: 2000,
		Status:         "baseline",
	}
	periods := []model.MealPeriod{
		{MealType: model.MealBreakfast, Name: "Завтрак", StartHour: 7, StartMinute: 0, CaloriesPercent: 25},
		{MealType: model.MealLunch, Name: "Обед", StartHour: 12, StartMinute: 0, CaloriesPercent: 35},
		{MealType: model.MealDinner, Name: "Ужин", StartHour: 18, StartMinute: 0, CaloriesPercent: 25},
		{MealType: model.MealSnack, Name: "Перекус", StartHour: 10, StartMinute: 0, CaloriesPercent: 15},
	}

	mockRecipe.EXPECT().
		FindByMealTypeExcludeIDs(gomock.Any(), gomock.Any()).
		Return([]model.Recipe{
			{Title: "Recipe", Calories: 200, ProteinG: 10, FatG: 5, CarbsG: 30},
		}, nil).AnyTimes()

	resp := svc.buildFromPeriods(nutrition, periods)

	require.Len(t, resp.Meals, 4)
	assert.Equal(t, 500.0, resp.Meals[0].CaloriesTarget)    // 2000 * 25% = 500
	assert.Equal(t, 700.0, resp.Meals[1].CaloriesTarget)    // 2000 * 35% = 700
	assert.Equal(t, 500.0, resp.Meals[2].CaloriesTarget)    // 2000 * 25% = 500
	assert.Equal(t, 300.0, resp.Meals[3].CaloriesTarget)    // 2000 * 15% = 300
}

func TestNutritionService_BuildFromStored_UsesRecipeIDsMap(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRecipe := mocks.NewMockRecipeRepository(ctrl)
	svc := NewNutritionService(nil, nil, nil, mockRecipe, nil)

	nutrition := &model.DailyNutrition{
		UserID:         1,
		CaloriesTarget: 2000,
		Status:         "baseline",
		RecipeIDs:      `{"breakfast":[10],"lunch":[20],"dinner":[30]}`,
	}
	periods := []model.MealPeriod{
		{MealType: model.MealBreakfast, Name: "Завтрак", StartHour: 7, StartMinute: 0, CaloriesPercent: 25},
		{MealType: model.MealLunch, Name: "Обед", StartHour: 12, StartMinute: 0, CaloriesPercent: 35},
		{MealType: model.MealDinner, Name: "Ужин", StartHour: 18, StartMinute: 0, CaloriesPercent: 25},
	}

	mockRecipe.EXPECT().
		FindByIDs(gomock.Any()).
		Return([]model.Recipe{
			{ID: 10, Title: "Oatmeal", MealType: model.MealBreakfast, Calories: 400, ProteinG: 12, FatG: 8, CarbsG: 50},
			{ID: 20, Title: "Soup", MealType: model.MealLunch, Calories: 600, ProteinG: 25, FatG: 10, CarbsG: 40},
			{ID: 30, Title: "Fish", MealType: model.MealDinner, Calories: 500, ProteinG: 35, FatG: 15, CarbsG: 10},
		}, nil)

	resp, err := svc.buildFromStored(nutrition, periods)
	require.NoError(t, err)
	require.Len(t, resp.Meals, 3)

	assert.Equal(t, 500.0, resp.Meals[0].CaloriesTarget) // 2000 * 25%
	assert.Equal(t, 400, resp.Meals[0].TotalCalories)    // oatmeal 400 ≤ 500 cap
	assert.Equal(t, "Oatmeal", resp.Meals[0].Dishes[0].Title)

	assert.Equal(t, 700.0, resp.Meals[1].CaloriesTarget) // 2000 * 35%
	assert.Equal(t, 600, resp.Meals[1].TotalCalories)

	assert.Equal(t, 500.0, resp.Meals[2].CaloriesTarget) // 2000 * 25%
}

func TestNutritionService_PickRecipesForMeal_RespectsMaxCalories(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRecipe := mocks.NewMockRecipeRepository(ctrl)
	svc := NewNutritionService(nil, nil, nil, mockRecipe, nil)

	mockRecipe.EXPECT().
		FindByMealTypeExcludeIDs("breakfast", gomock.Any()).
		Return([]model.Recipe{
			{ID: 1, Title: "Big", MealType: model.MealBreakfast, Calories: 900, ProteinG: 10, FatG: 5, CarbsG: 30},
			{ID: 2, Title: "Small", MealType: model.MealBreakfast, Calories: 200, ProteinG: 10, FatG: 5, CarbsG: 30},
			{ID: 3, Title: "Tiny", MealType: model.MealBreakfast, Calories: 100, ProteinG: 10, FatG: 5, CarbsG: 30},
		}, nil)

	dishes := svc.pickRecipesForMeal(model.MealBreakfast, 400, nil)

	totalCal := 0
	for _, d := range dishes {
		totalCal += d.Calories
	}
	assert.LessOrEqual(t, totalCal, 400, "total must not exceed target of 400")
}

func TestNutritionService_RefreshMeal_UsesPeriodTarget(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNutr := mocks.NewMockDailyNutritionRepository(ctrl)
	mockActivity := mocks.NewMockActivityRepository(ctrl)
	mockUser := mocks.NewMockUserRepository(ctrl)
	mockRecipe := mocks.NewMockRecipeRepository(ctrl)
	mockPeriod := mocks.NewMockMealPeriodRepository(ctrl)

	svc := NewNutritionService(mockNutr, mockActivity, mockUser, mockRecipe, mockPeriod)

	now := time.Now().Truncate(24 * time.Hour)

	mockNutr.EXPECT().
		FindByUserAndDate(gomock.Any(), uint(1), now).
		Return(&model.DailyNutrition{
			UserID:         1,
			Date:           now,
			CaloriesTarget: 2000,
			Status:         "baseline",
			RecipeIDs:      `{"breakfast":[5],"lunch":[10],"dinner":[15]}`,
		}, nil)

	mockPeriod.EXPECT().
		FindByUserID(uint(1)).
		Return([]model.MealPeriod{
			{MealType: model.MealBreakfast, Name: "Завтрак", StartHour: 7, StartMinute: 0, CaloriesPercent: 25},
			{MealType: model.MealLunch, Name: "Обед", StartHour: 12, StartMinute: 0, CaloriesPercent: 35},
			{MealType: model.MealDinner, Name: "Ужин", StartHour: 18, StartMinute: 0, CaloriesPercent: 25},
		}, nil)

	// RefreshMeal will exclude lunch[10] and dinner[15] IDs, only pick new breakfast
	mockRecipe.EXPECT().
		FindByMealTypeExcludeIDs("breakfast", gomock.Any()).
		Return([]model.Recipe{
			{ID: 2, Title: "New Breakfast", MealType: model.MealBreakfast, Calories: 300, ProteinG: 10, FatG: 5, CarbsG: 30},
		}, nil)

	// buildFromStored loads all recipes from the updated RecipeIDs map
	mockRecipe.EXPECT().
		FindByIDs(gomock.Any()).
		Return([]model.Recipe{
			{ID: 2, Title: "New Breakfast", MealType: model.MealBreakfast, Calories: 300, ProteinG: 10, FatG: 5, CarbsG: 30},
			{ID: 10, Title: "Lunch", MealType: model.MealLunch, Calories: 600, ProteinG: 25, FatG: 10, CarbsG: 40},
			{ID: 15, Title: "Dinner", MealType: model.MealDinner, Calories: 500, ProteinG: 35, FatG: 15, CarbsG: 10},
		}, nil)

	mockNutr.EXPECT().
		Upsert(gomock.Any(), gomock.Any()).
		Return(nil)

	resp, err := svc.RefreshMeal(context.Background(), 1, "breakfast")
	require.NoError(t, err)
	require.Len(t, resp.Meals, 3)

	// breakfast target should be 2000 * 25% = 500, not hardcoded 600
	assert.Equal(t, 500.0, resp.Meals[0].CaloriesTarget)
	assert.Equal(t, "New Breakfast", resp.Meals[0].Dishes[0].Title)
	assert.LessOrEqual(t, resp.Meals[0].TotalCalories, 500)
}

func TestNutritionService_Rebalance_FixesDeficit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRecipe := mocks.NewMockRecipeRepository(ctrl)
	svc := NewNutritionService(nil, nil, nil, mockRecipe, nil)

	var recipes []model.Recipe
	for i := 1; i <= 20; i++ {
		recipes = append(recipes, model.Recipe{
			ID:       uint(i),
			Calories: 100,
		})
	}

	mockRecipe.EXPECT().
		FindByMealTypeExcludeIDs(gomock.Any(), gomock.Any()).
		Return(recipes, nil).AnyTimes()

	nutrition := &model.DailyNutrition{
		UserID:         1,
		CaloriesTarget: 2000,
		Status:         "baseline",
	}
	periods := []model.MealPeriod{
		{MealType: model.MealBreakfast, Name: "Завтрак", StartHour: 7, StartMinute: 0, CaloriesPercent: 25},
		{MealType: model.MealLunch, Name: "Обед", StartHour: 12, StartMinute: 0, CaloriesPercent: 35},
		{MealType: model.MealDinner, Name: "Ужин", StartHour: 18, StartMinute: 0, CaloriesPercent: 25},
		{MealType: model.MealSnack, Name: "Перекус", StartHour: 10, StartMinute: 0, CaloriesPercent: 15},
	}

	resp := svc.buildFromPeriods(nutrition, periods)

	require.Len(t, resp.Meals, 4)
	total := 0
	for _, m := range resp.Meals {
		total += m.TotalCalories
	}
	assert.InDelta(t, 2000, float64(total), 100)
}

func TestNutritionService_Rebalance_WithinTolerance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRecipe := mocks.NewMockRecipeRepository(ctrl)
	svc := NewNutritionService(nil, nil, nil, mockRecipe, nil)

	mockRecipe.EXPECT().
		FindByMealTypeExcludeIDs(gomock.Any(), gomock.Any()).
		Return([]model.Recipe{
			{ID: 1, Title: "R1", MealType: model.MealBreakfast, Calories: 500},
			{ID: 2, Title: "R2", MealType: model.MealLunch, Calories: 700},
			{ID: 3, Title: "R3", MealType: model.MealDinner, Calories: 500},
			{ID: 4, Title: "R4", MealType: model.MealSnack, Calories: 300},
		}, nil).AnyTimes()

	nutrition := &model.DailyNutrition{
		UserID:         1,
		CaloriesTarget: 2000,
		Status:         "baseline",
	}
	periods := []model.MealPeriod{
		{MealType: model.MealBreakfast, Name: "Завтрак", StartHour: 7, StartMinute: 0, CaloriesPercent: 25},
		{MealType: model.MealLunch, Name: "Обед", StartHour: 12, StartMinute: 0, CaloriesPercent: 35},
		{MealType: model.MealDinner, Name: "Ужин", StartHour: 18, StartMinute: 0, CaloriesPercent: 25},
		{MealType: model.MealSnack, Name: "Перекус", StartHour: 10, StartMinute: 0, CaloriesPercent: 15},
	}

	resp := svc.buildFromPeriods(nutrition, periods)

	require.Len(t, resp.Meals, 4)
	total := 0
	for _, m := range resp.Meals {
		total += m.TotalCalories
	}
	assert.InDelta(t, 2000, float64(total), 100)
}
