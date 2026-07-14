package service

import (
	"fmt"

	"github.com/ikukhar/refuel-backend/internal/model"
)

type UpsertMealPeriodItem struct {
	MealType    string
	Name        string
	StartHour   int
	StartMinute int
}

type MealPeriodService struct {
	repo MealPeriodRepository
}

func NewMealPeriodService(repo MealPeriodRepository) *MealPeriodService {
	return &MealPeriodService{repo: repo}
}

func defaultCaloriesPercent(mt model.MealType) float64 {
	for _, d := range model.DefaultMealPeriods {
		if d.MealType == mt {
			return d.CaloriesPercent
		}
	}
	return 0
}

func (s *MealPeriodService) Upsert(userID uint, items []UpsertMealPeriodItem) ([]model.MealPeriod, error) {
	existing, err := s.repo.FindByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("find existing periods: %w", err)
	}

	existingByType := make(map[model.MealType]*model.MealPeriod, len(existing))
	for i := range existing {
		existingByType[existing[i].MealType] = &existing[i]
	}

	incomingTypes := make(map[model.MealType]struct{}, len(items))

	var result []model.MealPeriod

	for _, item := range items {
		mt := model.MealType(item.MealType)
		incomingTypes[mt] = struct{}{}

		if p, ok := existingByType[mt]; ok {
			p.Name = item.Name
			p.StartHour = item.StartHour
			p.StartMinute = item.StartMinute
			if err := s.repo.Update(p); err != nil {
				return nil, fmt.Errorf("update period %s: %w", mt, err)
			}
			result = append(result, *p)
		} else {
			p := &model.MealPeriod{
				UserID:          userID,
				MealType:        mt,
				Name:            item.Name,
				StartHour:       item.StartHour,
				StartMinute:     item.StartMinute,
				CaloriesPercent: defaultCaloriesPercent(mt),
			}
			if err := s.repo.Create(p); err != nil {
				return nil, fmt.Errorf("create period %s: %w", mt, err)
			}
			result = append(result, *p)
		}
	}

	for _, p := range existing {
		if _, ok := incomingTypes[p.MealType]; !ok {
			if err := s.repo.Delete(p.ID); err != nil {
				return nil, fmt.Errorf("delete period %s (id=%d): %w", p.MealType, p.ID, err)
			}
		}
	}

	return result, nil
}
