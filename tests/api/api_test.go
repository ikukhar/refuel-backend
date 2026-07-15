package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ikukhar/refuel-backend/internal/config"
	"github.com/ikukhar/refuel-backend/internal/handler"
	adminHandler "github.com/ikukhar/refuel-backend/internal/handler/admin"
	"github.com/ikukhar/refuel-backend/internal/model"
	"github.com/ikukhar/refuel-backend/internal/router"
	"github.com/ikukhar/refuel-backend/internal/service"
	mockrepo "github.com/ikukhar/refuel-backend/internal/service/mocks"
	"github.com/ikukhar/refuel-backend/internal/testutil"
	"github.com/ikukhar/refuel-backend/pkg/jwt"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type apiSuite struct {
	r           *gin.Engine
	jwtManager  *jwt.Manager
	ctrl        *gomock.Controller
	mockUser    *mockrepo.MockUserRepository
	mockAct     *mockrepo.MockActivityRepository
	mockNutr    *mockrepo.MockDailyNutritionRepository
	mockRecipe  *mockrepo.MockRecipeRepository
	mockMealPeriod *mockrepo.MockMealPeriodRepository
}

func setupAPI(t *testing.T) *apiSuite {
	t.Helper()
	gin.SetMode(gin.TestMode)

	wd, _ := os.Getwd()
	os.Chdir("../..")
	t.Cleanup(func() { os.Chdir(wd) })

	ctrl := gomock.NewController(t)
	logger := zerolog.Nop()

	jwtManager := jwt.NewManager("test-secret", 15*time.Minute, 72*time.Hour)

	mockUser := mockrepo.NewMockUserRepository(ctrl)
	mockAct := mockrepo.NewMockActivityRepository(ctrl)
	mockNutr := mockrepo.NewMockDailyNutritionRepository(ctrl)
	mockRecipe := mockrepo.NewMockRecipeRepository(ctrl)
	mockMealPeriod := mockrepo.NewMockMealPeriodRepository(ctrl)

	userSvc := service.NewUserService(mockUser, mockMealPeriod)
	authSvc := service.NewAuthService(mockUser, jwtManager, logger)
	activitySvc := service.NewActivityService(mockAct)
	nutritionSvc := service.NewNutritionService(mockNutr, mockAct, mockUser, mockRecipe, mockMealPeriod)
	mealPeriodSvc := service.NewMealPeriodService(mockMealPeriod)

	authH := handler.NewAuthHandler(authSvc)
	userH := handler.NewUserHandler(userSvc)
	activityH := handler.NewActivityHandler(activitySvc)
	nutritionH := handler.NewNutritionHandler(nutritionSvc)
	mealPeriodH := handler.NewMealPeriodHandler(mealPeriodSvc)
	recipeH := handler.NewRecipeHandler(mockRecipe)
	recipeAdminH := adminHandler.NewRecipeAdminHandler(nil)
	userMealPeriodsAdminH := adminHandler.NewMealPeriodAdminHandler(nil)

	cfg := &config.Config{AdminUser: "admin", AdminPass: "admin"}

	r := router.Setup(cfg, logger, jwtManager, authH, userH, activityH, nutritionH, mealPeriodH, recipeH, recipeAdminH, userMealPeriodsAdminH)

	return &apiSuite{
		r:          r,
		jwtManager: jwtManager,
		ctrl:       ctrl,
		mockUser:   mockUser,
		mockAct:    mockAct,
		mockNutr:   mockNutr,
		mockRecipe: mockRecipe,
		mockMealPeriod: mockMealPeriod,
	}
}

func authHeader(t *testing.T, jm *jwt.Manager, userID uint) string {
	t.Helper()
	token, err := jm.GenerateAccessToken(userID, "test@test.com", 0)
	require.NoError(t, err)
	return "Bearer " + token
}

// ───────────────────── HEALTH ─────────────────────

func TestHealthEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/v1/health", handler.HealthCheck)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/health", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "ok", resp["status"])
}

// ───────────────────── AUTH ─────────────────────

