package repository

import (
	"github.com/ikukhar/refuel-backend/internal/model"
	"gorm.io/gorm"
)

type UserMealPeriodRepository struct {
	db *gorm.DB
}

func NewUserMealPeriodRepository(db *gorm.DB) *UserMealPeriodRepository {
	return &UserMealPeriodRepository{db: db}
}

func (r *UserMealPeriodRepository) FindAll() ([]model.UserMealPeriod, error) {
	var periods []model.UserMealPeriod
	err := r.db.Order("user_id, start_hour, start_minute").Find(&periods).Error
	return periods, err
}

func (r *UserMealPeriodRepository) FindByUserID(userID uint) ([]model.UserMealPeriod, error) {
	var periods []model.UserMealPeriod
	err := r.db.Where("user_id = ?", userID).Order("start_hour, start_minute").Find(&periods).Error
	return periods, err
}

func (r *UserMealPeriodRepository) FindByID(id uint) (*model.UserMealPeriod, error) {
	var p model.UserMealPeriod
	err := r.db.First(&p, id).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *UserMealPeriodRepository) FindByUserIDAndMealType(userID uint, mealType model.MealType) (*model.UserMealPeriod, error) {
	var p model.UserMealPeriod
	err := r.db.Where("user_id = ? AND meal_type = ?", userID, mealType).First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *UserMealPeriodRepository) Create(p *model.UserMealPeriod) error {
	return r.db.Create(p).Error
}

func (r *UserMealPeriodRepository) Update(p *model.UserMealPeriod) error {
	return r.db.Save(p).Error
}

func (r *UserMealPeriodRepository) Delete(id uint) error {
	return r.db.Delete(&model.UserMealPeriod{}, id).Error
}

func (r *UserMealPeriodRepository) DeleteByUserIDAndMealType(userID uint, mealType model.MealType) error {
	return r.db.Where("user_id = ? AND meal_type = ?", userID, mealType).Delete(&model.UserMealPeriod{}).Error
}
