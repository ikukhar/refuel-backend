package model

type MealType string

const (
	MealBreakfast MealType = "breakfast"
	MealLunch     MealType = "lunch"
	MealDinner    MealType = "dinner"
	MealSnack     MealType = "snack"
)

func ValidMealTypes() []MealType {
	return []MealType{MealBreakfast, MealLunch, MealDinner, MealSnack}
}
