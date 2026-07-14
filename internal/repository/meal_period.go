package repository

import (
	"github.com/ikukhar/refuel-backend/internal/model"
	"gorm.io/gorm"
)

type MealPeriodRepository struct {
	db *gorm.DB
}

func NewMealPeriodRepository(db *gorm.DB) *MealPeriodRepository {
	return &MealPeriodRepository{db: db}
}

func (r *MealPeriodRepository) FindAll() ([]model.MealPeriod, error) {
	var periods []model.MealPeriod
	err := r.db.Order("user_id, start_hour, start_minute").Find(&periods).Error
	return periods, err
}

func (r *MealPeriodRepository) FindByUserID(userID uint) ([]model.MealPeriod, error) {
	var periods []model.MealPeriod
	err := r.db.Where("user_id = ?", userID).Order("start_hour, start_minute").Find(&periods).Error
	return periods, err
}

func (r *MealPeriodRepository) FindByID(id uint) (*model.MealPeriod, error) {
	var p model.MealPeriod
	err := r.db.First(&p, id).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *MealPeriodRepository) FindByUserIDAndMealType(userID uint, mealType model.MealType) (*model.MealPeriod, error) {
	var p model.MealPeriod
	err := r.db.Where("user_id = ? AND meal_type = ?", userID, mealType).First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *MealPeriodRepository) Create(p *model.MealPeriod) error {
	return r.db.Create(p).Error
}

func (r *MealPeriodRepository) Update(p *model.MealPeriod) error {
	return r.db.Save(p).Error
}

func (r *MealPeriodRepository) Delete(id uint) error {
	return r.db.Delete(&model.MealPeriod{}, id).Error
}

func (r *MealPeriodRepository) DeleteByUserIDAndMealType(userID uint, mealType model.MealType) error {
	return r.db.Where("user_id = ? AND meal_type = ?", userID, mealType).Delete(&model.MealPeriod{}).Error
}

func (r *MealPeriodRepository) DeleteByUserID(userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&model.MealPeriod{}).Error
}
