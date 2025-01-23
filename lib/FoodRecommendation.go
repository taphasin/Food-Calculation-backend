package lib

import (
	models "Softdev/models"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ----------	global variable ------------------
var MealBreakfast []models.Food
var MealLunch []models.Food
var MealDinner []models.Food
var MealOther []models.Food

var FoodDB []models.Food

var fBuffer []models.Food

var plannersData []models.PlanAppRecommend
var plannerData models.PlanAppRecommend

var total_cal int

//---------- end global variable ------------------

// var cal int
var c_remainder int

func CheckTag(tags []string, additional_opts string) (int, int, int, int, int) {
	countB := 0
	countL := 0
	countD := 0
	countO := 0
	c := 0
	for index := 0; index < len(tags); index++ {
		switch tags[index] {
		case "Breakfast":
			countB += 1
		case "Lunch":
			countL += 1
		case "Dinner":
			countD += 1
		case "Other":
			countO += 1
		}
		if additional_opts == tags[index] {
			c += 1
		}
	}
	if c > 0 {
		return 1, countB, countL, countD, countO
	} else if c == 0 && len(additional_opts) == 0 {
		return 0, countB, countL, countD, countO
	} else {
		return 0, 0, 0, 0, 0
	}
}

func containsFood(listOfFood []models.Food, food models.Food) bool {
	for _, item := range listOfFood {
		if item.FoodId == food.FoodId {
			return true // Found a matching Food struct
		}
	}
	return false // No match found
}

func containsMeal(listOfFood []models.Food, food models.Food) bool {
	for _, item := range listOfFood {
		if item.FoodId == food.FoodId {
			return false // Found a matching Food struct
		}
	}
	return true // No match found
}

// Function check size of Data in list is it enough for Process
func checkDataSize(listofFood []models.Food, additional_opts []string, time int) (bool, []models.Food) {
	countB := 0
	countL := 0
	countD := 0
	countO := 0
	var listFood []models.Food
	CanUse := false
	if len(additional_opts) != 0 {

		for _, additional_opt := range additional_opts {
			for _, Food := range listofFood {
				//debug fmt.Print(Food)
				checkt, c_b, c_l, c_d, c_o := CheckTag(Food.Tags, additional_opt)
				//debug fmt.Println(checkt)
				if checkt == 1 {
					if !(containsFood(listFood, Food)) {
						listFood = append(listFood, Food)
						countB += c_b
						countL += c_l
						countD += c_d
						countO += c_o
					}
				}
			}
		}
	} else {
		for _, Food := range listofFood {
			listFood = append(listFood, Food)
			_, c_b, c_l, c_d, c_o := CheckTag(Food.Tags, "")
			countB += c_b
			countL += c_l
			countD += c_d
			countO += c_o
		}
	}
	if countB >= time && countL >= time && countD >= time && countO >= 0 {
		CanUse = true
	}
	return CanUse, listFood
}

func CreateListofMeals(listofFood []models.Food) ([]models.Food, []models.Food, []models.Food, []models.Food) {
	for _, F := range listofFood {
		for _, tag := range F.Tags {
			switch tag {
			case "Breakfast":
				if !(containsFood(MealBreakfast, F)) {
					MealBreakfast = append(MealBreakfast, F)
				}
			case "Lunch":
				if !(containsFood(MealLunch, F)) {
					MealLunch = append(MealLunch, F)
				}
			case "Dinner":
				if !(containsFood(MealDinner, F)) {
					MealDinner = append(MealDinner, F)
				}
			case "Other":
				if !(containsFood(MealOther, F)) {
					MealOther = append(MealOther, F)
				}
			}
		}
	}
	return MealBreakfast, MealLunch, MealDinner, MealOther
}

func AddtoPlannerData(f models.Food, meal string) {
	switch meal {
	case "Breakfast":
		plannerData.Breakfast = append(plannerData.Breakfast, f.Name)
	case "Lunch":
		plannerData.Lunch = append(plannerData.Lunch, f.Name)
	case "Dinner":
		plannerData.Dinner = append(plannerData.Dinner, f.Name)
	case "Other":
		plannerData.Others = append(plannerData.Others, f.Name)
	}
}

// generate food for each day (Meal) in main create Plnner instance
func CreateMeal(mealId string, mealsctr models.MealController, calperday int) {
	c := float32(calperday) * mealsctr.Proportion_Cal
	c_min := c - 50
	c_max := c + 50
	cmin := int(c_min)
	cmax := int(c_max)
	cint := int(c)

	count := 0

	temp := models.Food{}
	// rand.Seed(time.Now().UnixNano())

	var lst []models.Food

	switch mealsctr.Meal {
	case "Breakfast":
		lst = MealBreakfast
		// fmt.Println(lst)
	case "Lunch":
		lst = MealLunch
		// fmt.Println(lst)
	case "Dinner":
		lst = MealDinner
		// fmt.Println(lst)
	case "Other":
		lst = MealOther
		// fmt.Println(lst)
	}

	if mealsctr.Meal == "Other" && (c_remainder*(-1)) < cmin {
		// fmt.Printf("%d", c_remainder)
		return
	}

	for {
		//------ debug --------
		if mealsctr.Meal == "Other" {
			// fmt.Printf("Cal Remainder : %d \r\n", c_remainder*(-1))
		}
		//---- end debug ----
		index := rand.Intn(len(lst))
		f := lst[index]
		if f.Calories <= cmax && f.Calories >= cmin && containsMeal(fBuffer, f) {
			ff := f
			// แก้ไข
			ff.Meal = mealId

			fBuffer = append(fBuffer, ff)
			// ff.FoodId = GenerateFoodId()
			FoodDB = append(FoodDB, ff)
			c_remainder += ff.Calories - cint
			total_cal += ff.Calories

			//Add data to Planner (for frontend)
			AddtoPlannerData(ff, mealsctr.Meal)
			break
		} else if count >= 50 {
			temp.Meal = mealId
			fBuffer = append(fBuffer, temp)
			// temp.FoodId = GenerateFoodId()
			c_remainder += temp.Calories - cint
			total_cal += temp.Calories
			FoodDB = append(FoodDB, temp)

			//Add data to Planner (for frontend)
			AddtoPlannerData(temp, mealsctr.Meal)
			break
		} else {
			if math.Abs(float64(f.Calories)-float64(cint)) <= math.Abs(float64(temp.Calories)-float64(cint)) {
				temp = f
			}
		}
		count += 1
		// fmt.Printf("Remainder: %d \n", c_remainder)

	}

}

func FoodRecommendation(calperday int, t int, addtional_opts []string, listofFood []models.Food) (string, []models.Food, []models.Meal, []models.Planner, []models.PlanAppRecommend) {
	// planners := []models.PlanAppRecommend{}
	// meals := []models.Meal{}
	// foods := []models.Food{}
	// exercises := []models.Exercise{}
	// ingredients := []models.Ingredient{}

	//clear prev data from varible
	MealBreakfast = nil
	MealLunch = nil
	MealDinner = nil
	MealOther = nil
	plannersData = nil

	mealsctr := [4]models.MealController{
		{Meal: "Breakfast", Proportion_Cal: 0.3},
		{Meal: "Lunch", Proportion_Cal: 0.4},
		{Meal: "Dinner", Proportion_Cal: 0.4},
		{Meal: "Other", Proportion_Cal: 0.1},
	}
	status := ""

	//make list of Food By Additional_opts and return status if it can make planner using this opts
	var listFood []models.Food
	s, listFood := checkDataSize(listofFood, addtional_opts, t)

	var listPlanners []models.Planner
	var listMeals []models.Meal

	FoodDB = nil

	if s {
		status = "OK"
		//For loop to Create list of Meals
		CreateListofMeals(listFood)
		timeNow := time.Now()

		//loop per day
		for i := 0; i < t; i++ {
			//crate planner Id for 1 day
			planner := models.Planner{}
			plannerData = models.PlanAppRecommend{}
			planid := uuid.New()
			splitID := strings.Split(planid.String(), "-")
			pid := splitID[0] + splitID[1] + splitID[2] + splitID[3] + splitID[4] //planner ID

			//add info to planner
			planner.PlannerDate = timeNow.AddDate(0, 0, i)
			planner.PlannerID = pid
			planner.CreatedAt = planner.PlannerDate

			//add info to planner (for frontend)
			plannerData.PlannerDate = planner.PlannerDate
			plannerData.PlannerID = planner.PlannerID
			plannerData.CreatedAt = planner.CreatedAt

			//define empty list for Ohter (case can be null)
			plannerData.Others = []string{}
			for _, m := range mealsctr {
				meal := models.Meal{}
				mealID := uuid.New()
				splitID := strings.Split(mealID.String(), "-")
				mid := splitID[0] + splitID[1] + splitID[2] + splitID[3] + splitID[4] //meal ID

				meal.MealID = mid
				meal.Meal = m.Meal
				meal.Planner = pid

				//fmt.Printf("Meals test1: %s", m.Meal)
				CreateMeal(mid, m, calperday) //function recommend food

				listMeals = append(listMeals, meal)

			}
			//add total calorie to PlannerData
			plannerData.Calories = total_cal

			//clear buffer and remainder cal
			fBuffer = nil
			c_remainder = 0
			total_cal = 0
			//add Planner to list (frontend data)
			plannersData = append(plannersData, plannerData)

			listPlanners = append(listPlanners, planner)
		}
	} else {
		status = "Failed"

		// return status, []models.Food{}
		return status, []models.Food{}, []models.Meal{}, []models.Planner{}, []models.PlanAppRecommend{}
	}

	return status, FoodDB, listMeals, listPlanners, plannersData
	// return status, listFood
}
