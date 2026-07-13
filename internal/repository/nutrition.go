package repository

import (
	"time"

	"github.com/ikukhar/refuel-backend/internal/model"
	"gorm.io/gorm"
)

type NutritionRepository struct {
	db *gorm.DB
}

func NewNutritionRepository(db *gorm.DB) *NutritionRepository {
	return &NutritionRepository{db: db}
}

func (r *NutritionRepository) Upsert(n *model.DailyNutrition) error {
	return r.db.Where("user_id = ? AND date = ?", n.UserID, n.Date).
		Assign(n).
		FirstOrCreate(n).Error
}

func (r *NutritionRepository) FindByUserAndDate(userID uint, date time.Time) (*model.DailyNutrition, error) {
	var n model.DailyNutrition
	err := r.db.Where("user_id = ? AND date = ?", userID, date).First(&n).Error
	if err != nil {
		return nil, err
	}
	return &n, nil
}
