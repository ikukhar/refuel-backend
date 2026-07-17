package service

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/ikukhar/refuel-backend/internal/model"
)

type RecipeService struct {
	repo RecipeRepository
}

func NewRecipeService(repo RecipeRepository) *RecipeService {
	return &RecipeService{repo: repo}
}

type CreateRecipeInput struct {
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Calories    int            `json:"calories"`
	ProteinG    float64        `json:"protein_g"`
	FatG        float64        `json:"fat_g"`
	CarbsG      float64        `json:"carbs_g"`
	ImageURL    *string        `json:"image_url"`
	MealType    model.MealType `json:"meal_type"`
	Steps       []string       `json:"steps"`
	Ingredients []string       `json:"ingredients"`
}

type UpdateRecipeInput struct {
	Title       *string         `json:"title"`
	Description *string         `json:"description"`
	Calories    *int            `json:"calories"`
	ProteinG    *float64        `json:"protein_g"`
	FatG        *float64        `json:"fat_g"`
	CarbsG      *float64        `json:"carbs_g"`
	ImageURL    **string        `json:"image_url"`
	MealType    *model.MealType `json:"meal_type"`
	Steps       *[]string       `json:"steps"`
	Ingredients *[]string       `json:"ingredients"`
}

type RecipeResponse struct {
	ID          uint           `json:"id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Calories    int            `json:"calories"`
	ProteinG    float64        `json:"protein_g"`
	FatG        float64        `json:"fat_g"`
	CarbsG      float64        `json:"carbs_g"`
	ImageURL    *string        `json:"image_url"`
	MealType    model.MealType `json:"meal_type"`
	Steps       []string       `json:"steps"`
	Ingredients []string       `json:"ingredients"`
}

func (s *RecipeService) Create(ctx context.Context, input CreateRecipeInput) (*model.Recipe, error) {
	if input.Title == "" {
		return nil, errors.New("title is required")
	}

	stepsJSON := marshalSlice(input.Steps)
	ingredientsJSON := marshalSlice(input.Ingredients)

	recipe := &model.Recipe{
		Title:       input.Title,
		Description: input.Description,
		Calories:    input.Calories,
		ProteinG:    input.ProteinG,
		FatG:        input.FatG,
		CarbsG:      input.CarbsG,
		ImageURL:    input.ImageURL,
		MealType:    input.MealType,
		Steps:       stepsJSON,
		Ingredients: ingredientsJSON,
	}

	if err := s.repo.Create(recipe); err != nil {
		return nil, err
	}

	return recipe, nil
}

func (s *RecipeService) GetByID(ctx context.Context, id uint) (*model.Recipe, error) {
	return s.repo.FindByID(id)
}

func (s *RecipeService) List(ctx context.Context) ([]model.Recipe, error) {
	return s.repo.FindAll()
}

func (s *RecipeService) Update(ctx context.Context, id uint, input UpdateRecipeInput) (*model.Recipe, error) {
	recipe, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if input.Title != nil {
		recipe.Title = *input.Title
	}
	if input.Description != nil {
		recipe.Description = *input.Description
	}
	if input.Calories != nil {
		recipe.Calories = *input.Calories
	}
	if input.ProteinG != nil {
		recipe.ProteinG = *input.ProteinG
	}
	if input.FatG != nil {
		recipe.FatG = *input.FatG
	}
	if input.CarbsG != nil {
		recipe.CarbsG = *input.CarbsG
	}
	if input.ImageURL != nil {
		recipe.ImageURL = *input.ImageURL
	}
	if input.MealType != nil {
		recipe.MealType = *input.MealType
	}
	if input.Steps != nil {
		recipe.Steps = marshalSlice(*input.Steps)
	}
	if input.Ingredients != nil {
		recipe.Ingredients = marshalSlice(*input.Ingredients)
	}

	if err := s.repo.Update(recipe); err != nil {
		return nil, err
	}

	return recipe, nil
}

func (s *RecipeService) Delete(ctx context.Context, id uint) error {
	return s.repo.Delete(id)
}

func marshalSlice(s []string) string {
	if len(s) == 0 {
		return "[]"
	}
	b, err := json.Marshal(s)
	if err != nil {
		return "[]"
	}
	return string(b)
}
