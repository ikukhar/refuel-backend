package service

import (
	"context"
	"testing"
	"time"

	"github.com/ikukhar/refuel-backend/internal/model"
	"github.com/ikukhar/refuel-backend/internal/service/mocks"
	"github.com/ikukhar/refuel-backend/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func newTestActivityService(ctrl *gomock.Controller) (*ActivityService, *mocks.MockActivityRepository, *mocks.MockUserRepository) {
	mockAct := mocks.NewMockActivityRepository(ctrl)
	mockUser := mocks.NewMockUserRepository(ctrl)
	svc := NewActivityService(mockAct, mockUser)
	return svc, mockAct, mockUser
}

func TestActivityService_Create_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockActivityRepo, _ := newTestActivityService(ctrl)

	now := time.Now()
	input := CreateActivityInput{
		Type:      "run",
		Distance:  testutil.PtrFloat64(5000),
		Duration:  testutil.PtrInt(1800),
		Calories:  testutil.PtrInt(350),
		StartedAt: now,
		Source:    "manual",
		SourceID:  "unique-id-123",
	}

	mockActivityRepo.EXPECT().
		FindBySourceID("unique-id-123").
		Return(nil, assert.AnError)

	mockActivityRepo.EXPECT().
		Create(gomock.Any()).
		DoAndReturn(func(a *model.Activity) error {
			a.ID = 1
			a.CreatedAt = time.Now()
			return nil
		})

	resp, created, err := svc.Create(context.Background(), 1, input)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.True(t, created)
	assert.Equal(t, "run", resp.Type)
	assert.Equal(t, testutil.PtrFloat64(5000), resp.Distance)
	assert.Equal(t, testutil.PtrInt(1800), resp.Duration)
	assert.Equal(t, "manual", resp.Source)
}

func TestActivityService_Create_Idempotency(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockActivityRepo, _ := newTestActivityService(ctrl)

	existing := &model.Activity{
		ID:     5,
		UserID: 1,
		Type:   "run",
		Source: "health_connect",
	}

	mockActivityRepo.EXPECT().
		FindBySourceID("existing-id").
		Return(existing, nil)

	input := CreateActivityInput{
		Type:      "run",
		StartedAt: time.Now(),
		Source:    "health_connect",
		SourceID:  "existing-id",
	}

	resp, created, err := svc.Create(context.Background(), 1, input)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.False(t, created)
	assert.Equal(t, uint(5), resp.ID)
}

func TestActivityService_Create_MissingSourceID(t *testing.T) {
	svc := NewActivityService(nil, nil)
	input := CreateActivityInput{Type: "run", StartedAt: time.Now()}

	resp, created, err := svc.Create(context.Background(), 1, input)

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.False(t, created)
	assert.Equal(t, "source_id is required", err.Error())
}

func TestActivityService_Create_InvalidType(t *testing.T) {
	svc := NewActivityService(nil, nil)

	input := CreateActivityInput{
		Type:      "flying",
		StartedAt: time.Now(),
		SourceID:  "test-id",
	}

	resp, created, err := svc.Create(context.Background(), 1, input)

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.False(t, created)
	assert.Contains(t, err.Error(), "invalid activity type")
	assert.Contains(t, err.Error(), "run")
}

var _ = testutil.PtrInt
