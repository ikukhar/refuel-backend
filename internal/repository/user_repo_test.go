package repository

import (
	"testing"

	"github.com/ikukhar/refuel-backend/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestUserRepository_CreateAndFind(t *testing.T) {
	db := testDB
	repo := NewUserRepository(db)

	user := &model.User{
		Email:    "find-test@test.com",
		Password: "hashed",
		Name:     "Find Test",
	}
	err := repo.Create(user)
	require.NoError(t, err)
	assert.NotZero(t, user.ID)

	found, err := repo.FindByID(user.ID)
	require.NoError(t, err)
	assert.Equal(t, user.Email, found.Email)
	assert.Equal(t, user.Name, found.Name)
}

func TestUserRepository_FindByEmail(t *testing.T) {
	db := testDB
	repo := NewUserRepository(db)

	user := createUser(t, db)

	found, err := repo.FindByEmail(user.Email)
	require.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)
}

func TestUserRepository_FindByEmail_NotFound(t *testing.T) {
	db := testDB
	repo := NewUserRepository(db)

	_, err := repo.FindByEmail("nonexistent@test.com")
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestUserRepository_FindByID_NotFound(t *testing.T) {
	db := testDB
	repo := NewUserRepository(db)

	_, err := repo.FindByID(99999)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestUserRepository_Update(t *testing.T) {
	db := testDB
	repo := NewUserRepository(db)

	user := createUser(t, db)
	user.Name = "Updated Name"
	user.Weight = 80

	err := repo.Update(user)
	require.NoError(t, err)

	found, err := repo.FindByID(user.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", found.Name)
	assert.Equal(t, float64(80), found.Weight)
}
