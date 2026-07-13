package model

type MealType string

const (
	MealBreakfast MealType = "breakfast"
	MealLunch     MealType = "lunch"
	MealDinner    MealType = "dinner"
	MealSnack     MealType = "snack"
)

func (m MealType) Name() string {
	switch m {
	case MealBreakfast:
		return "Завтрак"
	case MealLunch:
		return "Обед"
	case MealDinner:
		return "Ужин"
	case MealSnack:
		return "Перекус"
	}
	return string(m)
}

func ValidMealTypes() []MealType {
	return []MealType{MealBreakfast, MealLunch, MealDinner, MealSnack}
}
