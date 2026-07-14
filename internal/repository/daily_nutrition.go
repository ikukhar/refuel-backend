package repository

import (
	"context"
	"time"

	"github.com/ikukhar/refuel-backend/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DailyNutritionRepository struct {
	db *gorm.DB
}

func NewDailyNutritionRepository(db *gorm.DB) *DailyNutritionRepository {
	return &DailyNutritionRepository{db: db}
}

func (r *DailyNutritionRepository) Upsert(ctx context.Context, n *model.DailyNutrition) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "date"}},
		UpdateAll: true,
	}).Create(n).Error
}

func (r *DailyNutritionRepository) FindByUserAndDate(ctx context.Context, userID uint, date time.Time) (*model.DailyNutrition, error) {
	var n model.DailyNutrition
	err := r.db.WithContext(ctx).Where("user_id = ? AND date = ?", userID, date).First(&n).Error
	if err != nil {
		return nil, err
	}
	return &n, nil
}
