package lib

import (
	"math"
	"strconv"
	"strings"
)

func CaloriePerDay(gender string, weight_management string, weightdef int, weightdiff int, height int, age int, activity_level string, additional_opts []string, time string) (int, int) {
	var tdee float32
	var bmr float32
	var act_level float32 = Activity(activity_level)
	var calorieperday = CheckWeightManagement(weight_management)
	// var calorieperday = 500

	//find TDEE (Total Daily Energy Expenditure)
	if gender == "male" {
		bmr = (10 * float32(weightdef)) + (6.25 * float32(height)) - (5 * float32(age)) + 5
	} else if gender == "female" {
		bmr = (10 * float32(weightdef)) + (6.25 * float32(height)) - (5 * float32(age)) - 161
	}
	// fmt.Printf("BMR: %f \n", bmr)

	//TDEE
	tdee = bmr * act_level
	// fmt.Printf("BMR: %f \n", tdee)

	//find weight
	var weight int = (weightdiff - weightdef)
	// fmt.Printf("weight: %d \n", weight)

	var totalDay = math.Abs(float64(TimeCompareStr(time, weight_management, weight) * 7))

	var totalcalorie_requirement float32 = float32(calorieperday) + tdee
	// fmt.Printf("totalcalorie_requirement: %f \n", totalcalorie_requirement)

	var totalcalorie_requirement_int int = int(totalcalorie_requirement)
	// fmt.Printf("totalcalorie_requirement: %d \n", totalcalorie_requirement_int)

	return totalcalorie_requirement_int, int(totalDay)
}

func Activity(activity_level string) float32 {
	var act_level float32
	switch activity_level {
	case "Sedentary":
		act_level = 1.2
	case "1-3 day/week":
		act_level = 1.375
	case "3-5 day/week":
		act_level = 1.54
	case "6-7 day/week":
		act_level = 1.725
	case "2 times/day":
		act_level = 1.9
	}
	return act_level
}

func TimeCompareStr(time string, w_management string, weight int) int {
	var week int = 0
	if time != "" && w_management == "Stable" {
		t := strings.Split(time, " ")
		// string to int
		i, err := strconv.Atoi(t[0])
		if err != nil {
			// ... handle error
			panic(err)
		}
		week = i
	} else if w_management == "Loss" || w_management == "Gain" {
		week_float := float32(weight) / 0.5
		week = int(week_float)
	}

	return week
}

func CheckWeightManagement(w_management string) int {
	var cal int = 500
	if w_management == "Loss" {
		cal = cal * (-1)
	} else if w_management == "Stable" {
		cal = 0
	}
	return cal
}
