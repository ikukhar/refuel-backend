package model

import "time"

type UserMealPeriod struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"index;not null" json:"user_id"`
	User        User      `gorm:"foreignKey:UserID" json:"-"`
	MealType    MealType  `gorm:"not null" json:"meal_type"`
	StartHour   int       `gorm:"not null" json:"start_hour"`
	StartMinute int       `gorm:"not null;default:0" json:"start_minute"`
	EndHour     int       `gorm:"not null" json:"end_hour"`
	EndMinute   int       `gorm:"not null;default:0" json:"end_minute"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
