package repository

import (
	"time"

	"github.com/ikukhar/refuel-backend/internal/model"
	"gorm.io/gorm"
)

type ActivityRepository struct {
	db *gorm.DB
}

func NewActivityRepository(db *gorm.DB) *ActivityRepository {
	return &ActivityRepository{db: db}
}

func (r *ActivityRepository) Create(activity *model.Activity) error {
	return r.db.Create(activity).Error
}

func (r *ActivityRepository) FindByUserID(userID uint, from, to *time.Time, limit, offset int) ([]model.Activity, error) {
	q := r.db.Where("user_id = ?", userID).Order("started_at DESC")

	if from != nil {
		q = q.Where("started_at >= ?", from)
	}
	if to != nil {
		q = q.Where("started_at <= ?", to)
	}
	if limit <= 0 {
		limit = 20
	}

	var activities []model.Activity
	err := q.Limit(limit).Offset(offset).Find(&activities).Error
	return activities, err
}

func (r *ActivityRepository) FindBySourceID(sourceID string) (*model.Activity, error) {
	var activity model.Activity
	err := r.db.Where("source_id = ?", sourceID).First(&activity).Error
	if err != nil {
		return nil, err
	}
	return &activity, nil
}

func (r *ActivityRepository) FindByID(id uint) (*model.Activity, error) {
	var activity model.Activity
	err := r.db.First(&activity, id).Error
	if err != nil {
		return nil, err
	}
	return &activity, nil
}
