package model

import "time"

type DailyNutrition struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	UserID            uint      `gorm:"index:idx_nutrition_user_date,not null" json:"user_id"`
	Date              time.Time `gorm:"index:idx_nutrition_user_date,not null" json:"date"`
	CaloriesTarget    float64   `gorm:"not null" json:"calories_target"`
	ProteinG          float64   `gorm:"not null" json:"protein_g"`
	FatG              float64   `gorm:"not null" json:"fat_g"`
	CarbsG            float64   `gorm:"not null" json:"carbs_g"`
	Status            string    `gorm:"default:'baseline'" json:"status"`
	PreviousRecipeIDs string    `gorm:"type:text" json:"previous_recipe_ids"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
