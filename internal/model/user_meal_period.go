package model

import "time"

type UserMealPeriod struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	UserID          uint      `gorm:"index;not null" json:"user_id"`
	User            User      `gorm:"foreignKey:UserID" json:"-"`
	MealType        MealType  `gorm:"not null" json:"meal_type"`
	StartHour       int       `gorm:"not null" json:"start_hour"`
	StartMinute     int       `gorm:"not null;default:0" json:"start_minute"`
	CaloriesPercent float64   `gorm:"not null;default:0" json:"calories_percent"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
