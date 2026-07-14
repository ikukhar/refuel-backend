package repository

import (
	"github.com/ikukhar/refuel-backend/internal/model"
	"gorm.io/gorm"
)

type RecipeRepository struct {
	db *gorm.DB
}

func NewRecipeRepository(db *gorm.DB) *RecipeRepository {
	return &RecipeRepository{db: db}
}

func (r *RecipeRepository) Create(recipe *model.Recipe) error {
	return r.db.Create(recipe).Error
}

func (r *RecipeRepository) FindByID(id uint) (*model.Recipe, error) {
	var recipe model.Recipe
	err := r.db.First(&recipe, id).Error
	if err != nil {
		return nil, err
	}
	return &recipe, nil
}

func (r *RecipeRepository) FindAll() ([]model.Recipe, error) {
	var recipes []model.Recipe
	err := r.db.Order("created_at DESC").Find(&recipes).Error
	return recipes, err
}

func (r *RecipeRepository) FindByMealType(mealType string) ([]model.Recipe, error) {
	var recipes []model.Recipe
	err := r.db.Where("meal_type = ?", mealType).Order("created_at DESC").Find(&recipes).Error
	return recipes, err
}

func (r *RecipeRepository) FindByMealTypeExcludeIDs(mealType string, excludeIDs []uint) ([]model.Recipe, error) {
	var recipes []model.Recipe
	query := r.db.Where("meal_type = ?", mealType)
	if len(excludeIDs) > 0 {
		query = query.Where("id NOT IN ?", excludeIDs)
	}
	err := query.Order("created_at DESC").Find(&recipes).Error
	return recipes, err
}

func (r *RecipeRepository) Update(recipe *model.Recipe) error {
	return r.db.Save(recipe).Error
}

func (r *RecipeRepository) Delete(id uint) error {
	return r.db.Delete(&model.Recipe{}, id).Error
}

