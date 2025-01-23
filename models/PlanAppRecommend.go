package models

import "time"

type PlanAppRecommend struct {
	PlannerID   string    `json:"plannerID"`
	PlannerDate time.Time `json:"plannerDate"`
	UserID      string    `json:"userID"`
	PlanName    string    `json:"planname"`
	CreatedAt   time.Time `json:"createdAt"`
	Calories    int       `json:"calories"`
	Breakfast   []string  `json:"breakfast"`
	Lunch       []string  `json:"lunch"`
	Dinner      []string  `json:"dinner"`
	Others      []string  `json:"others"`
}
