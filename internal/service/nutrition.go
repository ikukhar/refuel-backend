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

type dailyLoad struct {
	Date time.Time
	Load float64
}

func effectiveLoad(history []dailyLoad, today time.Time) float64 {
	var total float64
	for _, day := range history {
		daysAgo := int(today.Sub(day.Date).Hours() / 24)
		if daysAgo > 3 {
			continue
		}
		weight := math.Pow(0.5, float64(daysAgo))
		total += day.Load * weight
	}
	return total
}

type NutritionService struct {
	nutritionRepo  DailyNutritionRepository
	activityRepo   ActivityRepository
	userRepo       UserRepository
	recipeRepo     RecipeRepository
	mealPeriodRepo MealPeriodRepository
}

func NewNutritionService(
	nutritionRepo DailyNutritionRepository,
	activityRepo ActivityRepository,
	userRepo UserRepository,
	recipeRepo RecipeRepository,
	mealPeriodRepo MealPeriodRepository,
) *NutritionService {
	return &NutritionService{
		nutritionRepo:  nutritionRepo,
		activityRepo:   activityRepo,
		userRepo:       userRepo,
		recipeRepo:     recipeRepo,
		mealPeriodRepo: mealPeriodRepo,
	}
}

type DishResponse struct {
	RecipeID uint   `json:"recipe_id"`
	Title    string `json:"title"`
	Calories int    `json:"calories"`
	MealType string `json:"meal_type"`
}

type MealSlotResponse struct {
	Time           string         `json:"time"`
	MealType       string         `json:"meal_type"`
	MealName       string         `json:"meal_name"`
	CaloriesTarget float64        `json:"calories_target"`
	Dishes         []DishResponse `json:"dishes"`
	TotalCalories  int            `json:"total_calories"`
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
	base := 10*weight + 6.25*height - 5*float64(age)

	if gender == "male" {
		return base + 5
	}
	return base - 161
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
			MealType: string(mealType),
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

func sumDishes(dishes []DishResponse) (totalCalories int) {
	for _, d := range dishes {
		totalCalories += d.Calories
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
	beginOfDay := time.Now().Truncate(24 * time.Hour)

	existing, err := s.nutritionRepo.FindByUserAndDate(ctx, userID, beginOfDay)
	if err == nil {
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

	threeDaysAgo := beginOfDay.Add(-3 * 24 * time.Hour)
	activities, err := s.activityRepo.FindByUserID(userID, &threeDaysAgo, nil, 200, 0)
	if err != nil {
		return nil, err
	}

	loadByDate := make(map[time.Time]float64)
	for _, a := range activities {
		if a.Calories == nil {
			continue
		}
		date := a.StartedAt.Truncate(24 * time.Hour)
		loadByDate[date] += float64(*a.Calories)
	}

	var history []dailyLoad
	for date, load := range loadByDate {
		history = append(history, dailyLoad{Date: date, Load: load})
	}

	effLoad := effectiveLoad(history, beginOfDay)

	status := "baseline"
	if effLoad > 0 {
		status = "final"
		tdee += effLoad
	}

	protein, fat, carbs := distributeMacros(tdee)

	periods, _ := s.mealPeriodRepo.FindByUserID(userID)

	nutrition := &model.DailyNutrition{
		UserID:         userID,
		Date:           beginOfDay,
		CaloriesTarget: math.Round(tdee*10) / 10,
		ProteinG:       protein,
		FatG:           fat,
		CarbsG:         carbs,
		Status:         status,
	}

	resp := s.buildResponse(nutrition, periods)

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
	mt := model.MealType(mealType)
	if !validMeals[mt] {
		return nil, fmt.Errorf("invalid meal: %s, allowed: breakfast, lunch, dinner", mealType)
	}

	periods, err := s.mealPeriodRepo.FindByUserID(userID)
	if err != nil || len(periods) == 0 {
		return s.pickMealFromDefaults(mt)
	}

	var slot *model.MealPeriod
	for _, p := range periods {
		if p.MealType == mt {
			slot = &p
			break
		}
	}
	if slot == nil {
		return nil, fmt.Errorf("no meal period found for %s", mealType)
	}

	dishes := s.pickRecipesForMeal(mt, 600, nil)
	if len(dishes) == 0 {
		return nil, fmt.Errorf("no recipes found for %s", mealType)
	}

	totalCal := sumDishes(dishes)
	return &MealSlotResponse{
		Time:          mealSlotTime(*slot),
		MealType:      string(mt),
		MealName:      slot.Name,
		Dishes:        dishes,
		TotalCalories: totalCal,
	}, nil
}

func (s *NutritionService) pickMealFromDefaults(mt model.MealType) (*MealSlotResponse, error) {
	dishes := s.pickRecipesForMeal(mt, 600, nil)
	if len(dishes) == 0 {
		return nil, fmt.Errorf("no recipes found for %s", mt)
	}

	slot := model.DefaultMealPeriods[0]
	for _, p := range model.DefaultMealPeriods {
		if p.MealType == mt {
			slot = p
			break
		}
	}

	totalCal := sumDishes(dishes)
	return &MealSlotResponse{
		Time:          mealSlotTime(slot),
		MealType:      string(mt),
		MealName:      mt.Name(),
		Dishes:        dishes,
		TotalCalories: totalCal,
	}, nil
}

func (s *NutritionService) buildResponse(n *model.DailyNutrition, periods []model.MealPeriod) *NutritionResponse {
	excludeIDs := parsePreviousRecipeIDs(n.PreviousRecipeIDs)

	var meals []MealSlotResponse
	for _, period := range periods {
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

		totalCal := sumDishes(dishes)
		meals = append(meals, MealSlotResponse{
			Time:           mealSlotTime(period),
			MealType:       string(period.MealType),
			MealName:       period.Name,
			CaloriesTarget: math.Round(mealTarget*10) / 10,
			Dishes:         dishes,
			TotalCalories:  totalCal,
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
	periods, _ := s.mealPeriodRepo.FindByUserID(n.UserID)
	return s.buildResponse(n, periods)
}
