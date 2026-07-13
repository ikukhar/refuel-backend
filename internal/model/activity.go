package model

import (
	"fmt"
	"time"
)

type ActivityType string

const (
	ActivityRun       ActivityType = "run"
	ActivityWalk      ActivityType = "walk"
	ActivityCycle     ActivityType = "cycle"
	ActivitySwim      ActivityType = "swim"
	ActivityHike      ActivityType = "hike"
	ActivityWorkout   ActivityType = "workout"
	ActivityYoga      ActivityType = "yoga"
	ActivityOther     ActivityType = "other"
)

var validActivityTypes = map[ActivityType]struct{}{
	ActivityRun:     {},
	ActivityWalk:    {},
	ActivityCycle:   {},
	ActivitySwim:    {},
	ActivityHike:    {},
	ActivityWorkout: {},
	ActivityYoga:    {},
	ActivityOther:   {},
}

func (t ActivityType) Valid() bool {
	_, ok := validActivityTypes[t]
	return ok
}

func ValidActivityTypes() []string {
	types := make([]string, 0, len(validActivityTypes))
	for t := range validActivityTypes {
		types = append(types, string(t))
	}
	return types
}

func ParseActivityType(s string) (ActivityType, error) {
	t := ActivityType(s)
	if !t.Valid() {
		return "", fmt.Errorf("invalid activity type: %s, allowed: %v", s, ValidActivityTypes())
	}
	return t, nil
}

type Source string

const (
	SourceManual        Source = "manual"
	SourceHealthConnect Source = "health_connect"
)

var validSources = map[Source]struct{}{
	SourceManual:        {},
	SourceHealthConnect: {},
}

func (s Source) Valid() bool {
	_, ok := validSources[s]
	return ok
}

func ValidSources() []string {
	srcs := make([]string, 0, len(validSources))
	for s := range validSources {
		srcs = append(srcs, string(s))
	}
	return srcs
}

func ParseSource(s string) (Source, error) {
	v := Source(s)
	if !v.Valid() {
		return "", fmt.Errorf("invalid source: %s, allowed: %v", s, ValidSources())
	}
	return v, nil
}

type Activity struct {
	ID        uint         `gorm:"primaryKey" json:"id"`
	UserID    uint         `gorm:"index;not null" json:"user_id"`
	Type      ActivityType `gorm:"not null" json:"type"`
	Distance  *float64     `gorm:"default:null" json:"distance"`
	Duration  *int         `gorm:"default:null" json:"duration"`
	Elevation *float64     `gorm:"default:null" json:"elevation"`
	Calories  *int         `gorm:"default:null" json:"calories"`
	StartedAt time.Time    `gorm:"not null" json:"started_at"`
	Source    Source       `gorm:"default:'manual'" json:"source"`
	SourceID  string       `gorm:"uniqueIndex;not null" json:"source_id"`
	CreatedAt time.Time    `json:"created_at"`
}
