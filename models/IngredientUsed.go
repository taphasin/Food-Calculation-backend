package models

type IngredientUsed struct {
	FoodId             string `json:"FoodId"`
	Name               string `json:"Name"`
	ImageUrl           string `json:"ImageUrl"`
	Protein            int    `json:"Protein"`
	Fat                int    `json:"Fat"`
	Carb               int    `json:"Carb"`
	Grams               int    `json:"Grams"`
}
