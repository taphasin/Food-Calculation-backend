package models

type FoodDisplay struct {
	Meal     string `json:"meal"`
	FoodId   string `json:"foodID"`
	ImageUrl string `json:"imageUrl"`
	Name     string `json:"name"`
	Calories int    `json:"calories"`
	Dish     int    `json:"dish"`
}
