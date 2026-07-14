package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/ikukhar/refuel-backend/internal/model"
)

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

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

type DishResponse struct {
	RecipeID uint    `json:"recipe_id"`
	Title    string  `json:"title"`
	Calories int     `json:"calories"`
	Protein  float64 `json:"protein"`
	Fat      float64 `json:"fat"`
	Carbs    float64 `json:"carbs"`
	ImageURL *string `json:"image_url,omitempty"`
}

type MealSlotResponse struct {
	Time          string         `json:"time"`
	MealType      string         `json:"meal_type"`
	MealName      string         `json:"meal_name"`
	CaloriesTarget float64       `json:"calories_target"`
	Dishes        []DishResponse `json:"dishes"`
	TotalCalories int            `json:"total_calories"`
	TotalProtein  float64        `json:"total_protein"`
	TotalFat      float64        `json:"total_fat"`
	TotalCarbs    float64        `json:"total_carbs"`
}

type NutritionResponse struct {
	Date           string            `json:"date"`
	CaloriesTarget float64           `json:"calories_target"`
	ProteinG       float64           `json:"protein_g"`
	FatG           float64           `json:"fat_g"`
	CarbsG         float64           `json:"carbs_g"`
	Status         string            `json:"status"`
	Meals          []MealSlotResponse `json:"meals"`
}

func calcBMR(weight, height float64, age int, gender string) float64 {
	if gender == "male" {
		return 10*weight + 6.25*height - 5*float64(age) + 5
	}
	return 10*weight + 6.25*height - 5*float64(age) - 161
}

func calcTDEE(bmr float64) float64 {
	return bmr * 1.2
}

func distributeMacros(tdee float64) (protein, fat, carbs float64) {
	protein = math.Round(tdee*0.3/4*10) / 10
	fat = math.Round(tdee*0.25/9*10) / 10
	carbs = math.Round(tdee*0.45/4*10) / 10
	return
}

func (s *NutritionService) pickRecipesForMeal(mealType model.MealType, targetCalories float64, excludeIDs []uint) []DishResponse {
	recipes, err := s.recipeRepo.FindByMealTypeExcludeIDs(string(mealType), excludeIDs)
	if err != nil || len(recipes) == 0 {
		recipes, err = s.recipeRepo.FindByMealType(string(mealType))
		if err != nil || len(recipes) == 0 {
			return nil
		}
	}

	perm := rng.Perm(len(recipes))
	var dishes []DishResponse
	totalCalories := 0
	threshold := int(math.Round(targetCalories * 0.85))

	for _, idx := range perm {
		if totalCalories >= threshold {
			break
		}
		r := recipes[idx]
		dishes = append(dishes, DishResponse{
			RecipeID: r.ID,
			Title:    r.Title,
			Calories: r.Calories,
			Protein:  r.ProteinG,
			Fat:      r.FatG,
			Carbs:    r.CarbsG,
			ImageURL: r.ImageURL,
		})
		totalCalories += r.Calories
	}

	return dishes
}

func mealSlotTime(period model.MealPeriod) string {
	h := period.StartHour + rng.Intn(2)
	m := period.StartMinute
	if h > 23 {
		h = 23
	}
	return fmt.Sprintf("%d:%02d", h, m)
}

func sumDishes(dishes []DishResponse) (totalCalories int, totalProtein, totalFat, totalCarbs float64) {
	for _, d := range dishes {
		totalCalories += d.Calories
		totalProtein += d.Protein
		totalFat += d.Fat
		totalCarbs += d.Carbs
	}
	return
}

func parsePreviousRecipeIDs(s string) []uint {
	if s == "" {
		return nil
	}
	var ids []uint
	if err := json.Unmarshal([]byte(s), &ids); err != nil {
		return nil
	}
	return ids
}

