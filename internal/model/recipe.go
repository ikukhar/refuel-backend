package model

import "time"

type Recipe struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Title       string    `gorm:"not null" json:"title"`
	Description string    `gorm:"type:text" json:"description"`
	Calories    int       `gorm:"not null" json:"calories"`
	ProteinG    float64   `gorm:"not null" json:"protein_g"`
	FatG        float64   `gorm:"not null" json:"fat_g"`
	CarbsG      float64   `gorm:"not null" json:"carbs_g"`
	ImageURL    *string   `gorm:"default:null" json:"image_url"`
	MealType    string    `gorm:"not null;default:'other'" json:"meal_type"`
	Steps       string    `gorm:"type:text" json:"steps"`
	Ingredients string    `gorm:"type:text" json:"ingredients"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
