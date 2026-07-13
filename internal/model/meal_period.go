package model

type MealPeriod struct {
	MealType    MealType
	Name        string
	StartHour   int
	StartMinute int
	SortOrder   int
}

var DefaultMealPeriods = []MealPeriod{
	{MealType: MealBreakfast, Name: "Завтрак", StartHour: 7, StartMinute: 0, SortOrder: 0},
	{MealType: MealSnack, Name: "1-й перекус", StartHour: 10, StartMinute: 0, SortOrder: 1},
	{MealType: MealLunch, Name: "Обед", StartHour: 12, StartMinute: 0, SortOrder: 2},
	{MealType: MealSnack, Name: "2-й перекус", StartHour: 10, StartMinute: 0, SortOrder: 3},
	{MealType: MealDinner, Name: "Ужин", StartHour: 17, StartMinute: 0, SortOrder: 4},
}
