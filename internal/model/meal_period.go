package model

func MealTypeName(mt MealType) string {
	for _, p := range DefaultMealPeriods {
		if p.MealType == mt {
			return p.Name
		}
	}
	return string(mt)
}

type MealPeriod struct {
	MealType    MealType
	Name        string
	StartHour   int
	StartMinute int
	EndHour     int
	EndMinute   int
	SortOrder   int
}

var DefaultMealPeriods = []MealPeriod{
	{MealType: MealBreakfast, Name: "Завтрак", StartHour: 7, StartMinute: 0, EndHour: 9, EndMinute: 0, SortOrder: 0},
	{MealType: MealSnack, Name: "1-й перекус", StartHour: 10, StartMinute: 0, EndHour: 11, EndMinute: 0, SortOrder: 1},
	{MealType: MealLunch, Name: "Обед", StartHour: 12, StartMinute: 0, EndHour: 14, EndMinute: 0, SortOrder: 2},
	{MealType: MealSnack, Name: "2-й перекус", StartHour: 10, StartMinute: 0, EndHour: 16, EndMinute: 0, SortOrder: 3},
	{MealType: MealDinner, Name: "Ужин", StartHour: 17, StartMinute: 0, EndHour: 20, EndMinute: 0, SortOrder: 4},
}
