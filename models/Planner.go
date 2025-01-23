package models

import "time"

type Planner struct {
	PlannerID   string    `json:"plannerID"`
	PlannerDate time.Time `json:"plannerDate"`
	UserID      string    `json:"userID"`
	PlanName    string    `json:"planname"`
	CreatedAt   time.Time `json:"createdAt"`
}