func TestRegister_Success(t *testing.T) {
	s := setupAPI(t)
	defer s.ctrl.Finish()

	s.mockUser.EXPECT().
		FindByEmail("new@test.com").
		Return(nil, gorm.ErrRecordNotFound)

	s.mockUser.EXPECT().
		Create(gomock.Any()).
		DoAndReturn(func(u *model.User) error {
			u.ID = 10
			u.CreatedAt = time.Now()
			assert.Equal(t, "new@test.com", u.Email)
			return nil
		})

	body := `{"email":"new@test.com","password":"pass123","name":"New User","weight":70,"height":175,"age":25,"gender":"male"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	s.r.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "new@test.com", resp["user"].(map[string]interface{})["email"])
	assert.NotEmpty(t, resp["access_token"])
	assert.NotEmpty(t, resp["refresh_token"])
}

func TestRegister_DuplicateEmail(t *testing.T) {
	s := setupAPI(t)
	defer s.ctrl.Finish()

	s.mockUser.EXPECT().
		FindByEmail("dup@test.com").
		Return(&model.User{Email: "dup@test.com"}, nil)

	body := `{"email":"dup@test.com","password":"pass123","name":"Dup","weight":70,"height":175,"age":25,"gender":"female"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	s.r.ServeHTTP(w, req)

	require.Equal(t, http.StatusConflict, w.Code)
}

func TestRegister_ValidationError(t *testing.T) {
	s := setupAPI(t)
	defer s.ctrl.Finish()

	body := `{"email":"bad","password":"12","name":""}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	s.r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLogin_Success(t *testing.T) {
	s := setupAPI(t)
	defer s.ctrl.Finish()

	hash, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	require.NoError(t, err)

	s.mockUser.EXPECT().
		FindByEmail("a@b.com").
		Return(&model.User{ID: 5, Email: "a@b.com", Password: string(hash), Name: "Alice"}, nil)

	body := `{"email":"a@b.com","password":"secret123"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	s.r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "a@b.com", resp["user"].(map[string]interface{})["email"])
	assert.NotEmpty(t, resp["access_token"])
	assert.NotEmpty(t, resp["refresh_token"])
}

func TestLogin_WrongPassword(t *testing.T) {
	s := setupAPI(t)
	defer s.ctrl.Finish()

	hash, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)

	s.mockUser.EXPECT().
		FindByEmail("a@b.com").
		Return(&model.User{ID: 5, Email: "a@b.com", Password: string(hash)}, nil)

	body := `{"email":"a@b.com","password":"wrong"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	s.r.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestLogin_UserNotFound(t *testing.T) {
	s := setupAPI(t)
	defer s.ctrl.Finish()

	s.mockUser.EXPECT().
		FindByEmail("nobody@test.com").
		Return(nil, gorm.ErrRecordNotFound)

	body := `{"email":"nobody@test.com","password":"x"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	s.r.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRefresh_Success(t *testing.T) {
	s := setupAPI(t)
	defer s.ctrl.Finish()

	refresh, err := s.jwtManager.GenerateRefreshToken(3, "test@test.com", 0)
	require.NoError(t, err)

	s.mockUser.EXPECT().
		FindByID(uint(3)).
		Return(&model.User{ID: 3, Email: "test@test.com"}, nil)

	s.mockUser.EXPECT().
		Update(gomock.Any()).
		Return(nil)

	body := fmt.Sprintf(`{"refresh_token":"%s"}`, refresh)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	s.r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NotEmpty(t, resp["access_token"])
	assert.NotEmpty(t, resp["refresh_token"])
}

func TestRefresh_InvalidToken(t *testing.T) {
	s := setupAPI(t)
	defer s.ctrl.Finish()

	body := `{"refresh_token":"garbage-token"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	s.r.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

// ───────────────────── USER PROFILE ─────────────────────

func TestGetProfile_Success(t *testing.T) {
	s := setupAPI(t)
	defer s.ctrl.Finish()

	s.mockUser.EXPECT().
		FindByID(uint(1)).
		Return(&model.User{ID: 1, Email: "test@test.com", Name: "Test User"}, nil)

	s.mockMealPeriod.EXPECT().
		FindByUserID(uint(1)).
		Return(nil, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/user/profile", nil)
	req.Header.Set("Authorization", authHeader(t, s.jwtManager, 1))
	s.r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "test@test.com", resp["email"])
	assert.Equal(t, "Test User", resp["name"])
}

func TestGetProfile_Unauthenticated(t *testing.T) {
	s := setupAPI(t)
	defer s.ctrl.Finish()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/user/profile", nil)
	s.r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUpdateProfile_Success(t *testing.T) {
	s := setupAPI(t)
	defer s.ctrl.Finish()

	s.mockUser.EXPECT().
		FindByID(uint(1)).
		Return(&model.User{ID: 1, Email: "test@test.com", Name: "Old Name"}, nil)

	s.mockUser.EXPECT().
		Update(gomock.Any()).
		DoAndReturn(func(u *model.User) error {
			assert.Equal(t, "New Name", u.Name)
			return nil
		})

	body := `{"name":"New Name"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/user/profile", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(t, s.jwtManager, 1))
	s.r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "profile updated", resp["message"])
}

// ───────────────────── ACTIVITIES ─────────────────────

func TestCreateActivity_Success(t *testing.T) {
	s := setupAPI(t)
	defer s.ctrl.Finish()

	s.mockAct.EXPECT().
		FindBySourceID("test-src-99").
		Return(nil, gorm.ErrRecordNotFound)

	s.mockAct.EXPECT().
		Create(gomock.Any()).
		DoAndReturn(func(a *model.Activity) error {
			a.ID = 100
			a.CreatedAt = time.Now()
			assert.Equal(t, model.ActivityRun, a.Type)
			assert.Equal(t, "test-src-99", a.SourceID)
			return nil
		})

	body := `{"type":"run","started_at":"2026-07-13T10:00:00Z","source_id":"test-src-99","source":"manual"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/activities", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(t, s.jwtManager, 1))
	s.r.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(100), resp["id"])
}

func TestCreateActivity_Idempotent(t *testing.T) {
	s := setupAPI(t)
	defer s.ctrl.Finish()

	s.mockAct.EXPECT().
		FindBySourceID("existing-42").
		Return(&model.Activity{ID: 42, Type: model.ActivityWalk, Source: "manual"}, nil)

	body := `{"type":"walk","started_at":"2026-07-13T10:00:00Z","source_id":"existing-42","source":"manual"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/activities", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(t, s.jwtManager, 1))
	s.r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(42), resp["id"])
}

func TestCreateActivity_InvalidType(t *testing.T) {
	s := setupAPI(t)
	defer s.ctrl.Finish()

	body := `{"type":"teleport","started_at":"2026-07-13T10:00:00Z","source_id":"bad","source":"manual"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/activities", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(t, s.jwtManager, 1))
	s.r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateActivity_MissingSourceID(t *testing.T) {
	s := setupAPI(t)
	defer s.ctrl.Finish()

	body := `{"type":"run","started_at":"2026-07-13T10:00:00Z"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/activities", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(t, s.jwtManager, 1))
	s.r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestListActivities_Success(t *testing.T) {
	s := setupAPI(t)
	defer s.ctrl.Finish()

	s.mockAct.EXPECT().
		FindByUserID(uint(1), nil, nil, 20, 0).
		Return([]model.Activity{
			{ID: 1, UserID: 1, Type: model.ActivityRun, Source: "manual"},
			{ID: 2, UserID: 1, Type: model.ActivityCycle, Source: "health_connect"},
		}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/activities", nil)
	req.Header.Set("Authorization", authHeader(t, s.jwtManager, 1))
	s.r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp []interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Len(t, resp, 2)
}

func TestListActivities_FilterByFrom(t *testing.T) {
	s := setupAPI(t)
	defer s.ctrl.Finish()

	from := time.Date(2026, 7, 10, 0, 0, 0, 0, time.UTC)

	s.mockAct.EXPECT().
		FindByUserID(uint(1), &from, nil, 20, 0).
		Return([]model.Activity{
			{ID: 3, UserID: 1, Type: model.ActivityRun, Source: "manual"},
		}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/activities?from=2026-07-10", nil)
	req.Header.Set("Authorization", authHeader(t, s.jwtManager, 1))
	s.r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp []interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Len(t, resp, 1)
}

func TestListActivities_FilterByTo(t *testing.T) {
	s := setupAPI(t)
	defer s.ctrl.Finish()

	to := time.Date(2026, 7, 15, 23, 59, 59, 999999999, time.UTC)

	s.mockAct.EXPECT().
		FindByUserID(uint(1), nil, &to, 20, 0).
		Return([]model.Activity{
			{ID: 4, UserID: 1, Type: model.ActivityWalk, Source: "manual"},
		}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/activities?to=2026-07-15", nil)
	req.Header.Set("Authorization", authHeader(t, s.jwtManager, 1))
	s.r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp []interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Len(t, resp, 1)
}

func TestListActivities_FilterByFromAndTo(t *testing.T) {
	s := setupAPI(t)
	defer s.ctrl.Finish()

	from := time.Date(2026, 7, 10, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 7, 15, 23, 59, 59, 999999999, time.UTC)

	s.mockAct.EXPECT().
		FindByUserID(uint(1), &from, &to, 20, 0).
		Return([]model.Activity{
			{ID: 5, UserID: 1, Type: model.ActivityCycle, Source: "health_connect"},
		}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/activities?from=2026-07-10&to=2026-07-15", nil)
	req.Header.Set("Authorization", authHeader(t, s.jwtManager, 1))
	s.r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp []interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Len(t, resp, 1)
}

func TestDeleteActivity_Success(t *testing.T) {
	s := setupAPI(t)
	defer s.ctrl.Finish()

	s.mockAct.EXPECT().
		FindByID(uint(42)).
		Return(&model.Activity{ID: 42, UserID: 1, Type: model.ActivityRun, Source: "manual"}, nil)

	s.mockAct.EXPECT().
		Delete(uint(42)).
		Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/activities/42", nil)
	req.Header.Set("Authorization", authHeader(t, s.jwtManager, 1))
	s.r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "activity deleted", resp["message"])
}

func TestDeleteActivity_NotFound(t *testing.T) {
	s := setupAPI(t)
	defer s.ctrl.Finish()

	s.mockAct.EXPECT().
		FindByID(uint(999)).
		Return(nil, gorm.ErrRecordNotFound)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/activities/999", nil)
	req.Header.Set("Authorization", authHeader(t, s.jwtManager, 1))
	s.r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeleteActivity_WrongOwner(t *testing.T) {
	s := setupAPI(t)
	defer s.ctrl.Finish()

	s.mockAct.EXPECT().
		FindByID(uint(42)).
		Return(&model.Activity{ID: 42, UserID: 99, Type: model.ActivityRun, Source: "manual"}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/activities/42", nil)
	req.Header.Set("Authorization", authHeader(t, s.jwtManager, 1))
	s.r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeleteActivity_InvalidID(t *testing.T) {
	s := setupAPI(t)
	defer s.ctrl.Finish()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/activities/abc", nil)
	req.Header.Set("Authorization", authHeader(t, s.jwtManager, 1))
	s.r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

// ───────────────────── NUTRITION ─────────────────────

func TestGetNutrition_FullDay(t *testing.T) {
	s := setupAPI(t)
	defer s.ctrl.Finish()

	now := time.Now().Truncate(24 * time.Hour)

	s.mockNutr.EXPECT().
		FindByUserAndDate(gomock.Any(), uint(1), now).
		Return(nil, gorm.ErrRecordNotFound)

	s.mockUser.EXPECT().
		FindByID(uint(1)).
		Return(&model.User{ID: 1, Weight: 70, Height: 175, Age: 25, Gender: "female"}, nil)

	s.mockAct.EXPECT().
		FindByUserID(uint(1), gomock.Any(), nil, 200, 0).
		Return([]model.Activity{
			{ID: 1, Calories: testutil.PtrInt(300), StartedAt: time.Now()},
		}, nil)

	s.mockMealPeriod.EXPECT().
		FindByUserID(uint(1)).
		Return([]model.MealPeriod{
			{MealType: model.MealBreakfast, Name: "Завтрак", StartHour: 7, StartMinute: 0, CaloriesPercent: 25},
			{MealType: model.MealLunch, Name: "Обед", StartHour: 12, StartMinute: 0, CaloriesPercent: 35},
			{MealType: model.MealDinner, Name: "Ужин", StartHour: 18, StartMinute: 0, CaloriesPercent: 25},
		}, nil)

	s.mockNutr.EXPECT().
		Upsert(gomock.Any(), gomock.Any()).
		Return(nil)

	s.mockRecipe.EXPECT().
		FindByMealTypeExcludeIDs(gomock.Any(), gomock.Any()).
		Return([]model.Recipe{}, nil).AnyTimes()

	s.mockRecipe.EXPECT().
		FindByMealType(gomock.Any()).
		Return([]model.Recipe{{Title: "Default", MealType: model.MealBreakfast, Calories: 200, ProteinG: 10, FatG: 5, CarbsG: 30}}, nil).AnyTimes()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/nutrition/today", nil)
	req.Header.Set("Authorization", authHeader(t, s.jwtManager, 1))
	s.r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "final", resp["status"])
	// BMR for 70kg/175cm/25y/female = 1507.75, TDEE = 1809.3, + 300 from effectiveLoad (today, weight 1.0) = 2109.3
	assert.InDelta(t, 2109.3, resp["calories_target"].(float64), 1)
	assert.NotEmpty(t, resp["meals"])
}

func TestGetNutrition_WithMeal(t *testing.T) {
	s := setupAPI(t)
	defer s.ctrl.Finish()

	s.mockMealPeriod.EXPECT().
		FindByUserID(uint(1)).
		Return([]model.MealPeriod{
			{MealType: model.MealLunch, Name: "Обед", StartHour: 12, StartMinute: 0},
		}, nil)

	s.mockRecipe.EXPECT().
		FindByMealTypeExcludeIDs("lunch", gomock.Any()).
		Return([]model.Recipe{
			{Title: "Chicken Salad", MealType: model.MealLunch, Calories: 500, ProteinG: 35, FatG: 15, CarbsG: 30},
		}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/nutrition/today?meal=lunch", nil)
	req.Header.Set("Authorization", authHeader(t, s.jwtManager, 1))
	s.r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	lunch := resp["lunch"].(map[string]interface{})
	dishes := lunch["dishes"].([]interface{})
	require.Len(t, dishes, 1)
	firstDish := dishes[0].(map[string]interface{})
	assert.Equal(t, "Chicken Salad", firstDish["title"])
}

func TestGetNutrition_InvalidMeal(t *testing.T) {
	s := setupAPI(t)
	defer s.ctrl.Finish()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/nutrition/today?meal=snack", nil)
	req.Header.Set("Authorization", authHeader(t, s.jwtManager, 1))
	s.r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

// ───────────────────── RECIPES ─────────────────────

func TestListRecipesByIDs_Success(t *testing.T) {
	s := setupAPI(t)
	defer s.ctrl.Finish()

	s.mockRecipe.EXPECT().
		FindByIDs([]uint{1, 2}).
		Return([]model.Recipe{
			{ID: 1, Title: "Oatmeal", MealType: model.MealBreakfast, Calories: 320, ProteinG: 12, FatG: 8, CarbsG: 52},
			{ID: 2, Title: "Pasta", MealType: model.MealLunch, Calories: 520, ProteinG: 28, FatG: 22, CarbsG: 52},
		}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/recipes?ids=1,2", nil)
	req.Header.Set("Authorization", authHeader(t, s.jwtManager, 1))
	s.r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	require.Len(t, resp, 2)
	assert.Equal(t, "Oatmeal", resp[0]["title"])
	assert.Equal(t, "Pasta", resp[1]["title"])
}

func TestListRecipesByIDs_MissingIDs(t *testing.T) {
	s := setupAPI(t)
	defer s.ctrl.Finish()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/recipes", nil)
	req.Header.Set("Authorization", authHeader(t, s.jwtManager, 1))
	s.r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestListRecipesByIDs_InvalidID(t *testing.T) {
	s := setupAPI(t)
	defer s.ctrl.Finish()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/recipes?ids=abc", nil)
	req.Header.Set("Authorization", authHeader(t, s.jwtManager, 1))
	s.r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestListRecipesByIDs_EmptyResult(t *testing.T) {
	s := setupAPI(t)
	defer s.ctrl.Finish()

	s.mockRecipe.EXPECT().
		FindByIDs([]uint{999}).
		Return(nil, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/recipes?ids=999", nil)
	req.Header.Set("Authorization", authHeader(t, s.jwtManager, 1))
	s.r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp []interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Len(t, resp, 0)
}

// ───────────────────── PROTECTED ENDPOINTS ─────────────────────

func TestProtectedEndpoints_RejectWithoutAuth(t *testing.T) {
	s := setupAPI(t)
	defer s.ctrl.Finish()

	endpoints := []struct {
		method string
		path   string
		body   string
	}{
		{"GET", "/api/v1/user/profile", ""},
		{"PUT", "/api/v1/user/profile", `{"name":"Test"}`},
		{"GET", "/api/v1/activities", ""},
		{"POST", "/api/v1/activities", `{"type":"run"}`},
		{"DELETE", "/api/v1/activities/1", ""},
		{"GET", "/api/v1/nutrition/today", ""},
		{"GET", "/api/v1/recipes?ids=1,2", ""},
	}

	for _, ep := range endpoints {
		t.Run(fmt.Sprintf("%s %s", ep.method, ep.path), func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(ep.method, ep.path, bytes.NewBufferString(ep.body))
			if ep.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}
			s.r.ServeHTTP(w, req)
			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}
}

var _ = testutil.PtrInt
