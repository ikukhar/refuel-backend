package service

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type UserService struct {
	userRepo UserRepository
}

func NewUserService(userRepo UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) GetByID(ctx context.Context, id uint) (*UserResponse, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &UserResponse{
		ID:     user.ID,
		Email:  user.Email,
		Name:   user.Name,
		Weight: user.Weight,
		Height: user.Height,
		Age:    user.Age,
	}, nil
}

func (s *UserService) UpdateProfile(ctx context.Context, id uint, name *string, weight, height *float64, age *int) error {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return err
	}

	if name != nil {
		user.Name = *name
	}
	if weight != nil {
		user.Weight = weight
	}
	if height != nil {
		user.Height = height
	}
	if age != nil {
		user.Age = age
	}

	return s.userRepo.Update(user)
}
