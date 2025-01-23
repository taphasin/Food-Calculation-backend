package lib

type TemplateResponse struct {
	FoodCalculated  Food_calculated
}

type Ingredient_body_st struct {
	Name string `json:"name"`
	Gram int    `json:"gram"`
}

type Food_calculated struct {
	Calories int `json:"calories"`
	Fat     int `json:"fats"`
	Protein  int `json:"protein"`
	Carb    int `json:"carbs"`
}

type IngredientResponse struct {
	FoodCalculated  Food_calculated
}