func (s *NutritionService) GetToday(ctx context.Context, userID uint) (*NutritionResponse, error) {
	now := time.Now().Truncate(24 * time.Hour)

	existing, err := s.nutritionRepo.FindByUserAndDate(ctx, userID, now)
	if err == nil {
		// context is used for cancellation; pass background as we already have data
		return s.buildResponseFromNutrition(existing), nil
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	var tdee float64
	if user.Weight > 0 && user.Height > 0 && user.Age > 0 && user.Gender != "" {
		bmr := calcBMR(user.Weight, user.Height, user.Age, user.Gender)
		tdee = calcTDEE(bmr)
	} else {
		tdee = 2000
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
		tdee += caloriesBurned * 0.5
	}

	protein, fat, carbs := distributeMacros(tdee)

	nutrition := &model.DailyNutrition{
		UserID:         userID,
		Date:           now,
		CaloriesTarget: math.Round(tdee*10) / 10,
		ProteinG:       protein,
		FatG:           fat,
		CarbsG:         carbs,
		Status:         status,
	}

	resp := s.buildResponse(nutrition)

	idsJSON, _ := json.Marshal(collectRecipeIDs(resp.Meals))
	nutrition.PreviousRecipeIDs = string(idsJSON)

	if err := s.nutritionRepo.Upsert(ctx, nutrition); err != nil {
		return nil, err
	}

	return resp, nil
}

func collectRecipeIDs(meals []MealSlotResponse) []uint {
	var ids []uint
	for _, m := range meals {
		for _, d := range m.Dishes {
			ids = append(ids, d.RecipeID)
		}
	}
	return ids
}

var validMeals = map[model.MealType]bool{
	model.MealBreakfast: true,
	model.MealLunch:     true,
	model.MealDinner:    true,
}

func (s *NutritionService) GetMeal(ctx context.Context, userID uint, mealType string) (*MealSlotResponse, error) {
	_ = ctx
	_ = userID
	mt := model.MealType(mealType)
	if !validMeals[mt] {
		return nil, fmt.Errorf("invalid meal: %s, allowed: breakfast, lunch, dinner", mealType)
	}

	dishes := s.pickRecipesForMeal(mt, 600, nil)
	if len(dishes) == 0 {
		return nil, fmt.Errorf("no recipes found for %s", mealType)
	}

	slot := model.DefaultMealPeriods[0]
	for _, p := range model.DefaultMealPeriods {
		if p.MealType == mt {
			slot = p
			break
		}
	}

	totalCal, totalProt, totalFat, totalCarbs := sumDishes(dishes)
	return &MealSlotResponse{
		Time:     mealSlotTime(slot),
		MealType: string(mt),
		MealName: mt.Name(),
		Dishes:   dishes,
		TotalCalories:  totalCal,
		TotalProtein:   totalProt,
		TotalFat:       totalFat,
		TotalCarbs:     totalCarbs,
	}, nil
}

func (s *NutritionService) buildResponse(n *model.DailyNutrition) *NutritionResponse {
	excludeIDs := parsePreviousRecipeIDs(n.PreviousRecipeIDs)

	var meals []MealSlotResponse
	for _, period := range model.DefaultMealPeriods {
		mealTarget := n.CaloriesTarget * period.CaloriesPercent / 100
		dishes := s.pickRecipesForMeal(period.MealType, mealTarget, excludeIDs)
		if len(dishes) == 0 {
			continue
		}

		dishesIDs := make([]uint, len(dishes))
		for i, d := range dishes {
			dishesIDs[i] = d.RecipeID
		}
		excludeIDs = append(excludeIDs, dishesIDs...)

		totalCal, totalProt, totalFat, totalCarbs := sumDishes(dishes)
		meals = append(meals, MealSlotResponse{
			Time:           mealSlotTime(period),
			MealType:       string(period.MealType),
			MealName:       period.Name,
			CaloriesTarget: math.Round(mealTarget*10) / 10,
			Dishes:         dishes,
			TotalCalories:  totalCal,
			TotalProtein:   math.Round(totalProt*10) / 10,
			TotalFat:       math.Round(totalFat*10) / 10,
			TotalCarbs:     math.Round(totalCarbs*10) / 10,
		})
	}

	return &NutritionResponse{
		Date:           n.Date.Format("2006-01-02"),
		CaloriesTarget: math.Round(n.CaloriesTarget*10) / 10,
		ProteinG:       math.Round(n.ProteinG*10) / 10,
		FatG:           math.Round(n.FatG*10) / 10,
		CarbsG:         math.Round(n.CarbsG*10) / 10,
		Status:         n.Status,
		Meals:          meals,
	}
}

func (s *NutritionService) buildResponseFromNutrition(n *model.DailyNutrition) *NutritionResponse {
	return s.buildResponse(n)
}
