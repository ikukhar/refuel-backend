package model

type MealPeriod struct {
	MealType        MealType
	Name            string
	StartHour       int
	StartMinute     int
	SortOrder       int
	CaloriesPercent float64
}

var DefaultMealPeriods = []MealPeriod{
	{MealType: MealBreakfast, Name: "Завтрак", StartHour: 7, StartMinute: 0, SortOrder: 0, CaloriesPercent: 25},
	{MealType: MealSnack, Name: "Перекус 1", StartHour: 10, StartMinute: 0, SortOrder: 1, CaloriesPercent: 10},
	{MealType: MealLunch, Name: "Обед", StartHour: 12, StartMinute: 0, SortOrder: 2, CaloriesPercent: 35},
	{MealType: MealSnack, Name: "Перекус 2", StartHour: 15, StartMinute: 0, SortOrder: 3, CaloriesPercent: 10},
	{MealType: MealDinner, Name: "Ужин", StartHour: 17, StartMinute: 0, SortOrder: 4, CaloriesPercent: 20},
}

func DefaultCaloriesPercent(mealType MealType) float64 {
	for _, p := range DefaultMealPeriods {
		if p.MealType == mealType {
			return p.CaloriesPercent
		}
	}
	return 0
}
