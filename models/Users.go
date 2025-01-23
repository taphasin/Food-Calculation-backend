package models

type Users struct {
	UserID       string `json:"UserID"`
	Email        string `json:"Email"`
	Password     string `json:"password"`
	Username     string `json:"Username"`
	Description  string `json:"Description"`
	ProfileImage string `json:"ProfileImage"`
	Age          int    `json:"Age"`
	Height       int    `json:"Height"`
	Weight       int    `json:"Weight"`
	Gender       string `json:"Gender"`
}
