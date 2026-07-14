package service

import (
	"context"
	"errors"
	"strings"

	"github.com/ikukhar/refuel-backend/internal/model"
	"github.com/ikukhar/refuel-backend/pkg/jwt"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	userRepo   UserRepository
	jwtManager *jwt.Manager
	logger     zerolog.Logger
}

func NewAuthService(userRepo UserRepository, jwtManager *jwt.Manager, logger zerolog.Logger) *AuthService {
	return &AuthService{userRepo: userRepo, jwtManager: jwtManager, logger: logger}
}

type AuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	User         UserResponse `json:"user"`
}

type UserResponse struct {
	ID          uint              `json:"id"`
	Email       string            `json:"email"`
	Name        string            `json:"name"`
	Weight      float64           `json:"weight"`
	Height      float64           `json:"height"`
	Age         int               `json:"age"`
	Gender      string            `json:"gender"`
	MealPeriods []model.MealPeriod `json:"meal_periods"`
}

func (s *AuthService) Register(ctx context.Context, email, password, name string, weight, height float64, age int, gender string) (*AuthResponse, error) {
	existing, err := s.userRepo.FindByEmail(email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("email already registered")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Email:    email,
		Password: string(hashedPassword),
		Name:     name,
		Weight:   weight,
		Height:   height,
		Age:      age,
		Gender:   gender,
	}

	if err := s.userRepo.Create(user); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return nil, errors.New("email already registered")
		}
		return nil, err
	}

	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Email, user.TokenVersion)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID, user.Email, user.TokenVersion)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: UserResponse{
			ID:     user.ID,
			Email:  user.Email,
			Name:   user.Name,
			Weight: user.Weight,
			Height: user.Height,
			Age:    user.Age,
			Gender: user.Gender,
		},
	}, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid email or password")
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Email, user.TokenVersion)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID, user.Email, user.TokenVersion)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: UserResponse{
			ID:     user.ID,
			Email:  user.Email,
			Name:   user.Name,
			Weight: user.Weight,
			Height: user.Height,
			Age:    user.Age,
			Gender: user.Gender,
		},
	}, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*AuthResponse, error) {
	s.logger.Debug().Str("refresh_token", refreshToken).Msg("Refreshing access token")

	claims, err := s.jwtManager.ValidateToken(refreshToken)
	if err != nil {
		return nil, errors.New("invalid or expired refresh token")
	}

	user, err := s.userRepo.FindByID(claims.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if claims.TokenVersion != user.TokenVersion {
		return nil, errors.New("refresh token has been revoked")
	}

	user.TokenVersion++
	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Email, user.TokenVersion)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID, user.Email, user.TokenVersion)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		User: UserResponse{
			ID:     user.ID,
			Email:  user.Email,
			Name:   user.Name,
			Weight: user.Weight,
			Height: user.Height,
			Age:    user.Age,
			Gender: user.Gender,
		},
	}, nil
}
