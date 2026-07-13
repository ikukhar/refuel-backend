package service

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/ikukhar/refuel-backend/internal/model"
	"github.com/ikukhar/refuel-backend/internal/repository"
)

type NutritionService struct {
	nutritionRepo *repository.NutritionRepository
	activityRepo  *repository.ActivityRepository
	userRepo      *repository.UserRepository
	recipeRepo    *repository.RecipeRepository
}

func NewNutritionService(
	nutritionRepo *repository.NutritionRepository,
	activityRepo *repository.ActivityRepository,
	userRepo *repository.UserRepository,
	recipeRepo *repository.RecipeRepository,
) *NutritionService {
	return &NutritionService{
		nutritionRepo: nutritionRepo,
		activityRepo:  activityRepo,
		userRepo:      userRepo,
		recipeRepo:    recipeRepo,
	}
}

type MealResponse struct {
	Time     string  `json:"time"`
	Dish     string  `json:"dish"`
	Calories int     `json:"calories"`
	Protein  float64 `json:"protein"`
	Fat      float64 `json:"fat"`
	Carbs    float64 `json:"carbs"`
}

type NutritionResponse struct {
	Date           string        `json:"date"`
	CaloriesTarget float64       `json:"calories_target"`
	ProteinG       float64       `json:"protein_g"`
	FatG           float64       `json:"fat_g"`
	CarbsG         float64       `json:"carbs_g"`
	Status         string        `json:"status"`
	Breakfast      *MealResponse `json:"breakfast"`
	Lunch          *MealResponse `json:"lunch"`
	Dinner         *MealResponse `json:"dinner"`
}

type mealSlot struct {
	mealType string
	minHour  int
	maxHour  int
	minMin   int
}

var mealSlots = map[string]mealSlot{
	"breakfast": {mealType: "breakfast", minHour: 7, maxHour: 9, minMin: 0},
	"lunch":     {mealType: "lunch", minHour: 12, maxHour: 14, minMin: 0},
	"dinner":    {mealType: "dinner", minHour: 17, maxHour: 20, minMin: 0},
}

func randomTime(slot mealSlot) string {
	h := slot.minHour + rand.Intn(slot.maxHour-slot.minHour+1)
	m := slot.minMin + rand.Intn(4)*15
	if m >= 60 {
		m = 0
	}
	return fmt.Sprintf("%d:%02d", h, m)
}

func (s *NutritionService) pickRecipe(mealType string) *MealResponse {
	recipes, err := s.recipeRepo.FindByMealType(mealType)
	if err != nil || len(recipes) == 0 {
		return nil
	}

	r := recipes[rand.Intn(len(recipes))]
	slot := mealSlots[mealType]

	return &MealResponse{
		Time:     randomTime(slot),
		Dish:     r.Title,
		Calories: r.Calories,
		Protein:  r.ProteinG,
		Fat:      r.FatG,
		Carbs:    r.CarbsG,
	}
}

func (s *NutritionService) GetToday(ctx context.Context, userID uint) (*NutritionResponse, error) {
	now := time.Now().Truncate(24 * time.Hour)

	existing, err := s.nutritionRepo.FindByUserAndDate(userID, now)
	if err == nil {
		return s.buildResponse(existing.CaloriesTarget, existing.ProteinG, existing.FatG, existing.CarbsG, existing.Status, existing.Date), nil
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	baselineCalories := 2000.0
	baselineProtein := 80.0
	baselineFat := 65.0
	baselineCarbs := 250.0

	if user.Weight != nil {
		baselineCalories = *user.Weight * 30
		baselineProtein = *user.Weight * 1.6
		baselineFat = *user.Weight * 0.8
		baselineCarbs = *user.Weight * 4.0
	}

	activities, err := s.activityRepo.FindByUserID(userID, &now, nil, 50, 0)
	if err != nil {
		return nil, err
	}

	status := "baseline"
	caloriesBurned := 0.0
	for _, a := range activities {
		if a.Calories != nil {
			caloriesBurned += float64(*a.Calories)
		}
	}

	if caloriesBurned > 0 {
		status = "final"
		baselineCalories += caloriesBurned * 0.5
	}

	nutrition := &model.DailyNutrition{
		UserID:         userID,
		Date:           now,
		CaloriesTarget: baselineCalories,
		ProteinG:       baselineProtein,
		FatG:           baselineFat,
		CarbsG:         baselineCarbs,
		Status:         status,
	}

	if err := s.nutritionRepo.Upsert(nutrition); err != nil {
		return nil, err
	}

	return s.buildResponse(baselineCalories, baselineProtein, baselineFat, baselineCarbs, status, now), nil
}

var validMeals = map[string]bool{"breakfast": true, "lunch": true, "dinner": true}

func (s *NutritionService) GetMeal(ctx context.Context, userID uint, mealType string) (*MealResponse, error) {
	if !validMeals[mealType] {
		return nil, fmt.Errorf("invalid meal: %s, allowed: breakfast, lunch, dinner", mealType)
	}

	meal := s.pickRecipe(mealType)
	if meal == nil {
		return nil, fmt.Errorf("no recipes found for %s", mealType)
	}
	return meal, nil
}

func (s *NutritionService) buildResponse(calories float64, protein float64, fat float64, carbs float64, status string, date time.Time) *NutritionResponse {
	return &NutritionResponse{
		Date:           date.Format("2006-01-02"),
		CaloriesTarget: calories,
		ProteinG:       protein,
		FatG:           fat,
		CarbsG:         carbs,
		Status:         status,
		Breakfast:      s.pickRecipe("breakfast"),
		Lunch:          s.pickRecipe("lunch"),
		Dinner:         s.pickRecipe("dinner"),
	}
}
