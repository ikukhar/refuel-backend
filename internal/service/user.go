package service

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type UserService struct {
	userRepo       UserRepository
	mealPeriodRepo MealPeriodRepository
}

func NewUserService(userRepo UserRepository, mealPeriodRepo MealPeriodRepository) *UserService {
	return &UserService{userRepo: userRepo, mealPeriodRepo: mealPeriodRepo}
}

func (s *UserService) GetByID(ctx context.Context, id uint) (*UserResponse, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	periods, _ := s.mealPeriodRepo.FindByUserID(id)
	if len(periods) == 0 {
		periods = nil
	}

	return &UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		Name:        user.Name,
		Weight:      user.Weight,
		Height:      user.Height,
		Age:         user.Age,
		Gender:      user.Gender,
		MealPeriods: periods,
	}, nil
}

func (s *UserService) UpdateProfile(ctx context.Context, id uint, name *string, weight, height *float64, age *int, gender *string) error {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return err
	}

	if name != nil {
		user.Name = *name
	}
	if weight != nil {
		user.Weight = *weight
	}
	if height != nil {
		user.Height = *height
	}
	if age != nil {
		user.Age = *age
	}
	if gender != nil {
		user.Gender = *gender
	}

	return s.userRepo.Update(user)
}
