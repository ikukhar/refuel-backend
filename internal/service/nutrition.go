package service

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/ikukhar/refuel-backend/internal/model"
)

type NutritionService struct {
	nutritionRepo NutritionRepository
	activityRepo  ActivityRepository
	userRepo      UserRepository
	recipeRepo    RecipeRepository
}

func NewNutritionService(
	nutritionRepo NutritionRepository,
	activityRepo ActivityRepository,
	userRepo UserRepository,
	recipeRepo RecipeRepository,
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
	minHour int
	maxHour int
	minMin  int
}

var defaultMealSlots = map[model.MealType]mealSlot{
	model.MealBreakfast: {minHour: 7, maxHour: 9, minMin: 0},
	model.MealLunch:     {minHour: 12, maxHour: 14, minMin: 0},
	model.MealDinner:    {minHour: 17, maxHour: 20, minMin: 0},
}

func (s *NutritionService) pickRecipe(mealType model.MealType) *MealResponse {
	recipes, err := s.recipeRepo.FindByMealType(string(mealType))
	if err != nil || len(recipes) == 0 {
		return nil
	}

	r := recipes[rand.Intn(len(recipes))]
	slot, ok := defaultMealSlots[mealType]
	if !ok {
		return nil
	}

	return &MealResponse{
		Time:     randomTime(slot),
		Dish:     r.Title,
		Calories: r.Calories,
		Protein:  r.ProteinG,
		Fat:      r.FatG,
		Carbs:    r.CarbsG,
	}
}

func randomTime(slot mealSlot) string {
	h := slot.minHour + rand.Intn(slot.maxHour-slot.minHour+1)
	m := slot.minMin + rand.Intn(4)*15
	if m >= 60 {
		m = 0
	}
	return fmt.Sprintf("%d:%02d", h, m)
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

	if user.Weight > 0 {
		baselineCalories = user.Weight * 30
		baselineProtein = user.Weight * 1.6
		baselineFat = user.Weight * 0.8
		baselineCarbs = user.Weight * 4.0
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

var validMeals = map[model.MealType]bool{
	model.MealBreakfast: true,
	model.MealLunch:     true,
	model.MealDinner:    true,
}

func (s *NutritionService) GetMeal(ctx context.Context, userID uint, mealType string) (*MealResponse, error) {
	mt := model.MealType(mealType)
	if !validMeals[mt] {
		return nil, fmt.Errorf("invalid meal: %s, allowed: breakfast, lunch, dinner", mealType)
	}

	meal := s.pickRecipe(mt)
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
		Breakfast:      s.pickRecipe(model.MealBreakfast),
		Lunch:          s.pickRecipe(model.MealLunch),
		Dinner:         s.pickRecipe(model.MealDinner),
	}
}
