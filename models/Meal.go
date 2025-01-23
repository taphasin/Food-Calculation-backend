package models

type Meal struct {
	Planner string `json:"plannerID"`
	MealID  string `json:"mealID"`
	Meal    string `json:"meal"`
}
