package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ikukhar/refuel-backend/internal/model"
	"github.com/ikukhar/refuel-backend/internal/service/mocks"
	"github.com/ikukhar/refuel-backend/pkg/jwt"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func TestAuthService_Register_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	logger := zerolog.Nop()

	weight := 75.0
	height := 180.0

	mockUserRepo.EXPECT().
		FindByEmail("test@test.com").
		Return(nil, gorm.ErrRecordNotFound)

	mockUserRepo.EXPECT().
		Create(gomock.Any()).
		DoAndReturn(func(user *model.User) error {
			user.ID = 1
			user.CreatedAt = time.Now()
			return nil
		})

	jwtManager := jwt.NewManager("test-secret", 15*time.Minute, 72*time.Hour)
	svc := NewAuthService(mockUserRepo, jwtManager, logger)

	resp, err := svc.Register(context.Background(), "test@test.com", "pass123", "Test", &weight, &height, nil)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "test@test.com", resp.User.Email)
	assert.Equal(t, "Test", resp.User.Name)
	assert.Equal(t, &weight, resp.User.Weight)
	assert.Equal(t, &height, resp.User.Height)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
}

func TestAuthService_Register_DuplicateEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	logger := zerolog.Nop()

	mockUserRepo.EXPECT().
		FindByEmail("existing@test.com").
		Return(&model.User{Email: "existing@test.com"}, nil)

	jwtManager := jwt.NewManager("test-secret", 15*time.Minute, 72*time.Hour)
	svc := NewAuthService(mockUserRepo, jwtManager, logger)

	resp, err := svc.Register(context.Background(), "existing@test.com", "pass123", "Test", nil, nil, nil)

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, "email already registered", err.Error())
}

func TestAuthService_Login_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	logger := zerolog.Nop()

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.DefaultCost)
	mockUserRepo.EXPECT().
		FindByEmail("test@test.com").
		Return(&model.User{ID: 1, Email: "test@test.com", Password: string(hashedPassword), Name: "Test"}, nil)

	jwtManager := jwt.NewManager("test-secret", 15*time.Minute, 72*time.Hour)
	svc := NewAuthService(mockUserRepo, jwtManager, logger)

	resp, err := svc.Login(context.Background(), "test@test.com", "correct-password")

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "test@test.com", resp.User.Email)
	assert.NotEmpty(t, resp.AccessToken)
}

func TestAuthService_Login_InvalidPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	logger := zerolog.Nop()

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("real-password"), bcrypt.DefaultCost)
	mockUserRepo.EXPECT().
		FindByEmail("test@test.com").
		Return(&model.User{Email: "test@test.com", Password: string(hashedPassword)}, nil)

	jwtManager := jwt.NewManager("test-secret", 15*time.Minute, 72*time.Hour)
	svc := NewAuthService(mockUserRepo, jwtManager, logger)

	resp, err := svc.Login(context.Background(), "test@test.com", "wrong-password")

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, "invalid email or password", err.Error())
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	logger := zerolog.Nop()

	mockUserRepo.EXPECT().
		FindByEmail("notfound@test.com").
		Return(nil, gorm.ErrRecordNotFound)

	jwtManager := jwt.NewManager("test-secret", 15*time.Minute, 72*time.Hour)
	svc := NewAuthService(mockUserRepo, jwtManager, logger)

	resp, err := svc.Login(context.Background(), "notfound@test.com", "pass")

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, "invalid email or password", err.Error())
}

func TestAuthService_Register_RepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	logger := zerolog.Nop()

	mockUserRepo.EXPECT().
		FindByEmail("test@test.com").
		Return(nil, errors.New("db connection failed"))

	jwtManager := jwt.NewManager("test-secret", 15*time.Minute, 72*time.Hour)
	svc := NewAuthService(mockUserRepo, jwtManager, logger)

	resp, err := svc.Register(context.Background(), "test@test.com", "pass123", "Test", nil, nil, nil)

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "db connection failed")
}