func (r *RecipeRepository) SeedRecipes() error {
	var count int64
	r.db.Model(&model.Recipe{}).Count(&count)
	if count > 0 {
		return nil
	}

	recipes := []model.Recipe{
		// ── ЗАВТРАК ──
		{Title: "Овсяная каша с ягодами", MealType: model.MealBreakfast, Calories: 320, ProteinG: 12, FatG: 8, CarbsG: 52,
		Servings: 1, Description: "Геркулес на молоке с сезонными ягодами и мёдом",
			Ingredients: `["овсяные хлопья 60г","молоко 200мл","ягоды 80г","мёд 1ст.л."]`, Steps: `["Залить хлопья молоком","Варить 5 мин","Добавить ягоды и мёд"]`},
		{Title: "Сырники со сметаной", MealType: model.MealBreakfast, Calories: 380, ProteinG: 24, FatG: 16, CarbsG: 34,
		Servings: 1, Description: "Нежные творожные сырники",
			Ingredients: `["творог 200г","яйцо 1шт","мука 2ст.л.","сахар 1ст.л.","сметана 30г"]`, Steps: `["Смешать творог с яйцом и мукой","Сформировать сырники","Обжарить до золотистой корочки"]`},
		{Title: "Яичница с авокадо", MealType: model.MealBreakfast, Calories: 350, ProteinG: 20, FatG: 24, CarbsG: 10,
		Servings: 1, Description: "Глазунья на тосте с авокадо",
			Ingredients: `["яйца 2шт","авокадо 1/2шт","хлеб цельнозерновой 1ломтик","соль, перец"]`, Steps: `["Поджарить хлеб","Размять авокадо на тост","Сверху выложить яичницу"]`},
		{Title: "Греческий йогурт с гранолой", MealType: model.MealBreakfast, Calories: 290, ProteinG: 18, FatG: 10, CarbsG: 36,
		Servings: 1, Description: "Густой йогурт с хрустящей гранолой и бананом",
			Ingredients: `["греческий йогурт 200г","гранола 40г","банан 1шт","мёд 1ч.л."]`, Steps: `["Выложить йогурт в миску","Посыпать гранолой","Добавить нарезанный банан"]`},
		{Title: "Блинчики с творогом", MealType: model.MealBreakfast, Calories: 410, ProteinG: 22, FatG: 14, CarbsG: 48,
		Servings: 1, Description: "Тонкие блины с творожной начинкой",
			Ingredients: `["мука 100г","молоко 250мл","яйцо 1шт","творог 150г","сахар 1ст.л."]`, Steps: `["Замесить тесто и испечь блины","Смешать творог с сахаром","Завернуть начинку в блины"]`},
		{Title: "Каша рисовая с тыквой", MealType: model.MealBreakfast, Calories: 280, ProteinG: 8, FatG: 6, CarbsG: 50,
		Servings: 1, Description: "Сладкая рисовая каша с запечённой тыквой",
			Ingredients: `["рис круглозёрный 60г","тыква 100г","молоко 200мл","сливочное масло 10г","сахар 1ст.л."]`, Steps: `["Тыкву запечь 20 мин","Сварить рисовую кашу","Смешать с тыквой и маслом"]`},
		{Title: "Тост с арахисовой пастой", MealType: model.MealBreakfast, Calories: 340, ProteinG: 14, FatG: 18, CarbsG: 32,
		Servings: 1, Description: "Хрустящий тост с пастой и бананом",
			Ingredients: `["хлеб цельнозерновой 2ломтика","арахисовая паста 30г","банан 1шт","корица"]`, Steps: `["Поджарить тосты","Намазать пастой","Сверху выложить кружочки банана и корицу"]`},
		{Title: "Омлет с овощами", MealType: model.MealBreakfast, Calories: 310, ProteinG: 22, FatG: 18, CarbsG: 12,
		Servings: 1, Description: "Пышный омлет с помидорами и шпинатом",
			Ingredients: `["яйца 3шт","молоко 50мл","помидор 1шт","шпинат 30г","масло оливковое 1ст.л."]`, Steps: `["Взбить яйца с молоком","Нарезать овощи","Залить смесь и запечь 10 мин"]`},
		{Title: "Пудинг чиа с манго", MealType: model.MealBreakfast, Calories: 270, ProteinG: 10, FatG: 12, CarbsG: 32,
		Servings: 1, Description: "Холодный завтрак из семян чиа на кокосовом молоке",
			Ingredients: `["семена чиа 3ст.л.","кокосовое молоко 200мл","манго 100г","мёд 1ч.л."]`, Steps: `["Смешать чиа с молоком","Убрать в холодильник на ночь","Утром добавить манго"]`},
		{Title: "Драники со сметаной", MealType: model.MealBreakfast, Calories: 360, ProteinG: 10, FatG: 16, CarbsG: 44,
		Servings: 1, Description: "Хрустящие картофельные оладьи",
			Ingredients: `["картофель 3шт","лук 1/2шт","яйцо 1шт","мука 2ст.л.","сметана 30г"]`, Steps: `["Натереть картофель","Отжать и смешать с луком и яйцом","Обжарить до румяной корочки"]`},
		{Title: "Запеканка творожная", MealType: model.MealBreakfast, Calories: 340, ProteinG: 26, FatG: 12, CarbsG: 32,
		Servings: 1, Description: "Классическая творожная запеканка с изюмом",
			Ingredients: `["творог 300г","яйца 2шт","манка 2ст.л.","изюм 40г","сахар 2ст.л."]`, Steps: `["Смешать все ингредиенты","Выложить в форму","Запекать 30 мин при 180°C"]`},
		{Title: "Смузи-боул с киноа", MealType: model.MealBreakfast, Calories: 310, ProteinG: 14, FatG: 9, CarbsG: 46,
		Servings: 1, Description: "Густой смузи с киноа и topping'ами",
			Ingredients: `["банан 1шт","шпинат 50г","молоко 200мл","киноа варёная 60г","гранола 20г"]`, Steps: `["Смешать банан, шпинат и молоко","Перелить в миску","Добавить киноа и гранолу"]`},

		// ── ОБЕД ──
		{Title: "Куриный суп с лапшой", MealType: model.MealLunch, Calories: 280, ProteinG: 22, FatG: 8, CarbsG: 30,
		Servings: 1, Description: "Наваристый куриный суп с домашней лапшой",
			Ingredients: `["куриное филе 200г","морковь 1шт","лук 1шт","лапша 80г","картофель 2шт"]`, Steps: `["Сварить бульон","Добавить нарезанные овощи","За 5 мин до готовности добавить лапшу"]`},
		{Title: "Паста карбонара", MealType: model.MealLunch, Calories: 520, ProteinG: 28, FatG: 22, CarbsG: 52,
		Servings: 1, Description: "Спагетти с беконом и сливочным соусом",
			Ingredients: `["спагетти 200г","бекон 80г","сливки 100мл","желток 2шт","пармезан 30г"]`, Steps: `["Сварить пасту","Обжарить бекон","Смешать с желтками и сливками"]`},
		{Title: "Борщ с говядиной", MealType: model.MealLunch, Calories: 320, ProteinG: 20, FatG: 12, CarbsG: 34,
		Servings: 1, Description: "Красный борщ с мясом и сметаной",
			Ingredients: `["говядина 250г","свёкла 1шт","капуста 200г","картофель 2шт","сметана 30г"]`, Steps: `["Сварить мясной бульон","Добавить нашинкованные овощи","Подавать со сметаной"]`},
		{Title: "Рис с курицей и овощами", MealType: model.MealLunch, Calories: 420, ProteinG: 32, FatG: 10, CarbsG: 52,
		Servings: 1, Description: "Рассольник с перловой крупой",
			Ingredients: `["куриное филе 200г","рис 100г","морковь 1шт","перец болгарский 1шт","соус соевый 2ст.л."]`, Steps: `["Обжарить курицу","Добавить овощи и рис","Тушить под крышкой 20 мин"]`},
		{Title: "Стейк из лосося с пюре", MealType: model.MealLunch, Calories: 480, ProteinG: 36, FatG: 22, CarbsG: 32,
		Servings: 1, Description: "Сочный лосось с нежным картофельным пюре",
			Ingredients: `["лосось 200г","картофель 300г","сливочное масло 20г","лимон 1/2шт","укроп"]`, Steps: `["Обжарить стейк лосося 4 мин с каждой стороны","Сварить картофель и сделать пюре","Подавать с долькой лимона"]`},
		{Title: "Гречка с тушёнкой", MealType: model.MealLunch, Calories: 450, ProteinG: 28, FatG: 18, CarbsG: 46,
		Servings: 1, Description: "Гречневая каша с мясной тушёнкой",
			Ingredients: `["гречка 150г","тушёнка говяжья 200г","лук 1шт","морковь 1шт","чеснок 2зубчика"]`, Steps: `["Обжарить лук и морковь","Добавить гречку и воду","Варить 15 мин, добавить тушёнку"]`},
		{Title: "Цезарь с курицей", MealType: model.MealLunch, Calories: 380, ProteinG: 32, FatG: 18, CarbsG: 18,
		Servings: 1, Description: "Классический салат Цезарь с куриной грудкой",
			Ingredients: `["куриное филе 150г","салат айсберг 200г","сухарики 40г","пармезан 30г","соус цезарь 30мл"]`, Steps: `["Обжарить курицу","Нарвать салат","Смешать с сухариками соусом и сыром"]`},
		{Title: "Том-ям с креветками", MealType: model.MealLunch, Calories: 260, ProteinG: 18, FatG: 14, CarbsG: 16,
		Servings: 1, Description: "Острый тайский суп на кокосовом молоке",
			Ingredients: `["креветки 200г","кокосовое молоко 400мл","паста том-ям 2ст.л.","шампиньоны 100г","рис 60г"]`, Steps: `["Вскипятить кокосовое молоко","Добавить пасту и грибы","За 3 мин до готовности добавить креветки"]`},
		{Title: "Лазанья болоньезе", MealType: model.MealLunch, Calories: 550, ProteinG: 30, FatG: 24, CarbsG: 50,
		Servings: 1, Description: "Многослойная паста с мясным соусом и бешамелью",
			Ingredients: `["листы лазаньи 6шт","фарш говяжий 300г","моцарелла 150г","томатная паста 100г","мука 30г","молоко 500мл"]`, Steps: `["Приготовить соус болоньезе","Приготовить соус бешамель","Собрать слои и запечь 40 мин"]`},
		{Title: "Котлеты по-киевски с пюре", MealType: model.MealLunch, Calories: 510, ProteinG: 30, FatG: 26, CarbsG: 38,
		Servings: 1, Description: "Куриная котлета с маслом внутри",
			Ingredients: `["куриное филе 200г","сливочное масло 30г","панировочные сухари 40г","яйцо 1шт","картофель 300г"]`, Steps: `["Отбить филе, завернуть масло","Панировать и обжарить","Сделать пюре на гарнир"]`},
		{Title: "Куриная грудка с киноа", MealType: model.MealLunch, Calories: 400, ProteinG: 38, FatG: 8, CarbsG: 42,
		Servings: 1, Description: "Запечённая курица с киноа и овощами гриль",
			Ingredients: `["куриное-филе 200г","киноа 80г","цукини 100г","перец болгарский 1шт","оливковое масло 1ст.л."]`, Steps: `["Приправить курицу и запечь 25 мин","Сварить киноа","Овощи нарезать и обжарить на гриле"]`},
		{Title: "Уха рыбацкая", MealType: model.MealLunch, Calories: 190, ProteinG: 24, FatG: 4, CarbsG: 16,
		Servings: 1, Description: "Ароматный суп из речной рыбы",
			Ingredients: `["судак 300г","картофель 2шт","морковь 1шт","лук 1шт","лавровый лист"]`, Steps: `["Сварить рыбный бульон","Процедить, добавить овощи","Варить 20 мин, подавать с зеленью"]`},

		// ── УЖИН ──
		{Title: "Свинина с овощами в духовке", MealType: model.MealDinner, Calories: 420, ProteinG: 30, FatG: 20, CarbsG: 28,
		Servings: 1, Description: "Запечённая свинина с овощами",
			Ingredients: `["свинина 250г","картофель 2шт","морковь 1шт","лук 1шт","чеснок 3зубчика"]`, Steps: `["Нарезать мясо и овощи","Замариновать с чесноком","Запекать 40 мин при 200°C"]`},
		{Title: "Рыбное филе на пару", MealType: model.MealDinner, Calories: 220, ProteinG: 32, FatG: 8, CarbsG: 6,
		Servings: 1, Description: "Нежное филе трески на пару с брокколи",
			Ingredients: `["треска 250г","брокколи 200г","лимон 1/2шт","оливковое масло 1ст.л."]`, Steps: `["Филе приправить","Брокколи разобрать на соцветия","Готовить на пару 15 мин"]`},
		{Title: "Спагетти с морепродуктами", MealType: model.MealDinner, Calories: 380, ProteinG: 26, FatG: 12, CarbsG: 46,
		Servings: 1, Description: "Паста с креветками и мидиями в сливочном соусе",
			Ingredients: `["спагетти 180г","креветки 150г","мидии 100г","сливки 100мл","чеснок 2зубчика"]`, Steps: `["Сварить пасту","Обжарить морепродукты с чесноком","Добавить сливки и смешать с пастой"]`},
		{Title: "Индейка с овощами гриль", MealType: model.MealDinner, Calories: 350, ProteinG: 34, FatG: 12, CarbsG: 24,
		Servings: 1, Description: "Сочное филе индейки с гриль-овощами",
			Ingredients: `["филе индейки 250г","баклажан 1шт","перец 1шт","цукини 100г","оливковое масло 1ст.л."]`, Steps: `["Нарезать овощи","Обжарить индейку 5 мин","Добавить овощи и готовить ещё 10 мин"]`},
		{Title: "Овощное рагу с нутом", MealType: model.MealDinner, Calories: 310, ProteinG: 14, FatG: 10, CarbsG: 42,
		Servings: 1, Description: "Сытное рагу без мяса",
			Ingredients: `["нут консервированный 200г","баклажан 1шт","помидоры 2шт","перец 1шт","томатная паста 2ст.л."]`, Steps: `["Обжарить лук","Добавить нарезанные овощи","Тушить 20 мин, добавить нут и пасту"]`},
		{Title: "Куриные крылья BBQ", MealType: model.MealDinner, Calories: 450, ProteinG: 28, FatG: 24, CarbsG: 30,
		Servings: 1, Description: "Запечённые крылья в соусе барбекю",
			Ingredients: `["куриные крылья 400г","соус BBQ 50мл","чеснок 3зубчика","имбирь 10г","кунжут 1ч.л."]`, Steps: `["Замариновать крылья в соусе 1ч","Выложить на противень","Запекать 30 мин при 200°C"]`},
		{Title: "Салат с тунцом и яйцом", MealType: model.MealDinner, Calories: 280, ProteinG: 28, FatG: 14, CarbsG: 8,
		Servings: 1, Description: "Лёгкий салат с консервированным тунцом",
			Ingredients: `["тунец консервированный 180г","яйца 2шт","салат 200г","огурец 1шт","оливковое масло 1ст.л."]`, Steps: `["Отварить яйца","Размять тунец","Смешать с нарезанным огурцом и салатом"]`},
		{Title: "Тыквенный крем-суп", MealType: model.MealDinner, Calories: 240, ProteinG: 8, FatG: 14, CarbsG: 24,
		Servings: 1, Description: "Нежный суп-пюре из запечённой тыквы",
			Ingredients: `["тыква 400г","кокосовое молоко 200мл","имбирь 10г","чеснок 2зубчика","семечки тыквенные 20г"]`, Steps: `["Запечь тыкву 30 мин","Измельчить блендером","Добавить кокосовое молоко и прогреть"]`},
		{Title: "Стейк из говядины с рукколой", MealType: model.MealDinner, Calories: 480, ProteinG: 38, FatG: 28, CarbsG: 12,
		Servings: 1, Description: "Сочный стейк Рибай с салатом из рукколы",
			Ingredients: `["говядина рибай 250г","руккола 80г","пармезан 20г","оливковое масло 2ст.л.","бальзамический уксус"]`, Steps: `["Обжарить стейк 3 мин с каждой стороны","Дать отдохнуть 5 мин","Подавать на подушке из рукколы с пармезаном"]`},
		{Title: "Фаршированные перцы", MealType: model.MealDinner, Calories: 380, ProteinG: 24, FatG: 16, CarbsG: 32,
		Servings: 1, Description: "Болгарские перцы с мясной начинкой",
			Ingredients: `["перец болгарский 4шт","фарш говяжий 300г","рис 80г","лук 1шт","томатная паста 3ст.л.","сметана 50г"]`, Steps: `["Очистить перцы от семян","Смешать фарш с рисом и луком","Нафаршировать и тушить в соусе 40 мин"]`},
		{Title: "Плов узбекский", MealType: model.MealDinner, Calories: 520, ProteinG: 24, FatG: 18, CarbsG: 64,
		Servings: 1, Description: "Классический узбекский плов с бараниной",
			Ingredients: `["баранина 300г","рис девзира 200г","морковь 2шт","лук 2шт","чеснок 1головка","зира 1ч.л."]`, Steps: `["Обжарить мясо до корочки","Добавить морковь и лук","Залить водой и добавить рис","Томить 30 мин под крышкой"]`},
		{Title: "Форель с лимоном и розмарином", MealType: model.MealDinner, Calories: 340, ProteinG: 34, FatG: 20, CarbsG: 4,
		Servings: 1, Description: "Запечённая форель с травами",
			Ingredients: `["форель 300г","лимон 1шт","розмарин 2веточки","оливковое масло 1ст.л.","спаржа 100г"]`, Steps: `["Натереть рыбу солью и маслом","Внутрь положить дольки лимона и розмарин","Запекать 25 мин при 190°C, добавить спаржу за 5 мин"]`},

		// ── ПЕРЕКУС ──
		{Title: "Энергетические батончики", MealType: model.MealSnack, Calories: 200, ProteinG: 8, FatG: 10, CarbsG: 22,
		Servings: 1, Description: "Домашние батончики из овсянки и орехов",
			Ingredients: `["овсяные хлопья 100г","орехи 50г","мёд 3ст.л.","масло арахисовое 2ст.л.","сухофрукты 50г"]`, Steps: `["Смешать сухие ингредиенты","Добавить мёд и масло","Утрамбовать в форму и охладить 2ч"]`},
		{Title: "Гренки с сыром и чесноком", MealType: model.MealSnack, Calories: 250, ProteinG: 12, FatG: 14, CarbsG: 22,
		Servings: 1, Description: "Хрустящие гренки к супу или пиву",
			Ingredients: `["хлеб 200г","сыр твёрдый 80г","чеснок 2зубчика","масло сливочное 30г"]`, Steps: `["Натереть хлеб чесноком","Посыпать тёртым сыром","Запечь 7 мин при 180°C"]`},
		{Title: "Фруктовый салат", MealType: model.MealSnack, Calories: 150, ProteinG: 2, FatG: 4, CarbsG: 30,
		Servings: 1, Description: "Микс из сезонных фруктов",
			Ingredients: `["яблоко 1шт","груша 1шт","апельсин 1шт","гранат 50г","йогурт 50мл"]`, Steps: `["Нарезать фрукты кубиками","Смешать","Заправить йогуртом"]`},
		{Title: "Бутерброд с лососем", MealType: model.MealSnack, Calories: 280, ProteinG: 16, FatG: 14, CarbsG: 22,
		Servings: 1, Description: "Открытый сэндвич с лососем и сливочным сыром",
			Ingredients: `["хлеб ржаной 2ломтика","лосось слабосолёный 80г","сливочный сыр 30г","огурец 1/2шт","укроп"]`, Steps: `["Намазать сыр на хлеб","Выложить ломтики лосося","Добавить огурец и укроп"]`},
		{Title: "Орехи и сухофрукты", MealType: model.MealSnack, Calories: 180, ProteinG: 6, FatG: 12, CarbsG: 16,
		Servings: 1, Description: "Микс из орехов и сухофруктов на 30г",
			Ingredients: `["миндаль 10г","грецкий орех 10г","курага 10г","чернослив 10г"]`, Steps: `["Смешать всё в одной порции"]`},
		{Title: "Ролл из лаваша с курицей", MealType: model.MealSnack, Calories: 340, ProteinG: 22, FatG: 14, CarbsG: 28,
		Servings: 1, Description: "Лаваш с курицей, овощами и соусом",
			Ingredients: `["лаваш 1шт","куриная грудка 100г","салат 30г","помидор 1/2шт","огурец 1/2шт","сметана 30г"]`, Steps: `["Обжарить курицу","Нарезать овощи","Завернуть всё в лаваш"]`},
		{Title: "Чипсы из батата", MealType: model.MealSnack, Calories: 160, ProteinG: 2, FatG: 8, CarbsG: 22,
		Servings: 1, Description: "Полезные чипсы из батата в духовке",
			Ingredients: `["батат 200г","оливковое масло 1ст.л.","соль","паприка"]`, Steps: `["Нарезать батат тонкими слайсами","Сбрызнуть маслом и посыпать специями","Запекать 20 мин при 200°C"]`},
		{Title: "Сырные шарики", MealType: model.MealSnack, Calories: 230, ProteinG: 14, FatG: 16, CarbsG: 8,
		Servings: 1, Description: "Закусочные шарики из сыра с чесноком и зеленью",
			Ingredients: `["сыр твёрдый 150г","сливочный сыр 100г","чеснок 2зубчика","укроп 10г","грецкий орех 30г"]`, Steps: `["Натереть сыр","Смешать с творожным сыром и чесноком","Скатать шарики, обвалять в орехах"]`},
		{Title: "Хумус с овощными палочками", MealType: model.MealSnack, Calories: 190, ProteinG: 8, FatG: 12, CarbsG: 16,
		Servings: 1, Description: "Нутовый хумус с морковью и сельдереем",
			Ingredients: `["нут консервированный 200г","тахини 1ст.л.","лимонный сок 1ст.л.","оливковое масло 1ст.л.","морковь 1шт","сельдерей 2стебля"]`, Steps: `["Измельчить нут блендером","Добавить тахини, лимон и масло","Нарезать овощи палочками для подачи"]`},
		{Title: "Банановые панкейки", MealType: model.MealSnack, Calories: 220, ProteinG: 8, FatG: 6, CarbsG: 36,
		Servings: 1, Description: "Мини-панкейки на банановой основе",
			Ingredients: `["банан 1шт","яйцо 1шт","мука овсяная 50г","мёд 1ст.л.","масло кокосовое"]`, Steps: `["Размять банан","Смешать с яйцом и мукой","Жарить маленькие блинчики 2 мин с каждой стороны"]`},
		{Title: "Творог с зеленью", MealType: model.MealSnack, Calories: 140, ProteinG: 18, FatG: 6, CarbsG: 6,
		Servings: 1, Description: "Свежий творог с зеленью и огурцом",
			Ingredients: `["творог 150г","огурец 1шт","укроп 10г","чеснок 1зубчик","соль"]`, Steps: `["Натереть огурец","Смешать с творогом и зеленью","Посолить по вкусу"]`},
		{Title: "Авокадо с креветками", MealType: model.MealSnack, Calories: 220, ProteinG: 12, FatG: 16, CarbsG: 6,
		Servings: 1, Description: "Фаршированное авокадо с салатными креветками",
			Ingredients: `["авокадо 1шт","креветки отварные 80г","лимонный сок 1ч.л.","оливковое масло 1ч.л.","перец чёрный"]`, Steps: `["Разрезать авокадо пополам, удалить косточку","Смешать креветки с лимоном и маслом","Выложить в половинки авокадо"]`},
	}

	if err := r.db.Create(&recipes).Error; err != nil {
		return err
	}
	return nil
}
