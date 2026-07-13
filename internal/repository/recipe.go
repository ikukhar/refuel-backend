package repository

import (
	"github.com/ikukhar/refuel-backend/internal/model"
	"gorm.io/gorm"
)

type RecipeRepository struct {
	db *gorm.DB
}

func NewRecipeRepository(db *gorm.DB) *RecipeRepository {
	return &RecipeRepository{db: db}
}

func (r *RecipeRepository) Create(recipe *model.Recipe) error {
	return r.db.Create(recipe).Error
}

func (r *RecipeRepository) FindByID(id uint) (*model.Recipe, error) {
	var recipe model.Recipe
	err := r.db.First(&recipe, id).Error
	if err != nil {
		return nil, err
	}
	return &recipe, nil
}

func (r *RecipeRepository) FindAll() ([]model.Recipe, error) {
	var recipes []model.Recipe
	err := r.db.Order("created_at DESC").Find(&recipes).Error
	return recipes, err
}

func (r *RecipeRepository) FindByMealType(mealType string) ([]model.Recipe, error) {
	var recipes []model.Recipe
	err := r.db.Where("meal_type = ?", mealType).Order("created_at DESC").Find(&recipes).Error
	return recipes, err
}

func (r *RecipeRepository) Update(recipe *model.Recipe) error {
	return r.db.Save(recipe).Error
}

func (r *RecipeRepository) Delete(id uint) error {
	return r.db.Delete(&model.Recipe{}, id).Error
}
