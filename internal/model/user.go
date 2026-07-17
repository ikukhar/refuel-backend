package model

import "time"

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Email        string    `gorm:"uniqueIndex;not null" json:"email"`
	Password     string    `gorm:"not null" json:"-"`
	Name         string    `gorm:"not null" json:"name"`
	Weight       float64   `gorm:"not null;default:0" json:"weight"`
	Height       float64   `gorm:"not null;default:0" json:"height"`
	Age          int       `gorm:"not null;default:0" json:"age"`
	Gender       string    `gorm:"not null;default:''" json:"gender"`
	TokenVersion int       `gorm:"not null;default:0" json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
