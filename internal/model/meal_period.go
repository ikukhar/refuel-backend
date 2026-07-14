package model

import "time"

type MealPeriod struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	UserID          uint      `gorm:"index;not null" json:"user_id"`
	User            User      `gorm:"foreignKey:UserID" json:"-"`
	MealType        MealType  `gorm:"not null" json:"meal_type"`
	Name            string    `gorm:"not null" json:"name"`
	StartHour       int       `gorm:"not null" json:"start_hour"`
	StartMinute     int       `gorm:"not null;default:0" json:"start_minute"`
	SortOrder       int       `gorm:"not null;default:0" json:"sort_order"`
	CaloriesPercent float64   `gorm:"not null;default:0" json:"calories_percent"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

var DefaultMealPeriods = []MealPeriod{
	{Name: "Завтрак", MealType: MealBreakfast, StartHour: 7, StartMinute: 0, SortOrder: 0, CaloriesPercent: 25},
	{Name: "Первый перекус", MealType: MealSnack, StartHour: 10, StartMinute: 0, SortOrder: 1, CaloriesPercent: 10},
	{Name: "Обед", MealType: MealLunch, StartHour: 12, StartMinute: 0, SortOrder: 2, CaloriesPercent: 35},
	{Name: "Второй перекус", MealType: MealSnack, StartHour: 15, StartMinute: 0, SortOrder: 3, CaloriesPercent: 10},
	{Name: "Ужин", MealType: MealDinner, StartHour: 17, StartMinute: 0, SortOrder: 4, CaloriesPercent: 20},
}
