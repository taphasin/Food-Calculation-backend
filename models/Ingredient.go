package models

type Ingredient struct {
	FoodId   string `json:"foodID"`
	Name     string `json:"name"`
	ImageUrl string `json:"imageUrl"`
	Protein  int    `json:"protein"`
	Fat      int    `json:"fat"`
	Carb     int    `json:"carb"`
	Grams     int    `json:"grams"`
}