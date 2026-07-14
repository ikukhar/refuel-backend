package admin

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ikukhar/refuel-backend/internal/model"
	"github.com/ikukhar/refuel-backend/internal/repository"
)

type MealPeriodAdminHandler struct {
	repo *repository.MealPeriodRepository
}

func NewMealPeriodAdminHandler(repo *repository.MealPeriodRepository) *MealPeriodAdminHandler {
	return &MealPeriodAdminHandler{repo: repo}
}

func (h *MealPeriodAdminHandler) List(c *gin.Context) {
	periods, err := h.repo.FindAll()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.HTML(http.StatusOK, "meal_periods_list.html", gin.H{
		"Periods":      periods,
		"DefaultMeals": model.DefaultMealPeriods,
	})
}

func (h *MealPeriodAdminHandler) NewForm(c *gin.Context) {
	c.HTML(http.StatusOK, "meal_periods_form.html", gin.H{
		"Period":       nil,
		"DefaultMeals": model.DefaultMealPeriods,
	})
}

func (h *MealPeriodAdminHandler) Create(c *gin.Context) {
	p, errMsg := parseMealPeriodForm(c)
	if errMsg != "" {
		c.HTML(http.StatusUnprocessableEntity, "meal_periods_form.html", gin.H{
			"Period":       nil,
			"DefaultMeals": model.DefaultMealPeriods,
			"Error":        errMsg,
		})
		return
	}
	if err := h.repo.Create(p); err != nil {
		c.HTML(http.StatusUnprocessableEntity, "meal_periods_form.html", gin.H{
			"Period":       nil,
			"DefaultMeals": model.DefaultMealPeriods,
			"Error":        err.Error(),
		})
		return
	}
	c.Redirect(http.StatusFound, "/admin/user-meal-periods")
}

func (h *MealPeriodAdminHandler) EditForm(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	p, err := h.repo.FindByID(uint(id))
	if err != nil {
		c.String(http.StatusNotFound, "Период не найден")
		return
	}
	c.HTML(http.StatusOK, "meal_periods_form.html", gin.H{
		"Period":       p,
		"DefaultMeals": model.DefaultMealPeriods,
	})
}

func (h *MealPeriodAdminHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	existing, err := h.repo.FindByID(uint(id))
	if err != nil {
		c.String(http.StatusNotFound, "Период не найден")
		return
	}

	p, errMsg := parseMealPeriodForm(c)
	if errMsg != "" {
		c.HTML(http.StatusUnprocessableEntity, "meal_periods_form.html", gin.H{
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

func (h *MealPeriodAdminHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.repo.Delete(uint(id)); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.Status(http.StatusOK)
}

func parseMealPeriodForm(c *gin.Context) (*model.MealPeriod, string) {
	userIDStr := c.DefaultPostForm("user_id", "0")
	userID, _ := strconv.ParseUint(userIDStr, 10, 64)

	mealType := c.PostForm("meal_type")
	if mealType == "" {
		return nil, "Тип приёма обязателен"
	}

	startHour, _ := strconv.Atoi(c.PostForm("start_hour"))
	startMinute, _ := strconv.Atoi(c.DefaultPostForm("start_minute", "0"))

	name := c.PostForm("name")

	caloriesPercent, _ := strconv.ParseFloat(c.DefaultPostForm("calories_percent", "0"), 64)

	return &model.MealPeriod{
		UserID:          uint(userID),
		MealType:        model.MealType(mealType),
		Name:            name,
		StartHour:       startHour,
		StartMinute:     startMinute,
		CaloriesPercent: caloriesPercent,
	}, ""
}
