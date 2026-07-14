package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ikukhar/refuel-backend/internal/model"
	"github.com/ikukhar/refuel-backend/internal/service"
	"github.com/ikukhar/refuel-backend/internal/service/mocks"
	"github.com/ikukhar/refuel-backend/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func setupNutritionRouter(t *testing.T) (*gin.Engine, *mocks.MockNutritionRepository, *mocks.MockActivityRepository, *mocks.MockUserRepository, *mocks.MockRecipeRepository) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	mockNutrition := mocks.NewMockNutritionRepository(ctrl)
	mockActivity := mocks.NewMockActivityRepository(ctrl)
	mockUser := mocks.NewMockUserRepository(ctrl)
	mockRecipe := mocks.NewMockRecipeRepository(ctrl)

	svc := service.NewNutritionService(mockNutrition, mockActivity, mockUser, mockRecipe)
	h := NewNutritionHandler(svc)

	r := gin.New()
	r.GET("/api/v1/nutrition/today", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		h.GetToday(c)
	})

	return r, mockNutrition, mockActivity, mockUser, mockRecipe
}

func TestNutritionHandler_GetToday_Error(t *testing.T) {
	r, mockN, _, mockU, _ := setupNutritionRouter(t)

	mockN.EXPECT().
		FindByUserAndDate(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, assert.AnError)

	mockU.EXPECT().
		FindByID(uint(1)).
		Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/nutrition/today", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestNutritionHandler_GetToday_WithMealParam(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRecipe := mocks.NewMockRecipeRepository(ctrl)
	svc := service.NewNutritionService(nil, nil, nil, mockRecipe)
	h := NewNutritionHandler(svc)

	mockRecipe.EXPECT().
		FindByMealTypeExcludeIDs("lunch", gomock.Any()).
		Return([]model.Recipe{
			{Title: "Суп", MealType: model.MealLunch, Calories: 400, ProteinG: 20, FatG: 10, CarbsG: 40},
		}, nil)

	r := gin.New()
	r.GET("/api/v1/nutrition/today", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		h.GetToday(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/nutrition/today?meal=lunch", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	lunch, ok := resp["lunch"].(map[string]interface{})
	require.True(t, ok)
	dishes, ok := lunch["dishes"].([]interface{})
	require.True(t, ok)
	require.Len(t, dishes, 1)
	firstDish := dishes[0].(map[string]interface{})
	assert.Equal(t, "Суп", firstDish["title"])
}

func TestNutritionHandler_GetToday_WithInvalidMeal(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := service.NewNutritionService(nil, nil, nil, nil)
	h := NewNutritionHandler(svc)

	r := gin.New()
	r.GET("/api/v1/nutrition/today", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		h.GetToday(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/nutrition/today?meal=invalid", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Contains(t, resp["error"], "invalid meal")
}

func TestNutritionHandler_GetToday_FullResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockN := mocks.NewMockNutritionRepository(ctrl)
	mockA := mocks.NewMockActivityRepository(ctrl)
	mockU := mocks.NewMockUserRepository(ctrl)
	mockR := mocks.NewMockRecipeRepository(ctrl)

	svc := service.NewNutritionService(mockN, mockA, mockU, mockR)
	h := NewNutritionHandler(svc)

	mockN.EXPECT().
		FindByUserAndDate(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, assert.AnError)

	mockU.EXPECT().
		FindByID(uint(1)).
		Return(&model.User{ID: 1, Name: "Test", Weight: 70, Height: 175, Age: 25, Gender: "female"}, nil)

	mockA.EXPECT().
		FindByUserID(uint(1), gomock.Any(), nil, 50, 0).
		Return([]model.Activity{}, nil)

	mockN.EXPECT().
		Upsert(gomock.Any(), gomock.Any()).
		Return(nil)

	mockR.EXPECT().
		FindByMealTypeExcludeIDs(gomock.Any(), gomock.Any()).
		Return(nil, nil).AnyTimes()

	mockR.EXPECT().
		FindByMealType(gomock.Any()).
		Return([]model.Recipe{{Title: "Default", MealType: model.MealBreakfast, Calories: 200, ProteinG: 10, FatG: 5, CarbsG: 30}}, nil).AnyTimes()

	r := gin.New()
	r.GET("/api/v1/nutrition/today", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		h.GetToday(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/nutrition/today", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "baseline", resp["status"])
	assert.InDelta(t, 1809.3, resp["calories_target"].(float64), 1)
	assert.NotEmpty(t, resp["meals"])
}

var _ = testutil.PtrFloat64
