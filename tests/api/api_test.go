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
	"github.com/ikukhar/refuel-backend/internal/handler"
	adminHandler "github.com/ikukhar/refuel-backend/internal/handler/admin"
	"github.com/ikukhar/refuel-backend/internal/model"
	"github.com/ikukhar/refuel-backend/internal/router"
	"github.com/ikukhar/refuel-backend/internal/service"
	mockrepo "github.com/ikukhar/refuel-backend/internal/service/mocks"
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
	mockNutr    *mockrepo.MockNutritionRepository
	mockRecipe  *mockrepo.MockRecipeRepository
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
	mockNutr := mockrepo.NewMockNutritionRepository(ctrl)
	mockRecipe := mockrepo.NewMockRecipeRepository(ctrl)

	userSvc := service.NewUserService(mockUser)
	authSvc := service.NewAuthService(mockUser, jwtManager, logger)
	activitySvc := service.NewActivityService(mockAct)
	nutritionSvc := service.NewNutritionService(mockNutr, mockAct, mockUser, mockRecipe)

	authH := handler.NewAuthHandler(authSvc)
	userH := handler.NewUserHandler(userSvc)
	activityH := handler.NewActivityHandler(activitySvc)
	nutritionH := handler.NewNutritionHandler(nutritionSvc)
	recipeAdminH := adminHandler.NewRecipeAdminHandler(nil)
	userMealPeriodsAdminH := adminHandler.NewUserMealPeriodAdminHandler(nil)

	r := router.Setup(logger, jwtManager, authH, userH, activityH, nutritionH, recipeAdminH, userMealPeriodsAdminH)

	return &apiSuite{
		r:          r,
		jwtManager: jwtManager,
		ctrl:       ctrl,
		mockUser:   mockUser,
		mockAct:    mockAct,
		mockNutr:   mockNutr,
		mockRecipe: mockRecipe,
	}
}

func authHeader(t *testing.T, jm *jwt.Manager, userID uint) string {
	t.Helper()
	token, err := jm.GenerateAccessToken(userID, "test@test.com")
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

	refresh, err := s.jwtManager.GenerateRefreshToken(3, "test@test.com")
	require.NoError(t, err)

	s.mockUser.EXPECT().
		FindByID(uint(3)).
		Return(&model.User{ID: 3, Email: "test@test.com"}, nil)

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

// ───────────────────── NUTRITION ─────────────────────

func TestGetNutrition_FullDay(t *testing.T) {
	s := setupAPI(t)
	defer s.ctrl.Finish()

	now := time.Now().Truncate(24 * time.Hour)

	s.mockNutr.EXPECT().
		FindByUserAndDate(uint(1), now).
		Return(nil, gorm.ErrRecordNotFound)

	s.mockUser.EXPECT().
		FindByID(uint(1)).
		Return(&model.User{ID: 1, Weight: 70, Height: 175, Age: 25, Gender: "female"}, nil)

	s.mockAct.EXPECT().
		FindByUserID(uint(1), &now, nil, 50, 0).
		Return([]model.Activity{
			{ID: 1, Calories: ptrInt(300)},
		}, nil)

	s.mockNutr.EXPECT().
		Upsert(gomock.Any()).
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
	// BMR for 70kg/175cm/25y/female = 1507.75, TDEE = 1809.3, + 150 from activity = 1959.3
	assert.InDelta(t, 1959.3, resp["calories_target"].(float64), 1)
	assert.NotEmpty(t, resp["meals"])
}

func TestGetNutrition_WithMeal(t *testing.T) {
	s := setupAPI(t)
	defer s.ctrl.Finish()

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
		{"GET", "/api/v1/nutrition/today", ""},
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

func ptrInt(v int) *int { return &v }
