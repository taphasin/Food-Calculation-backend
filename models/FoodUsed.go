package models

type FoodUsed struct {
	Meal     string   `json:"mealID"`
	FoodId   string   `json:"foodID"`
	ImageUrl string   `json:"imageUrl"`
	Name     string   `json:"name"`
	Tags     []string `json:"tags"`
	Calories int      `json:"calories"`
	Disk     int      `json:"disk"`
}