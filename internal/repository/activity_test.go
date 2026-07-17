package repository

import (
	"testing"
	"time"

	"github.com/ikukhar/refuel-backend/internal/model"
	"github.com/ikukhar/refuel-backend/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActivityRepository_CreateAndFindByUserID(t *testing.T) {
	db := testDB
	user := createUser(t, db)
	repo := NewActivityRepository(db)

	now := time.Now()
	activity := &model.Activity{
		UserID:    user.ID,
		Type:      "run",
		Distance:  testutil.PtrFloat64(5000),
		Duration:  testutil.PtrInt(1800),
		Calories:  testutil.PtrInt(350),
		StartedAt: now,
		Source:    "manual",
		SourceID:  "src-" + time.Now().Format("150405.000"),
	}

	err := repo.Create(activity)
	require.NoError(t, err)
	assert.NotZero(t, activity.ID)

	activities, err := repo.FindByUserID(user.ID, nil, nil, 10, 0)
	require.NoError(t, err)
	assert.Len(t, activities, 1)
	assert.Equal(t, model.ActivityRun, activities[0].Type)
}

func TestActivityRepository_FindByUserID_WithDateFilter(t *testing.T) {
	db := testDB
	user := createUser(t, db)
	repo := NewActivityRepository(db)

	oldDate := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	newDate := time.Now()

	require.NoError(t, repo.Create(&model.Activity{
		UserID:    user.ID,
		Type:      "walk",
		StartedAt: oldDate,
		SourceID:  "filter-old-" + time.Now().Format("150405.00001"),
	}))
	require.NoError(t, repo.Create(&model.Activity{
		UserID:    user.ID,
		Type:      "cycle",
		StartedAt: newDate,
		SourceID:  "filter-new-" + time.Now().Format("150405.00002"),
	}))

	filterDate := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	activities, err := repo.FindByUserID(user.ID, &filterDate, nil, 10, 0)
	require.NoError(t, err)
	assert.Len(t, activities, 1)
	assert.Equal(t, model.ActivityCycle, activities[0].Type)
}

func TestActivityRepository_FindBySourceID(t *testing.T) {
	db := testDB
	user := createUser(t, db)
	repo := NewActivityRepository(db)

	sourceID := "find-src-" + time.Now().Format("150405.000")

	activity := &model.Activity{
		UserID:    user.ID,
		Type:      "swim",
		StartedAt: time.Now(),
		Source:    "health_connect",
		SourceID:  sourceID,
	}
	require.NoError(t, repo.Create(activity))

	found, err := repo.FindBySourceID(sourceID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, model.ActivitySwim, found.Type)
}

func TestActivityRepository_FindBySourceID_NotFound(t *testing.T) {
	db := testDB
	repo := NewActivityRepository(db)

	_, err := repo.FindBySourceID("nonexistent-source-id")
	assert.Error(t, err)
}

var _ = testutil.PtrInt
