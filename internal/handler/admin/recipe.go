package admin

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ikukhar/refuel-backend/internal/model"
	"github.com/ikukhar/refuel-backend/internal/repository"
)

type RecipeAdminHandler struct {
	recipeRepo *repository.RecipeRepository
}

func NewRecipeAdminHandler(recipeRepo *repository.RecipeRepository) *RecipeAdminHandler {
	return &RecipeAdminHandler{recipeRepo: recipeRepo}
}

type recipeView struct {
	model.Recipe
	IngredientsLines []string
	StepsLines       []string
}

func toView(r *model.Recipe) recipeView {
	return recipeView{
		Recipe:           *r,
		IngredientsLines: parseJSONList(r.Ingredients),
		StepsLines:       parseJSONList(r.Steps),
	}
}

func toViews(recipes []model.Recipe) []recipeView {
	views := make([]recipeView, len(recipes))
	for i, r := range recipes {
		views[i] = toView(&r)
	}
	return views
}

func parseJSONList(s string) []string {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, "[]")
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	lines := make([]string, 0, len(parts))
	for _, p := range parts {
		l := strings.Trim(strings.TrimSpace(p), "\"")
		if l != "" {
			lines = append(lines, l)
		}
	}
	return lines
}

func toJSONList(lines []string) string {
	escaped := make([]string, len(lines))
	for i, l := range lines {
		escaped[i] = `"` + l + `"`
	}
	return "[" + strings.Join(escaped, ",") + "]"
}

func (h *RecipeAdminHandler) List(c *gin.Context) {
	recipes, err := h.recipeRepo.FindAll()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.HTML(http.StatusOK, "list.html", gin.H{
		"Recipes": toViews(recipes),
	})
}

func (h *RecipeAdminHandler) NewForm(c *gin.Context) {
	c.HTML(http.StatusOK, "form.html", gin.H{
		"Recipe": nil,
	})
}

func (h *RecipeAdminHandler) Create(c *gin.Context) {
	recipe, errMsg := parseRecipeForm(c)
	if errMsg != "" {
		c.HTML(http.StatusUnprocessableEntity, "form.html", gin.H{
			"Recipe": nil,
			"Error":  errMsg,
		})
		return
	}

	if err := h.recipeRepo.Create(recipe); err != nil {
		c.HTML(http.StatusUnprocessableEntity, "form.html", gin.H{
			"Recipe": nil,
			"Error":  err.Error(),
		})
		return
	}

	c.Redirect(http.StatusFound, "/admin/recipes")
}

func (h *RecipeAdminHandler) EditForm(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	recipe, err := h.recipeRepo.FindByID(uint(id))
	if err != nil {
		c.String(http.StatusNotFound, "Рецепт не найден")
		return
	}

	c.HTML(http.StatusOK, "form.html", gin.H{
		"Recipe": toView(recipe),
	})
}

func (h *RecipeAdminHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	existing, err := h.recipeRepo.FindByID(uint(id))
	if err != nil {
		c.String(http.StatusNotFound, "Рецепт не найден")
		return
	}

	recipe, errMsg := parseRecipeForm(c)
	if errMsg != "" {
		c.HTML(http.StatusUnprocessableEntity, "form.html", gin.H{
			"Recipe": toView(existing),
			"Error":  errMsg,
		})
		return
	}

	recipe.ID = existing.ID
	recipe.CreatedAt = existing.CreatedAt

	if err := h.recipeRepo.Update(recipe); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.Redirect(http.StatusFound, "/admin/recipes")
}

func (h *RecipeAdminHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.recipeRepo.Delete(uint(id)); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.Status(http.StatusOK)
}

func getFormLines(c *gin.Context, key string) []string {
	val := strings.TrimSpace(c.PostForm(key))
	if val == "" {
		return nil
	}
	lines := strings.Split(val, "\n")
	result := make([]string, 0, len(lines))
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l != "" {
			result = append(result, l)
		}
	}
	return result
}

func parseRecipeForm(c *gin.Context) (*model.Recipe, string) {
	title := strings.TrimSpace(c.PostForm("title"))
	if title == "" {
		return nil, "Название обязательно"
	}

	calories, err := strconv.Atoi(c.PostForm("calories"))
	if err != nil || calories <= 0 {
		return nil, "Калории обязательны и должны быть > 0"
	}

	protein, err := strconv.ParseFloat(c.PostForm("protein_g"), 64)
	if err != nil || protein < 0 {
		return nil, "Белки обязательны"
	}

	fat, err := strconv.ParseFloat(c.PostForm("fat_g"), 64)
	if err != nil || fat < 0 {
		return nil, "Жиры обязательны"
	}

	carbs, err := strconv.ParseFloat(c.PostForm("carbs_g"), 64)
	if err != nil || carbs < 0 {
		return nil, "Углеводы обязательны"
	}

	imageURL := c.PostForm("image_url")
	var imageURLPtr *string
	if imageURL != "" {
		imageURLPtr = &imageURL
	}

	return &model.Recipe{
		Title:       title,
		Description: c.PostForm("description"),
		Calories:    calories,
		ProteinG:    protein,
		FatG:        fat,
		CarbsG:      carbs,
		ImageURL:    imageURLPtr,
		MealType:    c.PostForm("meal_type"),
		Steps:       toJSONList(getFormLines(c, "steps")),
		Ingredients: toJSONList(getFormLines(c, "ingredients")),
	}, ""
}
