package admin

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ikukhar/refuel-backend/internal/model"
	"github.com/ikukhar/refuel-backend/internal/repository"
)

type UserMealPeriodAdminHandler struct {
	repo *repository.UserMealPeriodRepository
}

func NewUserMealPeriodAdminHandler(repo *repository.UserMealPeriodRepository) *UserMealPeriodAdminHandler {
	return &UserMealPeriodAdminHandler{repo: repo}
}

func (h *UserMealPeriodAdminHandler) List(c *gin.Context) {
	periods, err := h.repo.FindByUserID(0)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.HTML(http.StatusOK, "user_meal_periods_list.html", gin.H{
		"Periods":      periods,
		"DefaultMeals": model.DefaultMealPeriods,
	})
}

func (h *UserMealPeriodAdminHandler) NewForm(c *gin.Context) {
	c.HTML(http.StatusOK, "user_meal_periods_form.html", gin.H{
		"Period":       nil,
		"DefaultMeals": model.DefaultMealPeriods,
	})
}

func (h *UserMealPeriodAdminHandler) Create(c *gin.Context) {
	p, errMsg := parseUserMealPeriodForm(c)
	if errMsg != "" {
		c.HTML(http.StatusUnprocessableEntity, "user_meal_periods_form.html", gin.H{
			"Period":       nil,
			"DefaultMeals": model.DefaultMealPeriods,
			"Error":        errMsg,
		})
		return
	}
	if err := h.repo.Create(p); err != nil {
		c.HTML(http.StatusUnprocessableEntity, "user_meal_periods_form.html", gin.H{
			"Period":       nil,
			"DefaultMeals": model.DefaultMealPeriods,
			"Error":        err.Error(),
		})
		return
	}
	c.Redirect(http.StatusFound, "/admin/user-meal-periods")
}

func (h *UserMealPeriodAdminHandler) EditForm(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	p, err := h.repo.FindByID(uint(id))
	if err != nil {
		c.String(http.StatusNotFound, "Период не найден")
		return
	}
	c.HTML(http.StatusOK, "user_meal_periods_form.html", gin.H{
		"Period":       p,
		"DefaultMeals": model.DefaultMealPeriods,
	})
}

func (h *UserMealPeriodAdminHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	existing, err := h.repo.FindByID(uint(id))
	if err != nil {
		c.String(http.StatusNotFound, "Период не найден")
		return
	}

	p, errMsg := parseUserMealPeriodForm(c)
	if errMsg != "" {
		c.HTML(http.StatusUnprocessableEntity, "user_meal_periods_form.html", gin.H{
			"Period":       existing,
			"DefaultMeals": model.DefaultMealPeriods,
			"Error":        errMsg,
		})
		return
	}
	p.ID = existing.ID
	p.CreatedAt = existing.CreatedAt

	if err := h.repo.Update(p); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.Redirect(http.StatusFound, "/admin/user-meal-periods")
}

func (h *UserMealPeriodAdminHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.repo.Delete(uint(id)); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.Status(http.StatusOK)
}

func parseUserMealPeriodForm(c *gin.Context) (*model.UserMealPeriod, string) {
	userIDStr := c.DefaultPostForm("user_id", "0")
	userID, _ := strconv.ParseUint(userIDStr, 10, 64)

	mealType := c.PostForm("meal_type")
	if mealType == "" {
		return nil, "Тип приёма обязателен"
	}

	startHour, _ := strconv.Atoi(c.PostForm("start_hour"))
	startMinute, _ := strconv.Atoi(c.DefaultPostForm("start_minute", "0"))

	return &model.UserMealPeriod{
		UserID:      uint(userID),
		MealType:    model.MealType(mealType),
		StartHour:   startHour,
		StartMinute: startMinute,
	}, ""
}
