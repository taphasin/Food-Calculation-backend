package main

import (
	"Softdev/lib"
	"Softdev/models"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	//	"text/template/parse"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/mitchellh/mapstructure"

	///	"golang.org/x/net/route"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

var PlannerDB []models.Planner
var MealDB = []models.Meal{}
var FoodDB = []models.Food{}
var IngredientDB = []models.Ingredient{}

var FoodDisplay = []models.FoodDisplay{}

var PlannersData []models.PlanAppRecommend

var current_Food = models.Food{}
var current_Ingredient = []models.Ingredient{}
var current_Calculation = lib.Food_calculated{}

var daily_Food_Breakfast = []models.Food{}
var daily_Food_Lunch = []models.Food{}
var daily_Food_Dinner = []models.Food{}
var daily_Food_Other = []models.Food{}

var daily_Ingredient_Breakfast = []models.Ingredient{}
var daily_Ingredient_Lunch = []models.Ingredient{}
var daily_Ingredient_Dinner = []models.Ingredient{}
var daily_Ingredient_Other = []models.Ingredient{}

var daily_Planner = models.Planner{}

var daily_Meal_Breakfast = models.Meal{Meal: "Breakfast"}
var daily_Meal_Lunch = models.Meal{Meal: "Lunch"}
var daily_Meal_Dinner = models.Meal{Meal: "Dinner"}
var daily_Meal_Other = models.Meal{Meal: "Other"}

var food_calculated = 0
var ingredient_selected = 0
var food_selected = 0
var s = false

// proportion
var proteinPerDay = 0
var carbPerDay = 0
var fatPerDay = 0
var calperDay = 0

type App struct {
	Router *mux.Router
	client *firestore.Client
	ctx    context.Context
}

func main() {
	godotenv.Load()
	route := App{}
	daily_Meal_Breakfast.MealID = route.GenerateFoodId()
	daily_Meal_Lunch.MealID = route.GenerateFoodId()
	daily_Meal_Dinner.MealID = route.GenerateFoodId()
	daily_Meal_Other.MealID = route.GenerateFoodId()
	route.Init()
	route.Run()

}

// REST API
func (route *App) initializeRoutes() {
	route.Router.HandleFunc("/", route.Home).Methods("GET")
	route.Router.HandleFunc("/users/{userID}", route.GetUser).Methods("GET")
	route.Router.HandleFunc("/users/{userID}", route.EditUser).Methods("PUT")
	route.Router.HandleFunc("/users/{userID}", route.DeleteUser).Methods("DELETE")
	route.Router.HandleFunc("/Create", route.CreateFood).Methods("POST")
	route.Router.HandleFunc("/FoodRecommendation", route.FoodRecommendation).Methods("POST")
	route.Router.HandleFunc("/FoodRecommendationView", route.FoodRecommendationView).Methods("GET")
	route.Router.HandleFunc("/FoodRecommendation/{pid}", route.ShowPlanner).Methods("GET") //หน้าที่แสดงผลว่าวันนี้มีกินอะไรบ้าง cal รวม เท่าไหร่ และ สัดส่วน โปรตีน, คาร์โบ, ไขมัน
	route.Router.HandleFunc("/AddtoPlanner/{id}", route.AddtoPlanner).Methods("POST")
	route.Router.HandleFunc("/planners/{userID}", route.GetPlannerByUserID).Methods("GET")
	route.Router.HandleFunc("/planners/planname/{planName}", route.GetPlannerByPlanName).Methods("GET")
	route.Router.HandleFunc("/planners/{plannerID}/meals", route.GetDetailFromPlannerID).Methods("GET")
	route.Router.HandleFunc("/food/{foodID}/editDish", route.EditFoodDish).Methods("PUT")
	route.Router.HandleFunc("/delete/planner/{planName}", route.DeletePlannerFromPlanName).Methods("DELETE")
	route.Router.HandleFunc("/delete/planner/planid/{plannerID}", route.DeletePlannerByID).Methods("DELETE")
	route.Router.HandleFunc("/delete/food/{foodID}", route.DeleteFoodByID).Methods("DELETE")
	route.Router.HandleFunc("/signup", route.SignUp).Methods("POST")
	route.Router.HandleFunc("/signin", route.SignIn).Methods("POST")
	route.Router.HandleFunc("/signout", route.SignOut).Methods("POST")
	route.Router.HandleFunc("/changepassword", route.ChangePassword).Methods("POST")
	route.Router.HandleFunc("/getCurrentUser", route.GetCurrentUser).Methods("GET")
	route.Router.HandleFunc("/deleteAccount", route.DeleteAccount).Methods("DELETE")
	route.Router.HandleFunc("/Selectfood/{id}", route.SelectFood).Methods("GET")
	route.Router.HandleFunc("/Selectingredient/{id}", route.SelectIngredient).Methods("GET")
	route.Router.HandleFunc("/Calculation/{id}/{dish}", route.Calculation).Methods("POST")
	route.Router.HandleFunc("/GetPlanner/{UserID}", route.GetPlanner).Methods("GET")
	route.Router.HandleFunc("/Getcurrentfood", route.Get_current_food).Methods("GET")
	route.Router.HandleFunc("/Getcurrentcalculation/{foodid}", route.Get_current_calculation).Methods("GET")
	route.Router.HandleFunc("/Getcurrentingredient", route.Get_current_ingredient).Methods("GET")
	route.Router.HandleFunc("/Getfoodfrommeal/{meal}", route.Get_food_from_meal).Methods("GET")
	route.Router.HandleFunc("/Addtomeal/{meal}", route.AddToMeal).Methods("GET")
	route.Router.HandleFunc("/Deletefrommeal/{meal}/{fid}", route.DeleteFromMeal).Methods("GET")
	route.Router.HandleFunc("/Calculationmeal/{meal}", route.Calculation_meal).Methods("GET")
	route.Router.HandleFunc("/Calculationplanner", route.Calculation_planner).Methods("GET")
	route.Router.HandleFunc("/Addtoexistingplanner/{planname}/{meal}", route.AddToExistingPlanner).Methods("POST")

	route.Router.HandleFunc("/AddDailyMealToDB/{userid}/{plan_name}", route.AddDailyMealToDB).Methods("POST")
}

func (route *App) Init() {
	route.ctx = context.Background()

	// Firebase setup
	sa := option.WithCredentialsFile("appsoftdev-d4314-firebase-adminsdk-vyv9x-3327a98e65.json")
	app, err := firebase.NewApp(route.ctx, nil, sa)
	if err != nil {
		log.Fatalln(err)
	}

	route.client, err = app.Firestore(route.ctx)
	if err != nil {
		log.Fatalln(err)
	}

	route.Router = mux.NewRouter()
	route.initializeRoutes()
	fmt.Println("Successfully connected at port : " + route.GetPort())
}

// Run starts the server and prints the local IP
//
//	func (route *App) Run() {
//		ipAddress := getLocalIP()
//		if ipAddress != "" {
//			fmt.Printf("Server running at http://%s%s\n", ipAddress, route.GetPort())
//		} else {
//			fmt.Println("Unable to determine local IP address")
//		}
//		log.Fatal(http.ListenAndServe("0.0.0.0"+route.GetPort(), route.Router))
//	}
func (route *App) Run() {
	fmt.Println("Server running on port", route.GetPort())
	log.Fatal(http.ListenAndServe("0.0.0.0"+route.GetPort(), route.Router)) // Bind to 0.0.0.0
}

// Get the port from the environment or use the default port
func (route *App) GetPort() string {
	var port = os.Getenv("Myport")
	if port == "" {
		port = "80"
	}
	return ":" + port
}

// Helper function to get the local IP address of the machine
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}

	for _, address := range addrs {
		// Check if the IP address is not a loopback and is an IPv4 address
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			return ipnet.IP.String()
		}
	}
	return ""
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func (route *App) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		IDToken  string `json:"idToken"`
		Password string `json:"password"`
	}

	// Decode request body
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil || payload.IDToken == "" || payload.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Firebase REST API URL for updating the password
	apiKey := os.Getenv("FIREBASE_API_KEY")
	updatePasswordUrl := fmt.Sprintf("https://identitytoolkit.googleapis.com/v1/accounts:update?key=%s", apiKey)

	// Prepare the payload
	body := map[string]interface{}{
		"idToken":           payload.IDToken,
		"password":          payload.Password,
		"returnSecureToken": true,
	}

	// Convert the payload to JSON
	jsonPayload, err := json.Marshal(body)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Send the request to Firebase Authentication
	req, err := http.NewRequest("POST", updatePasswordUrl, strings.NewReader(string(jsonPayload)))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error connecting to Firebase")
		return
	}
	defer resp.Body.Close()

	// Decode Firebase response
	var responseData map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&responseData)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error decoding response")
		return
	}

	// Check for errors in Firebase response
	if resp.StatusCode != http.StatusOK {
		respondWithError(w, http.StatusUnauthorized, responseData["error"].(map[string]interface{})["message"].(string))
		return
	}

	// Respond with a success message
	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message":   "Password updated successfully",
		"idToken":   responseData["idToken"],
		"email":     responseData["email"],
		"expiresIn": responseData["expiresIn"],
	})
}

// Sign-Up Handler
func (route *App) SignUp(w http.ResponseWriter, r *http.Request) {
	log.Println("Received request at /signup")
	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// Decode request body
	err := json.NewDecoder(r.Body).Decode(&credentials)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Firebase REST API URL for sign-up
	apiKey := os.Getenv("FIREBASE_API_KEY") // Your Firebase API Key
	signUpUrl := fmt.Sprintf("https://identitytoolkit.googleapis.com/v1/accounts:signUp?key=%s", apiKey)

	// Prepare the payload
	payload := map[string]interface{}{
		"email":             credentials.Email,
		"password":          credentials.Password,
		"returnSecureToken": true,
	}
	log.Printf("Email: %s, Password: %s", credentials.Email, credentials.Password)

	// Convert the payload to JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Send the request to Firebase Authentication
	req, err := http.NewRequest("POST", signUpUrl, strings.NewReader(string(jsonPayload)))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error connecting to Firebase")
		return
	}
	defer resp.Body.Close()

	// Decode Firebase response
	var responseData map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&responseData)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error decoding response")
		return
	}

	// Check for errors in Firebase response
	if resp.StatusCode != http.StatusOK {
		respondWithError(w, http.StatusUnauthorized, responseData["error"].(map[string]interface{})["message"].(string))
		return
	}

	// Firebase authentication was successful, now create a Users document in Firestore
	userId := responseData["localId"].(string)
	defaultUser := models.Users{
		UserID:       userId,
		Email:        credentials.Email,
		Password:     credentials.Password, // In practice, you should hash passwords before storing
		Username:     "New User",
		Description:  "This is a new user.",
		ProfileImage: "default_image_url",
		Age:          18,
		Height:       170,
		Weight:       70,
		Gender:       "unknown",
	}

	// Store the default user information in Firestore
	_, _, err = route.client.Collection("User").Add(route.ctx, defaultUser)
	if err != nil {
		log.Printf("Error creating user in Firestore: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Error creating user in Firestore")
		return
	}

	// Respond with the sign-up confirmation and user data
	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"userId":  userId,
		"email":   credentials.Email,
		"message": "Sign-up successful, user created in Firestore",
	})
}

// // Sign-Up Handler
// func (route *App) SignUp(w http.ResponseWriter, r *http.Request) {
// 	log.Println("Received request at /signup")
// 	var credentials struct {
// 		Email    string `json:"email"`
// 		Password string `json:"password"`
// 	}

// 	// Decode request body
// 	err := json.NewDecoder(r.Body).Decode(&credentials)
// 	if err != nil {
// 		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
// 		return
// 	}

// 	// Firebase REST API URL for sign-up
// 	apiKey := os.Getenv("FIREBASE_API_KEY") // Your Firebase API Key
// 	signUpUrl := fmt.Sprintf("https://identitytoolkit.googleapis.com/v1/accounts:signUp?key=%s", apiKey)

// 	// Prepare the payload
// 	payload := map[string]interface{}{
// 		"email":             credentials.Email,
// 		"password":          credentials.Password,
// 		"returnSecureToken": true,
// 	}
// 	log.Printf("Email: %s, Password: %s", credentials.Email, credentials.Password)
// 	// Convert the payload to JSON
// 	jsonPayload, err := json.Marshal(payload)
// 	if err != nil {
// 		respondWithError(w, http.StatusInternalServerError, "Internal server error")
// 		return
// 	}

// 	// Send the request to Firebase Authentication
// 	req, err := http.NewRequest("POST", signUpUrl, strings.NewReader(string(jsonPayload)))
// 	req.Header.Set("Content-Type", "application/json")

// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		respondWithError(w, http.StatusInternalServerError, "Error connecting to Firebase")
// 		return
// 	}
// 	defer resp.Body.Close()

// 	// Decode Firebase response
// 	var responseData map[string]interface{}
// 	err = json.NewDecoder(resp.Body).Decode(&responseData)
// 	if err != nil {
// 		respondWithError(w, http.StatusInternalServerError, "Error decoding response")
// 		return
// 	}

// 	// Check for errors in Firebase response
// 	if resp.StatusCode != http.StatusOK {
// 		respondWithError(w, http.StatusUnauthorized, responseData["error"].(map[string]interface{})["message"].(string))
// 		return
// 	}

// 	// Respond with the sign-up confirmation
// 	respondWithJSON(w, http.StatusOK, map[string]interface{}{
// 		"userId": responseData["localId"],
// 		"email":  responseData["email"],
// 	})
// }

// Sign-In Handler
func (route *App) SignIn(w http.ResponseWriter, r *http.Request) {
	log.Println("Received request at /signin")
	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// Decode request body
	err := json.NewDecoder(r.Body).Decode(&credentials)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Firebase REST API URL for sign-in
	apiKey := os.Getenv("FIREBASE_API_KEY") // Your Firebase API Key
	signInUrl := fmt.Sprintf("https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key=%s", apiKey)

	// Prepare the payload
	payload := map[string]interface{}{
		"email":             credentials.Email,
		"password":          credentials.Password,
		"returnSecureToken": true,
	}
	log.Printf("Email: %s, Password: %s", credentials.Email, credentials.Password)

	// Convert the payload to JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Send the request to Firebase Authentication
	req, err := http.NewRequest("POST", signInUrl, strings.NewReader(string(jsonPayload)))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error connecting to Firebase")
		return
	}
	defer resp.Body.Close()

	// Decode Firebase response
	var responseData map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&responseData)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error decoding response")
		return
	}

	// Check for errors in Firebase response
	if resp.StatusCode != http.StatusOK {
		respondWithError(w, http.StatusUnauthorized, responseData["error"].(map[string]interface{})["message"].(string))
		return
	}

	// Respond with the ID token
	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"idToken":   responseData["idToken"],
		"userId":    responseData["localId"],
		"email":     responseData["email"],
		"expiresIn": responseData["expiresIn"],
	})
}

func (route *App) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	// Extract the Authorization header (Bearer token)
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || len(authHeader) < 7 || !strings.HasPrefix(authHeader, "Bearer ") {
		respondWithError(w, http.StatusUnauthorized, "ID token is required")
		return
	}

	// Extract the ID token from the Authorization header
	idToken := authHeader[7:] // Remove "Bearer " prefix

	// Initialize Firebase app
	app := initFirebaseApp()
	ctx := context.Background()
	client, err := app.Auth(ctx)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error initializing Firebase Authentication")
		return
	}

	// Verify the ID token
	token, err := client.VerifyIDToken(ctx, idToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, fmt.Sprintf("Error verifying ID token: %v", err))
		return
	}

	// Respond with the UID and token information
	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"uid":     token.UID,
		"idToken": idToken,
	})
}

// Sign-Out Handler (Revoke Tokens)
func (route *App) SignOut(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		UID string `json:"uid"`
	}

	// Decode the request body to get the UID
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil || payload.UID == "" {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Initialize Firebase Admin SDK
	app := initFirebaseApp()
	ctx := context.Background()
	client, err := app.Auth(ctx)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error initializing Firebase Authentication")
		return
	}

	// Revoke all refresh tokens for the specified user
	err = client.RevokeRefreshTokens(ctx, payload.UID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error revoking tokens")
		return
	}

	// Success message
	respondWithJSON(w, http.StatusOK, map[string]string{"message": "User signed out successfully"})
}
func (route *App) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	// Extract the Authorization header (Bearer token)
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || len(authHeader) < 7 || !strings.HasPrefix(authHeader, "Bearer ") {
		respondWithError(w, http.StatusUnauthorized, "ID token is required")
		return
	}

	// Extract the ID token from the Authorization header
	idToken := authHeader[7:] // Remove "Bearer " prefix

	// Initialize Firebase app
	app := initFirebaseApp()
	ctx := context.Background()
	client, err := app.Auth(ctx)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error initializing Firebase Authentication")
		return
	}

	// Verify the ID token
	token, err := client.VerifyIDToken(ctx, idToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, fmt.Sprintf("Error verifying ID token: %v", err))
		return
	}

	// Delete the user account using the UID from the token
	err = client.DeleteUser(ctx, token.UID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error deleting user: %v", err))
		return
	}

	// Respond with success message
	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "User account deleted successfully",
	})
}

func (route *App) Home(w http.ResponseWriter, r *http.Request) {
	FoodsData := []models.Food{}
	food_calculated, food_selected, ingredient_selected = 0, 0, 0
	fmt.Println(food_calculated)
	fmt.Println(food_selected)
	fmt.Println(ingredient_selected)
	iter := route.client.Collection("Food").Documents(route.ctx)
	for {
		FoodData := models.Food{}
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			respondWithJSON(w, http.StatusInternalServerError, "Something wrong, please try again.")
		}
		mapstructure.Decode(doc.Data(), &FoodData)
		FoodsData = append(FoodsData, FoodData)
	}
	respondWithJSON(w, http.StatusOK, FoodsData)
}

// Add Food to firestore for Food template
func (route *App) CreateFood(w http.ResponseWriter, r *http.Request) {
	foodID, _ := uuid.NewV6()
	splitID := strings.Split(foodID.String(), "-")
	id := splitID[0] + splitID[1] + splitID[2] + splitID[3] + splitID[4]

	FoodData := models.Food{}

	Decoder := json.NewDecoder(r.Body)

	err := Decoder.Decode(&FoodData)

	FoodData.FoodId = id

	if err != nil {
		log.Printf("error: ")
	}

	_, _, err = route.client.Collection("Food").Add(route.ctx, FoodData)

	if err != nil {
		log.Printf("An error has occurred: %s", err)
	}

	respondWithJSON(w, http.StatusCreated, "Created Food sucess!")
}

func (route *App) GenerateFoodId() string {
	id := uuid.New()
	splitID := strings.Split(id.String(), "-")
	fid := splitID[0] + splitID[1] + splitID[2] + splitID[3] + splitID[4] //planner ID
	return fid
}

func (route *App) SelectFood(w http.ResponseWriter, r *http.Request) {
	if food_selected == 1 {
		route.Get_current_food(w, r)
		return
	}
	params := mux.Vars(r)
	paramsID := params["id"]

	iter := route.client.Collection("Food").Where("Name", "==", paramsID).Documents(route.ctx)

	foodFound := false
	for {
		FoodData := models.Food{}
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			respondWithJSON(w, http.StatusInternalServerError, "Something wrong, please try again.")
		}
		mapstructure.Decode(doc.Data(), &FoodData)
		current_Food = FoodData
		foodFound = true
	}

	if !foodFound {
		respondWithJSON(w, http.StatusNotFound, "No matching food found")
		return
	}
	food_selected = 1
	respondWithJSON(w, http.StatusOK, current_Food)
}

func (route *App) SelectIngredient(w http.ResponseWriter, r *http.Request) {
	fmt.Println(current_Ingredient)
	if ingredient_selected == 1 {
		route.Get_current_ingredient(w, r)
		return
	}
	params := mux.Vars(r)
	paramsID := params["id"]

	current_Ingredient = []models.Ingredient{}

	iter := route.client.Collection("Ingredient").Where("FoodId", "==", paramsID).Documents(route.ctx)

	for {
		IngredientData := models.Ingredient{}
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			respondWithJSON(w, http.StatusInternalServerError, "Something wrong, please try again.")
		}
		mapstructure.Decode(doc.Data(), &IngredientData)

		current_Ingredient = append(current_Ingredient, IngredientData)
	}
	ingredient_selected = 1
	respondWithJSON(w, http.StatusOK, current_Ingredient)
}

func (route *App) GetPlanner(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	uid := params["UserID"]

	PlannersData := []models.Planner{}

	iter := route.client.Collection("PlannerTestAPI").Where("UserID", "==", uid).Documents(route.ctx)

	for {
		PlannerData := models.Planner{}
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			respondWithJSON(w, http.StatusInternalServerError, "Something wrong, please try again.")
		}
		mapstructure.Decode(doc.Data(), &PlannerData)
		PlannersData = append(PlannersData, PlannerData)
	}
	respondWithJSON(w, http.StatusOK, PlannersData)
}
func (route *App) Calculation(w http.ResponseWriter, r *http.Request) {

	id := route.GenerateFoodId()

	params := mux.Vars(r)
	paramsID := params["id"]
	disk := params["dish"]

	Change_calories := 0
	Change_Fats := 0
	Change_protein := 0
	Change_Carbs := 0
	current_Ingredient = []models.Ingredient{}
	current_Food = models.Food{}
	current_Calculation = lib.Food_calculated{}

	//respond
	var calculated_respond lib.Food_calculated

	// find ingredient
	Ingredient_Used := models.Ingredient{}

	// body from user
	Ingredient_body := []models.Ingredient{}
	Decoder := json.NewDecoder(r.Body)
	err := Decoder.Decode(&Ingredient_body)
	if err != nil {
		log.Printf("error: %s", err)
	}

	// get static food data
	iter := route.client.Collection("Food").Where("FoodId", "==", paramsID).Documents(route.ctx)
	Food_Used := models.Food{}
	doc, err := iter.Next()
	if err != nil {
		log.Fatalf("Failed to fetch document: %v", err)
		return
	}
	mapstructure.Decode(doc.Data(), &Food_Used)

	//loop for change from calculate
	for _, item := range Ingredient_body {
		fmt.Printf("Processing food item: Name = %s\n", item.Name)

		iter_a := route.client.Collection("Ingredient").Where("FoodId", "==", paramsID).Where("Name", "==", item.Name).Documents(route.ctx)
		doc, err := iter_a.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Failed to fetch document: %v", err)
			return
		}
		mapstructure.Decode(doc.Data(), &Ingredient_Used)

		multiplier := float64(item.Grams) / float64(Ingredient_Used.Grams)
		fmt.Printf("%v\n", item.Grams)
		fmt.Printf("%v\n", float64(Ingredient_Used.Grams))
		fmt.Printf("%v\n", multiplier)
		fmt.Printf("%v\n", Ingredient_Used)

		Ingredient_Used.Carb = int(float64(Ingredient_Used.Carb) * multiplier)
		Ingredient_Used.Fat = int(float64(Ingredient_Used.Fat) * multiplier)
		Ingredient_Used.Protein = int(float64(Ingredient_Used.Protein) * multiplier)

		Change_Carbs += Ingredient_Used.Carb
		Change_Fats += Ingredient_Used.Fat
		Change_protein += Ingredient_Used.Protein
		Change_calories += (Ingredient_Used.Carb * 4) + (Ingredient_Used.Protein * 4) + (Ingredient_Used.Fat * 9)

		Ingredient_Used.FoodId = id
		Ingredient_Used.Grams = item.Grams
		current_Ingredient = append(current_Ingredient, Ingredient_Used)
		// response.IngredientUseds = append(response.IngredientUseds, Ingredient_Used)
	}

	Food_Used.FoodId = id
	disk_to_int, err := strconv.Atoi(disk)
	if err != nil {
		log.Fatalf("Failed to convert disk to string: %v", err)
		return
	}
	Food_Used.Dish = disk_to_int
	Food_Used.Calories = Change_calories * disk_to_int
	//	route.client.Collection("FoodUsed").Add(route.ctx, Food_Used)
	// response.FoodUseds = append(response.FoodUseds, Food_Used)
	current_Food = Food_Used

	calculated_respond.Calories = Change_calories * disk_to_int
	calculated_respond.Carb = Change_Carbs * disk_to_int
	calculated_respond.Fat = Change_Fats * disk_to_int
	calculated_respond.Protein = Change_protein * disk_to_int

	current_Calculation = calculated_respond
	food_calculated = 1
	respondWithJSON(w, http.StatusOK, calculated_respond)
}

func (route *App) AddToExistingPlanner(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	paramsID := params["planname"]
	meal := params["meal"]

	planter := route.client.Collection("PlannerTestAPI").Where("PlanName", "==", paramsID).Documents(route.ctx)
	PlannerData := models.Planner{}
	doc, err := planter.Next()
	if err == iterator.Done {
		respondWithJSON(w, http.StatusNotFound, "No matching planner name found")
		return
	}
	if err != nil {
		log.Fatalf("Failed to fetch document: %v", err)
		return
	}
	mapstructure.Decode(doc.Data(), &PlannerData)
	planid := PlannerData.PlannerID

	MealData := models.Meal{}

	iter := route.client.Collection("MealTestAPI").Where("Planner", "==", planid).Where("Meal", "==", meal).Documents(route.ctx)
	doc, err = iter.Next()
	if err == iterator.Done {
		respondWithJSON(w, http.StatusNotFound, "No matching meal found")
		return
	}
	if err != nil {
		log.Fatalf("Failed to fetch document: %v", err)
		return
	}
	mapstructure.Decode(doc.Data(), &MealData)
	current_Food.Meal = MealData.MealID

	route.client.Collection("FoodTestAPI").Add(route.ctx, current_Food)

	for _, i := range current_Ingredient {
		route.client.Collection("IngredientTestAPI").Add(route.ctx, i)
	}
	respondWithJSON(w, http.StatusOK, "Add to existing planner success")
}

func (route *App) AddToMeal(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	meal := params["meal"]

	switch meal {
	case "Breakfast":
		current_Food.Meal = daily_Meal_Breakfast.MealID
		daily_Food_Breakfast = append(daily_Food_Breakfast, current_Food)
		daily_Ingredient_Breakfast = append(daily_Ingredient_Breakfast, current_Ingredient...)
	case "Lunch":
		current_Food.Meal = daily_Meal_Lunch.MealID
		daily_Food_Lunch = append(daily_Food_Lunch, current_Food)
		daily_Ingredient_Lunch = append(daily_Ingredient_Lunch, current_Ingredient...)
	case "Dinner":
		current_Food.Meal = daily_Meal_Dinner.MealID
		daily_Food_Dinner = append(daily_Food_Dinner, current_Food)
		daily_Ingredient_Dinner = append(daily_Ingredient_Dinner, current_Ingredient...)
	case "Other":
		current_Food.Meal = daily_Meal_Other.MealID
		daily_Food_Other = append(daily_Food_Other, current_Food)
		daily_Ingredient_Other = append(daily_Ingredient_Other, current_Ingredient...)
	}
	fmt.Print(meal)
	respondWithJSON(w, http.StatusOK, current_Food)
}

func (route *App) DeleteFromMeal(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	meal := params["meal"]
	fid := params["fid"]

	if meal == "Breakfast" {
		for i, f := range daily_Food_Breakfast {
			fmt.Printf("%v\n", i)
			fmt.Printf("%v\n", f.Name)
			fmt.Printf("%v\n", len(daily_Food_Breakfast))
			if f.FoodId == fid {
				daily_Food_Breakfast = append(daily_Food_Breakfast[:i], daily_Food_Breakfast[i+1:]...)
			}
		}
		for i, ingre := range daily_Ingredient_Breakfast {
			if ingre.FoodId == fid {
				daily_Ingredient_Breakfast = append(daily_Ingredient_Breakfast[:i], daily_Ingredient_Breakfast[i+1:]...)
			}
		}
	} else if meal == "Lunch" {
		for i, f := range daily_Food_Lunch {
			if f.FoodId == fid {
				daily_Food_Lunch = append(daily_Food_Lunch[:i], daily_Food_Lunch[i+1:]...)
			}
		}
		for i, ingre := range daily_Ingredient_Lunch {
			if ingre.FoodId == fid {
				daily_Ingredient_Lunch = append(daily_Ingredient_Lunch[:i], daily_Ingredient_Lunch[i+1:]...)
			}
		}

	} else if meal == "Dinner" {
		for i, f := range daily_Food_Dinner {
			if f.FoodId == fid {
				daily_Food_Dinner = append(daily_Food_Dinner[:i], daily_Food_Dinner[i+1:]...)
			}
		}
		for i, ingre := range daily_Ingredient_Dinner {
			if ingre.FoodId == fid {
				daily_Ingredient_Dinner = append(daily_Ingredient_Dinner[:i], daily_Ingredient_Dinner[i+1:]...)
			}
		}

	} else if meal == "Other" {
		for i, f := range daily_Food_Other {
			if f.FoodId == fid {
				daily_Food_Other = append(daily_Food_Other[:i], daily_Food_Other[i+1:]...)
			}
		}
		for i, ingre := range daily_Ingredient_Other {
			if ingre.FoodId == fid {
				daily_Ingredient_Other = append(daily_Ingredient_Other[:i], daily_Ingredient_Other[i+1:]...)
			}
		}
	} else {
		respondWithJSON(w, http.StatusInternalServerError, "meal wrong, please try again.")
	}

	current_Food.Meal = ""
	respondWithJSON(w, http.StatusOK, current_Food)
}

func (route *App) Get_current_calculation(w http.ResponseWriter, r *http.Request) {
	fmt.Print(current_Calculation)
	if food_calculated == 1 {
		respondWithJSON(w, http.StatusOK, current_Calculation)
		return
	}
	params := mux.Vars(r)
	paramsID := params["foodid"]

	Change_calories := 0
	Change_Fats := 0
	Change_protein := 0
	Change_Carbs := 0

	//respond
	calculated_respond := lib.Food_calculated{}
	current_Calculation = lib.Food_calculated{}

	iter := route.client.Collection("Ingredient").Where("FoodId", "==", paramsID).Documents(route.ctx)

	for {
		IngredientData := models.Ingredient{}
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			respondWithJSON(w, http.StatusInternalServerError, "Something wrong, please try again.")
		}
		mapstructure.Decode(doc.Data(), &IngredientData)
		Change_Carbs = Change_Carbs + IngredientData.Carb
		Change_Fats = Change_Fats + IngredientData.Fat
		Change_protein = Change_protein + IngredientData.Protein
		Change_calories = Change_calories + (IngredientData.Carb * 4) + (IngredientData.Fat * 9) + (IngredientData.Protein * 4)
	}

	calculated_respond.Calories = Change_calories
	calculated_respond.Fat = Change_Fats
	calculated_respond.Protein = Change_protein
	calculated_respond.Carb = Change_Carbs

	current_Calculation = calculated_respond
	food_calculated = 1
	respondWithJSON(w, http.StatusOK, calculated_respond)
}

func (route *App) Get_current_food(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, current_Food)
}

func (route *App) Get_current_ingredient(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, current_Ingredient)
}

func (route *App) Get_food_from_meal(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	meal := params["meal"]

	switch meal {
	case "Breakfast":
		respondWithJSON(w, http.StatusOK, daily_Food_Breakfast)
	case "Lunch":
		respondWithJSON(w, http.StatusOK, daily_Food_Lunch)
	case "Dinner":
		respondWithJSON(w, http.StatusOK, daily_Food_Dinner)
	case "Other":
		respondWithJSON(w, http.StatusOK, daily_Food_Other)
	default:
		respondWithJSON(w, http.StatusInternalServerError, "meal wrong, please try again.")
	}
}

func (route *App) Calculation_meal(w http.ResponseWriter, r *http.Request) {
	var total_Fat, total_Protein, total_Carb int
	params := mux.Vars(r)
	meal := params["meal"]

	respond_calculation := lib.Food_calculated{}

	switch meal {
	case "Breakfast":
		for _, i := range daily_Ingredient_Breakfast {
			total_Fat += i.Fat
			total_Protein += i.Protein
			total_Carb += i.Carb
		}
	case "Lunch":
		for _, i := range daily_Ingredient_Lunch {
			total_Fat += i.Fat
			total_Protein += i.Protein
			total_Carb += i.Carb
		}
	case "Dinner":
		for _, i := range daily_Ingredient_Dinner {
			total_Fat += i.Fat
			total_Protein += i.Protein
			total_Carb += i.Carb
		}
	case "Other":
		for _, i := range daily_Ingredient_Other {
			total_Fat += i.Fat
			total_Protein += i.Protein
			total_Carb += i.Carb
		}
	default:
		respondWithJSON(w, http.StatusInternalServerError, "meal wrong, please try again.")
	}

	respond_calculation.Fat = total_Fat
	respond_calculation.Protein = total_Protein
	respond_calculation.Carb = total_Carb
	respond_calculation.Calories = total_Fat*9 + total_Protein*4 + total_Carb*4

	respondWithJSON(w, http.StatusOK, respond_calculation)
}

func (route *App) Calculation_planner(w http.ResponseWriter, r *http.Request) {
	var total_Fat, total_Protein, total_Carb int

	respond_calculation := lib.Food_calculated{}


	for _, fod := range daily_Food_Breakfast {
		for _, ing := range daily_Ingredient_Breakfast {
			if ing.FoodId == fod.FoodId {
				total_Fat += ing.Fat * fod.Dish
				total_Protein += ing.Protein * fod.Dish
				total_Carb += ing.Carb * fod.Dish
			}
		}
	}
	for _, fod := range daily_Food_Lunch {
		for _, ing := range daily_Ingredient_Lunch {
			if ing.FoodId == fod.FoodId {
				total_Fat += ing.Fat * fod.Dish
				total_Protein += ing.Protein * fod.Dish
				total_Carb += ing.Carb * fod.Dish
			}
		}
	}
	for _, fod := range daily_Food_Dinner {
		for _, ing := range daily_Ingredient_Dinner {
			if ing.FoodId == fod.FoodId {
				total_Fat += ing.Fat * fod.Dish
				total_Protein += ing.Protein * fod.Dish
				total_Carb += ing.Carb * fod.Dish
			}
		}
	}
	for _, fod := range daily_Food_Other {
		for _, ing := range daily_Ingredient_Other {
			if ing.FoodId == fod.FoodId {
				total_Fat += ing.Fat * fod.Dish
				total_Protein += ing.Protein * fod.Dish
				total_Carb += ing.Carb * fod.Dish
			}
		}
	}


	respond_calculation.Fat = total_Fat
	respond_calculation.Protein = total_Protein
	respond_calculation.Carb = total_Carb
	respond_calculation.Calories = total_Fat*9 + total_Protein*4 + total_Carb*4

	fmt.Println(daily_Food_Breakfast)
	fmt.Println(daily_Food_Lunch)
	fmt.Println(daily_Food_Dinner)
	fmt.Println(daily_Food_Other)
	fmt.Println(respond_calculation)
	respondWithJSON(w, http.StatusOK, respond_calculation)
}

func (route *App) AddDailyMealToDB(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	userid := params["userid"]
	Plan_name := params["plan_name"]

	id := route.GenerateFoodId()
	daily_Planner.PlannerID = id
	daily_Planner.PlanName = Plan_name
	daily_Meal_Breakfast.Planner = id
	daily_Meal_Lunch.Planner = id
	daily_Meal_Dinner.Planner = id
	daily_Meal_Other.Planner = id
	daily_Planner.PlannerDate = time.Now().AddDate(0, 0, 1)
	daily_Planner.CreatedAt = time.Now()
	daily_Planner.UserID = userid

	daily_Meal_Breakfast.MealID = route.GenerateFoodId()
	daily_Meal_Lunch.MealID = route.GenerateFoodId()
	daily_Meal_Dinner.MealID = route.GenerateFoodId()
	daily_Meal_Other.MealID = route.GenerateFoodId()

	for i := range daily_Food_Breakfast {
		fmt.Println(i)
		daily_Food_Breakfast[i].Meal = daily_Meal_Breakfast.MealID
	}
	for i := range daily_Food_Lunch {
		fmt.Println(i)
		daily_Food_Lunch[i].Meal = daily_Meal_Lunch.MealID
	}
	for i := range daily_Food_Dinner {
		fmt.Println(i)
		daily_Food_Dinner[i].Meal = daily_Meal_Dinner.MealID
	}
	for i := range daily_Food_Other {
		fmt.Println(i)
		daily_Food_Other[i].Meal = daily_Meal_Other.MealID
	}

	// var output struct {
	// 	Status string `json:"status"`
	// }

	// if !(route.CheckPlannerName(Plan_name, userid)) || Plan_name == "" {
	// 	output.Status = "Failed"
	// 	respondWithJSON(w, http.StatusBadRequest, output)
	// }

	route.client.Collection("PlannerTestAPI").Add(route.ctx, daily_Planner)

	for _, i := range daily_Food_Breakfast {
		route.client.Collection("FoodTestAPI").Add(route.ctx, i)
	}
	for _, i := range daily_Food_Lunch {
		route.client.Collection("FoodTestAPI").Add(route.ctx, i)
	}
	for _, i := range daily_Food_Dinner {
		route.client.Collection("FoodTestAPI").Add(route.ctx, i)
	}
	for _, i := range daily_Food_Other {
		route.client.Collection("FoodTestAPI").Add(route.ctx, i)
	}

	for _, i := range current_Ingredient {
		route.client.Collection("IngredientTestAPI").Add(route.ctx, i)
	}
	respondWithJSON(w, http.StatusOK, "Add to existing planner success")

	route.client.Collection("MealTestAPI").Add(route.ctx, daily_Meal_Breakfast)
	route.client.Collection("MealTestAPI").Add(route.ctx, daily_Meal_Lunch)
	route.client.Collection("MealTestAPI").Add(route.ctx, daily_Meal_Dinner)
	route.client.Collection("MealTestAPI").Add(route.ctx, daily_Meal_Other)

}

func (route *App) CreateIngredientUsed(w http.ResponseWriter, r *http.Request) []models.Ingredient {
	IngredientsData := []models.Ingredient{}
	IngredientData := models.Ingredient{}
	for index, f := range FoodDB {
		iter := route.client.Collection("Ingredient").Where("FoodId", "==", f.FoodId).Documents(route.ctx)
		fid := route.GenerateFoodId()
		for {
			IngredientData = models.Ingredient{}
			doc, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				log.Fatalf("Failed to iterate: %v", err)
			}

			mapstructure.Decode(doc.Data(), &IngredientData)

			//change foodid in Ingredient struct to new foodid
			IngredientData.FoodId = fid
			IngredientsData = append(IngredientsData, IngredientData)

		}
		FoodDB[index].FoodId = fid

	}
	return IngredientsData
}

func CheckValid(gender string, weight_management string, weightdef int, weightdiff int, height int, age int, activity_level string, additional_opts []string, time string) bool {
	if age >= 150 || age <= 0 || height >= 300 || height <= 0 || weightdef >= 300 || weightdef <= 0 || weightdiff >= 300 || weightdiff <= 0 {
		return false
	}
	switch weight_management {
	case "Gain":
		if (weightdiff - weightdef) < 0 {
			return false
		}
	case "Loss":
		if (weightdef - weightdiff) < 0 {
			return false
		}
	case "Stable":
		if (weightdef - weightdiff) != 0 {
			return false
		}
	}
	return true
}

func (route *App) FoodRecommendationView(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, PlannersData)
}

func (route *App) FoodRecommendation(w http.ResponseWriter, r *http.Request) {
	listOfFood := []models.Food{}
	type Response struct {
		IsValid bool `json:"isValid"`
	}

	var Userdetail struct {
		Gender           string   `json:"gender"`
		WeightManagement string   `json:"weight_management"`
		WeightDef        int      `json:"weightdef"`
		WeightDiff       int      `json:"weightdiff"`
		Height           int      `json:"height"`
		Age              int      `json:"age"`
		ActivityLevel    string   `json:"activity_level"`
		AdditionalOpts   []string `json:"additional_opts"`
		Time             string   `json:"time"`
	}

	//check Userdetail

	//import data from DB
	iter := route.client.Collection("Food").Documents(route.ctx)
	for {
		Food := models.Food{}
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			respondWithJSON(w, http.StatusInternalServerError, "Something wrong, please try again.")
		}

		mapstructure.Decode(doc.Data(), &Food)
		listOfFood = append(listOfFood, Food)

	}

	// Decode request body
	Decoder := json.NewDecoder(r.Body)

	err := Decoder.Decode(&Userdetail)

	if err != nil {
		log.Printf("error: %s", err)
	}

	if !(CheckValid(Userdetail.Gender, Userdetail.WeightManagement, Userdetail.WeightDef, Userdetail.WeightDiff, Userdetail.Height, Userdetail.Age, Userdetail.ActivityLevel, Userdetail.AdditionalOpts, Userdetail.Time)) {
		respone := Response{
			IsValid: false,
		}
		respondWithJSON(w, http.StatusBadRequest, respone)
	} else {

		c, t := lib.CaloriePerDay(Userdetail.Gender, Userdetail.WeightManagement, Userdetail.WeightDef, Userdetail.WeightDiff, Userdetail.Height, Userdetail.Age, Userdetail.ActivityLevel, Userdetail.AdditionalOpts, Userdetail.Time)
		// fmt.Printf("value : %d Time: %d  \n", c, t)
		// status, F := lib.FoodRecommendation(c, t, Userdetail.AdditionalOpts, listOfFood)
		status, Fdb, mealdb, planner, plannerD := lib.FoodRecommendation(c, t, Userdetail.AdditionalOpts, listOfFood)
		// fmt.Printf("Len of Food List: %d 	Status: %s\n", len(Fdb), status)

		//แสดงผล list ของ planner
		/*fmt.Printf("Len of Planner: %d 	Status: %s\n", len(planner), status)
		for _, food := range planner {
			fmt.Println(food)
		}
		fmt.Printf("Len of Meal: %d 	Status: %s\n", len(mealdb), status)
		for _, food := range mealdb {
			fmt.Println(food)
		}*/
		// fmt.Printf("Len of Food: %d 	Status: %s\n", len(Fdb), status)
		// for _, food := range Fdb {
		// 	fmt.Println(food)
		// }

		if status == "OK" {
			PlannerDB = planner
			MealDB = mealdb
			FoodDB = Fdb

			PlannersData = plannerD
			IngredientDB = route.CreateIngredientUsed(w, r)
			// fmt.Printf("Len of Food: %d 	Status: %s\n", len(FoodDB), status)
			// for _, food := range FoodDB {
			// 	fmt.Println(food)
			// }

			// fmt.Printf("Len of Ingredient: %d 	Status: %s\n", len(IngredientDB), status)
			// for _, food := range IngredientDB {
			// 	fmt.Println(food)
			// }
			if status == "OK" {
				s = true
			} else {
				s = false
			}

			response := Response{
				IsValid: s,
			}

			respondWithJSON(w, http.StatusOK, response)
		} else {
			response := Response{
				IsValid: s,
			}
			respondWithJSON(w, http.StatusOK, response)
		}
	}

}

func (route *App) CheckPlannerName(plannerName string, uid string) bool {
	status := true
	PlannersData := []models.Planner{}
	iter := route.client.Collection("PlannerTestAPI").Where("UserID", "==", uid).Documents(route.ctx)
	for {
		PlannerData := models.Planner{}
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Failed to iterate: %v", err)
		}

		mapstructure.Decode(doc.Data(), &PlannerData)
		PlannersData = append(PlannersData, PlannerData)
	}

	for _, p := range PlannersData {
		if p.PlanName == plannerName {
			status = false
			break
		}
	}
	return status
}

func (route *App) AddtoPlanner(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Name string `json:"PlannerName"`
	}

	var output struct {
		Status string `json:"status"`
	}

	params := mux.Vars(r)
	paramsID := params["id"]

	Decoder := json.NewDecoder(r.Body)

	err := Decoder.Decode(&payload)

	if err != nil {
		log.Printf("error: %s", err)
	}

	//check Duplicate Plnner name
	if !(route.CheckPlannerName(payload.Name, paramsID)) || payload.Name == "" {
		output.Status = "Failed"
		respondWithJSON(w, http.StatusBadRequest, output)
	} else {

		/*------ add Planner to database ------ */
		for index, p := range PlannerDB {

			//insert userID to planner when user add to planner
			p.UserID = paramsID
			p.PlanName = payload.Name

			PlannersData[index].UserID = p.UserID
			PlannersData[index].PlanName = p.PlanName

			_, _, err = route.client.Collection("PlannerTestAPI").Add(route.ctx, p)

			if err != nil {
				log.Printf("An error has occurred: %s", err)
			}
		}
		/*------ end add Planner to databse ------ */

		/*------ add Meal to databse ------ */
		for _, m := range MealDB {

			//insert userID to mlanner when user add to planner
			_, _, err = route.client.Collection("MealTestAPI").Add(route.ctx, m)

			if err != nil {
				log.Printf("An error has occurred: %s", err)
			}
		}
		/*------ end add Meal to databse ------ */

		/*------ add Food to databse ------ */
		for _, f := range FoodDB {

			//insert userID to planner when user add to planner
			_, _, err = route.client.Collection("FoodTestAPI").Add(route.ctx, f)

			if err != nil {
				log.Printf("An error has occurred: %s", err)
			}
		}
		/*------ end add Food to databse ------ */

		/*------ add Ingredient to databse ------ */
		for _, i := range IngredientDB {

			//insert userID to planner when user add to planner
			_, _, err = route.client.Collection("IngredientTestAPI").Add(route.ctx, i)

			if err != nil {
				log.Printf("An error has occurred: %s", err)
			}
		}
		/*------ end add Ingredient to databse ------ */

		output.Status = "OK"
		respondWithJSON(w, http.StatusOK, output)
	}
}

func (route *App) FindListMealId(pid string) []string {
	MealIDs := []string{}
	count := 0

	for _, p := range MealDB {
		if p.Planner == pid {
			if count == 0 {
				count += 1
			}
			MealIDs = append(MealIDs, p.MealID)
		} else if p.Planner != pid && count != 0 {
			break
		}
	}

	return MealIDs
}

func (route *App) NutrientsProportion(fid string) {
	for _, i := range IngredientDB {
		if i.FoodId == fid {
			carbPerDay += i.Carb
			proteinPerDay += i.Protein
			fatPerDay += i.Fat
		}
	}
}

func (route *App) FindListFood(mid string, index int) {
	count := 0
	m := ""
	switch index {
	case 0:
		m = "Breakfast"
	case 1:
		m = "Lunch"
	case 2:
		m = "Dinner"
	case 3:
		m = "Other"
	}

	for _, f := range FoodDB {
		if f.Meal == mid {
			display := models.FoodDisplay{}
			if count == 0 {
				count += 1
			}
			display.Calories = f.Calories
			display.Dish = f.Dish
			display.ImageUrl = f.ImageUrl
			display.FoodId = f.FoodId
			// display.MealID = f.Meal
			display.Meal = m
			display.Name = f.Name
			// display.Tags = f.Tags
			FoodDisplay = append(FoodDisplay, display)

			//calculate Proportion
			route.NutrientsProportion(f.FoodId)
			calperDay += display.Calories
		} else if f.Meal != mid && count != 0 {
			break
		}
	}
}

func (route *App) ShowPlanner(w http.ResponseWriter, r *http.Request) {
	/*
		1. เอา planner id
		2. ไป หา mealID แต่ละอัน
		3. เอา mealID ไปดึงหาอาหารตัวนั้นออกมาแล้ว append ลง list
		4. ส่ง list ของอาหารขึ้นไป
	*/
	type ProportionResp struct {
		Protein  int `json:"protein"`
		Carb     int `json:"carb"`
		Fat      int `json:"fat"`
		Calories int `json:"calories"`
	}

	type Response struct {
		Proportion ProportionResp       `json:"message"`
		Data       []models.FoodDisplay `json:"data"` // เก็บข้อมูลเป็น slice ของ struct
	}

	// Get
	params := mux.Vars(r)
	paramsID := params["pid"]

	listMealID := route.FindListMealId(paramsID)
	// fmt.Println(listMealID)

	for index, m := range listMealID {
		route.FindListFood(m, index)
	}

	//clear values when finished

	foodList := FoodDisplay
	FoodDisplay = nil

	proportion := ProportionResp{
		Protein:  proteinPerDay,
		Carb:     carbPerDay,
		Fat:      fatPerDay,
		Calories: calperDay,
	}

	response := Response{
		Proportion: proportion,
		Data:       foodList,
	}

	proteinPerDay = 0
	carbPerDay = 0
	fatPerDay = 0
	calperDay = 0

	respondWithJSON(w, http.StatusOK, response)
}

func (route *App) GetPlannerByUserID(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["userID"]
	plannerGroups := make(map[string]map[string]interface{})

	currentTime := time.Now()
	log.Printf("Fetching planners for userID: %s from current date: %v", userID, currentTime)

	iter := route.client.Collection("PlannerTestAPI").
		Where("UserID", "==", userID).
		Where("PlannerDate", ">=", currentTime).
		Documents(route.ctx)

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("Error fetching document: %s", err)
			respondWithError(w, http.StatusInternalServerError, "Something went wrong, please try again.")
			return
		}

		var plannerData models.Planner
		if err := mapstructure.Decode(doc.Data(), &plannerData); err != nil {
			log.Printf("Error decoding planner data for document %s: %s", doc.Ref.ID, err)
			respondWithError(w, http.StatusInternalServerError, "Error decoding planner data.")
			return
		}

		group, exists := plannerGroups[plannerData.PlanName]
		if !exists {

			group = map[string]interface{}{
				"planName":  plannerData.PlanName,
				"firstDate": plannerData.PlannerDate,
				"lastDate":  plannerData.PlannerDate,
			}
			plannerGroups[plannerData.PlanName] = group
		} else {

			firstDate := group["firstDate"].(time.Time)
			lastDate := group["lastDate"].(time.Time)

			if plannerData.PlannerDate.Before(firstDate) {
				group["firstDate"] = plannerData.PlannerDate
			}
			if plannerData.PlannerDate.After(lastDate) {
				group["lastDate"] = plannerData.PlannerDate
			}
		}
	}

	var groupedPlannersResponse []map[string]interface{}
	for _, group := range plannerGroups {
		groupedPlannersResponse = append(groupedPlannersResponse, group)
	}

	if len(groupedPlannersResponse) == 0 {
		log.Println("No planners found for userID:", userID)
		respondWithError(w, http.StatusNotFound, "No planners found.")
		return
	}

	respondWithJSON(w, http.StatusOK, groupedPlannersResponse)
}

func (route *App) GetPlannerByPlanName(w http.ResponseWriter, r *http.Request) {
	planName := mux.Vars(r)["planName"]
	currentTime := time.Now()
	log.Printf("Fetching planners for planName: %s from current date: %v", planName, currentTime)

	iter := route.client.Collection("PlannerTestAPI").
		Where("PlanName", "==", planName).
		Where("PlannerDate", ">=", currentTime).
		Documents(route.ctx)

	var plannersResponse []map[string]interface{}
	var plannerDates []time.Time

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("Error fetching document: %s", err)
			respondWithError(w, http.StatusInternalServerError, "Something went wrong, please try again.")
			return
		}

		var plannerData models.Planner
		if err := mapstructure.Decode(doc.Data(), &plannerData); err != nil {
			log.Printf("Error decoding planner data for document %s: %s", doc.Ref.ID, err)
			respondWithError(w, http.StatusInternalServerError, "Error decoding planner data.")
			return
		}

		plannerDates = append(plannerDates, plannerData.PlannerDate)

		plannerTotalCalories, plannerTotalProtein, plannerTotalCarb, plannerTotalFat := 0, 0, 0, 0

		mealIter := route.client.Collection("MealTestAPI").
			Where("Planner", "==", plannerData.PlannerID).
			Documents(route.ctx)

		for {
			mealDoc, err := mealIter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, "Error retrieving meals")
				return
			}

			mealData := models.Meal{}
			mapstructure.Decode(mealDoc.Data(), &mealData)

			foodIter := route.client.Collection("FoodTestAPI").
				Where("Meal", "==", mealData.MealID).
				Documents(route.ctx)

			for {
				foodDoc, err := foodIter.Next()
				if err == iterator.Done {
					break
				}
				if err != nil {
					respondWithError(w, http.StatusInternalServerError, "Error retrieving food data")
					return
				}

				foodData := models.Food{}
				mapstructure.Decode(foodDoc.Data(), &foodData)

				ingredientIter := route.client.Collection("IngredientTestAPI").
					Where("FoodId", "==", foodData.FoodId).
					Documents(route.ctx)

				for {
					ingredientDoc, err := ingredientIter.Next()
					if err == iterator.Done {
						break
					}
					if err != nil {
						respondWithError(w, http.StatusInternalServerError, "Error retrieving ingredients")
						return
					}

					ingredientData := models.Ingredient{}
					mapstructure.Decode(ingredientDoc.Data(), &ingredientData)

					plannerTotalProtein += ingredientData.Protein
					plannerTotalCarb += ingredientData.Carb
					plannerTotalFat += ingredientData.Fat
				}

				plannerTotalCalories += foodData.Calories
			}
		}

		plannerResponse := map[string]interface{}{
			"plannerID":     plannerData.PlannerID,
			"plannerDate":   plannerData.PlannerDate,
			"totalCalories": plannerTotalCalories,
			"totalProtein":  plannerTotalProtein,
			"totalCarb":     plannerTotalCarb,
			"totalFat":      plannerTotalFat,
		}

		plannersResponse = append(plannersResponse, plannerResponse)
	}

	if len(plannerDates) == 0 {
		log.Println("No planners found for planName:", planName)
		respondWithError(w, http.StatusNotFound, "No planners found.")
		return
	}

	sort.Slice(plannerDates, func(i, j int) bool {
		return plannerDates[i].Before(plannerDates[j])
	})

	firstDate := plannerDates[0]
	lastDate := plannerDates[len(plannerDates)-1]

	finalResponse := map[string]interface{}{
		"planName":  planName,
		"firstDate": firstDate,
		"lastDate":  lastDate,
		"planners":  plannersResponse,
	}

	respondWithJSON(w, http.StatusOK, finalResponse)
}

func (route *App) GetDetailFromPlannerID(w http.ResponseWriter, r *http.Request) {
	plannerID := mux.Vars(r)["plannerID"]
	log.Printf("Fetching meals and foods for plannerID: %s", plannerID)

	plannerQuery := route.client.Collection("PlannerTestAPI").Where("PlannerID", "==", plannerID).Limit(1).Documents(route.ctx)
	plannerDoc, err := plannerQuery.Next()
	if err == iterator.Done {
		log.Printf("Planner with PlannerID %s not found.", plannerID)
		respondWithError(w, http.StatusNotFound, "Planner not found.")
		return
	} else if err != nil {
		log.Printf("Error retrieving planner: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Error retrieving planner.")
		return
	}

	var plannerData map[string]interface{}
	if err := mapstructure.Decode(plannerDoc.Data(), &plannerData); err != nil {
		log.Printf("Error decoding planner data: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Error decoding planner data.")
		return
	}
	planName, _ := plannerData["PlanName"].(string)
	plannerDate, _ := plannerData["PlannerDate"].(time.Time)

	mealsMap := make(map[string]map[string]interface{})
	mealIter := route.client.Collection("MealTestAPI").Where("Planner", "==", plannerID).Documents(route.ctx)

	totalAllCalories, totalAllCarbs, totalAllProtein, totalAllFat := 0, 0, 0, 0

	for {
		mealDoc, err := mealIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error retrieving meals.")
			return
		}

		var mealData models.Meal
		if err := mapstructure.Decode(mealDoc.Data(), &mealData); err != nil {
			log.Printf("Error decoding meal data: %s", err)
			respondWithError(w, http.StatusInternalServerError, "Error decoding meal data.")
			return
		}
		log.Printf("Processing meal: %s", mealData.Meal)

		mealResponse, exists := mealsMap[mealData.Meal]
		if !exists {
			mealResponse = map[string]interface{}{
				"meal":          mealData.Meal,
				"foods":         []map[string]interface{}{},
				"totalCalories": 0,
			}
			mealsMap[mealData.Meal] = mealResponse
		}

		foodIter := route.client.Collection("FoodTestAPI").Where("Meal", "==", mealData.MealID).Documents(route.ctx)
		mealTotalCalories := 0

		for {
			foodDoc, err := foodIter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, "Error retrieving foods.")
				return
			}

			var foodData models.Food
			if err := mapstructure.Decode(foodDoc.Data(), &foodData); err != nil {
				log.Printf("Error decoding food data: %s", err)
				respondWithError(w, http.StatusInternalServerError, "Error decoding food data.")
				return
			}
			log.Printf("Adding food: %s, Dish: %d, Calories: %d", foodData.Name, foodData.Dish, foodData.Calories)
			tags := foodData.Tags
			imageURL := foodData.ImageUrl
			mealID := foodData.Meal
			foodID := foodData.FoodId

			ingredientIter := route.client.Collection("IngredientTestAPI").Where("FoodId", "==", foodData.FoodId).Documents(route.ctx)

			foodProtein, foodFat, foodCarb := 0, 0, 0

			for {
				ingredientDoc, err := ingredientIter.Next()
				if err == iterator.Done {
					break
				}
				if err != nil {
					respondWithError(w, http.StatusInternalServerError, "Error retrieving ingredients.")
					return
				}

				var ingredientData models.Ingredient
				if err := mapstructure.Decode(ingredientDoc.Data(), &ingredientData); err != nil {
					log.Printf("Error decoding ingredient data: %s", err)
					respondWithError(w, http.StatusInternalServerError, "Error decoding ingredient data.")
					return
				}

				foodProtein += ingredientData.Protein
				foodFat += ingredientData.Fat
				foodCarb += ingredientData.Carb
			}

			foodInfo := map[string]interface{}{
				"name":     foodData.Name,
				"dish":     foodData.Dish,
				"calories": foodData.Calories,
				"protein":  foodProtein,
				"fat":      foodFat,
				"carb":     foodCarb,
				"imageURL": imageURL,
				"tags":     tags,
				"mealID":   mealID,
				"foodID":   foodID,
			}

			foods := mealResponse["foods"].([]map[string]interface{})
			foods = append(foods, foodInfo)
			mealResponse["foods"] = foods

			mealTotalCalories += foodData.Calories
			totalAllCalories += foodData.Calories
			totalAllCarbs += foodCarb
			totalAllProtein += foodProtein
			totalAllFat += foodFat
		}

		mealResponse["totalCalories"] = mealResponse["totalCalories"].(int) + mealTotalCalories
		log.Printf("Updated total calories for meal %s: %d", mealData.Meal, mealResponse["totalCalories"])
	}

	var mealsResponse []map[string]interface{}
	for _, meal := range mealsMap {
		mealsResponse = append(mealsResponse, meal)
	}

	mealOrder := map[string]int{
		"Breakfast": 1,
		"Lunch":     2,
		"Dinner":    3,
		"Other":     4,
	}

	sort.Slice(mealsResponse, func(i, j int) bool {
		return mealOrder[mealsResponse[i]["meal"].(string)] < mealOrder[mealsResponse[j]["meal"].(string)]
	})

	if len(mealsResponse) == 0 {
		log.Println("No meals found for plannerID:", plannerID)
		respondWithError(w, http.StatusNotFound, "No meals found.")
		return
	}

	if len(mealsResponse) == 0 {
		log.Println("No meals found for plannerID:", plannerID)
		respondWithError(w, http.StatusNotFound, "No meals found.")
		return
	}

	// Add overall totals to the response
	finalResponse := map[string]interface{}{
		"plannerID":        plannerID,
		"planName":         planName,
		"plannerDate":      plannerDate,
		"meals":            mealsResponse,
		"totalAllCalories": totalAllCalories,
		"totalAllCarbs":    totalAllCarbs,
		"totalAllProtein":  totalAllProtein,
		"totalAllFat":      totalAllFat,
	}

	respondWithJSON(w, http.StatusOK, finalResponse)
}

func (route *App) EditFoodDish(w http.ResponseWriter, r *http.Request) {
	foodID := mux.Vars(r)["foodID"]
	log.Printf("Attempting to edit dish for FoodID: %s", foodID)

	var requestBody struct {
		Dish int `json:"dish"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		log.Printf("Error decoding request body: %s", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request body.")
		return
	}

	foodIter := route.client.Collection("FoodTestAPI").Where("FoodId", "==", foodID).Documents(route.ctx)

	var foodData models.Food
	foodFound := false

	for {
		foodDoc, err := foodIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error retrieving food.")
			return
		}

		if err := mapstructure.Decode(foodDoc.Data(), &foodData); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error decoding food data.")
			return
		}

		foodFound = true

		if foodData.Dish > 0 {
			caloriesPerDish := float64(foodData.Calories) / float64(foodData.Dish)
			newCalories := int(caloriesPerDish * float64(requestBody.Dish))

			foodData.Dish = requestBody.Dish
			foodData.Calories = newCalories

			_, err := route.client.Collection("FoodTestAPI").Doc(foodDoc.Ref.ID).Set(route.ctx, foodData)
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, "Error updating food.")
				return
			}

			respondWithJSON(w, http.StatusOK, map[string]interface{}{
				"message": "Food dish updated successfully.",
				"food":    foodData,
			})
			return
		} else {
			respondWithError(w, http.StatusBadRequest, "Invalid dish count. Cannot be zero.")
			return
		}
	}

	if !foodFound {
		respondWithError(w, http.StatusNotFound, "Food not found.")
		return
	}
}

// DeletePlanner deletes a planner by plannerID
func (route *App) DeletePlannerByID(w http.ResponseWriter, r *http.Request) {
	plannerID := mux.Vars(r)["plannerID"]
	log.Printf("Attempting to delete planner with PlannerID: %s", plannerID)

	iter := route.client.Collection("PlannerTestAPI").
		Where("PlannerID", "==", plannerID).
		Documents(route.ctx)

	plannerExists := false
	var plannerDoc *firestore.DocumentRef

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error retrieving planner.")
			return
		}

		plannerExists = true
		plannerDoc = doc.Ref
		break
	}

	if !plannerExists {
		log.Printf("No planner found with PlannerID: %s", plannerID)
		respondWithError(w, http.StatusNotFound, "Planner not found.")
		return
	}

	_, err := plannerDoc.Delete(route.ctx)
	if err != nil {
		log.Printf("Error deleting planner with PlannerID: %s, %v", plannerID, err)
		respondWithError(w, http.StatusInternalServerError, "Error deleting planner.")
		return
	}

	log.Printf("Successfully deleted planner with PlannerID: %s", plannerID)
	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Planner deleted successfully."})
}

func (route *App) DeletePlannerFromPlanName(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["planName"]
	log.Printf("Attempting to delete planners with name: %s", name)

	plannerIter := route.client.Collection("PlannerTestAPI").Where("PlanName", "==", name).Documents(route.ctx)

	plannersDeleted := 0

	for {
		plannerDoc, err := plannerIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error retrieving planners.")
			return
		}

		_, err = route.client.Collection("PlannerTestAPI").Doc(plannerDoc.Ref.ID).Delete(route.ctx)
		if err != nil {
			log.Printf("Error deleting planner: %s", err)
			respondWithError(w, http.StatusInternalServerError, "Error deleting planner.")
			return
		}

		plannersDeleted++
		log.Printf("Deleted planner with ID: %s", plannerDoc.Ref.ID)
	}

	if plannersDeleted == 0 {
		respondWithError(w, http.StatusNotFound, "No planners found with that name.")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": fmt.Sprintf("%d planners deleted successfully.", plannersDeleted),
	})
}

// DeleteFood deletes food by foodID
func (route *App) DeleteFoodByID(w http.ResponseWriter, r *http.Request) {
	foodID := mux.Vars(r)["foodID"]
	log.Printf("Attempting to delete food with FoodID: %s", foodID)

	iter := route.client.Collection("FoodTestAPI").Where("FoodId", "==", foodID).Documents(route.ctx)

	foodExists := false
	var foodDoc *firestore.DocumentRef

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error retrieving food.")
			return
		}

		foodExists = true
		foodDoc = doc.Ref
		break
	}

	if !foodExists {
		log.Printf("No food found with FoodID: %s", foodID)
		respondWithError(w, http.StatusNotFound, "Food not found.")
		return
	}

	_, err := foodDoc.Delete(route.ctx)
	if err != nil {
		log.Printf("Error deleting food with FoodID: %s, %v", foodID, err)
		respondWithError(w, http.StatusInternalServerError, "Error deleting food.")
		return
	}

	log.Printf("Successfully deleted food with FoodID: %s", foodID)
	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Food deleted successfully."})
}

func (route *App) GetUser(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["userID"]
	log.Printf("Fetching user information for UserID: %s", userID)

	// Query the correct collection name (check if it's "User" or "Users")
	iter := route.client.Collection("User").Where("UserID", "==", userID).Documents(route.ctx)
	doc, err := iter.Next()
	if err == iterator.Done {
		log.Printf("No user found with UserID: %s", userID)
		respondWithError(w, http.StatusNotFound, "User not found")
		return
	}
	if err != nil {
		log.Printf("Error retrieving user: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Error retrieving user")
		return
	}

	var user models.Users
	err = doc.DataTo(&user)
	if err != nil {
		log.Printf("Error decoding user data: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Error decoding user data")
		return
	}

	respondWithJSON(w, http.StatusOK, user)
}
func (route *App) EditUser(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["userID"]
	log.Printf("Editing user information for UserID: %s", userID)

	// Query the correct collection
	iter := route.client.Collection("User").Where("UserID", "==", userID).Documents(route.ctx)
	doc, err := iter.Next()
	if err == iterator.Done {
		log.Printf("No user found with UserID: %s", userID)
		respondWithError(w, http.StatusNotFound, "User not found")
		return
	}
	if err != nil {
		log.Printf("Error retrieving user: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Error retrieving user")
		return
	}

	// Decode the request body for updated user information
	var updatedUser models.Users
	err = json.NewDecoder(r.Body).Decode(&updatedUser)
	if err != nil {
		log.Printf("Error decoding request body: %v", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Prepare the slice of Firestore update operations
	var updates []firestore.Update
	if updatedUser.Username != "" {
		updates = append(updates, firestore.Update{Path: "Username", Value: updatedUser.Username})
	}
	if updatedUser.Email != "" {
		updates = append(updates, firestore.Update{Path: "Email", Value: updatedUser.Email})
	}
	if updatedUser.Password != "" {
		updates = append(updates, firestore.Update{Path: "Password", Value: updatedUser.Password})
	}
	if updatedUser.Description != "" {
		updates = append(updates, firestore.Update{Path: "Description", Value: updatedUser.Description})
	}
	if updatedUser.ProfileImage != "" {
		updates = append(updates, firestore.Update{Path: "ProfileImage", Value: updatedUser.ProfileImage})
	}
	if updatedUser.Age != 0 {
		updates = append(updates, firestore.Update{Path: "Age", Value: updatedUser.Age})
	}
	if updatedUser.Height != 0 {
		updates = append(updates, firestore.Update{Path: "Height", Value: updatedUser.Height})
	}
	if updatedUser.Weight != 0 {
		updates = append(updates, firestore.Update{Path: "Weight", Value: updatedUser.Weight})
	}
	if updatedUser.Gender != "" {
		updates = append(updates, firestore.Update{Path: "Gender", Value: updatedUser.Gender})
	}

	// Update the document in Firestore
	_, err = doc.Ref.Update(route.ctx, updates)
	if err != nil {
		log.Printf("Error updating user: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Error updating user")
		return
	}

	log.Printf("User with UserID: %s successfully updated", userID)
	respondWithJSON(w, http.StatusOK, map[string]string{"message": "User updated successfully"})
}

func (route *App) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["userID"]
	log.Printf("Attempting to delete user with UserID: %s", userID)

	// Query the correct collection to get the document reference
	iter := route.client.Collection("User").Where("UserID", "==", userID).Documents(route.ctx)
	doc, err := iter.Next()
	if err == iterator.Done {
		log.Printf("No user found with UserID: %s", userID)
		respondWithError(w, http.StatusNotFound, "User not found")
		return
	}
	if err != nil {
		log.Printf("Error retrieving user: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Error retrieving user")
		return
	}

	// Delete the user document from Firestore
	_, err = doc.Ref.Delete(route.ctx)
	if err != nil {
		log.Printf("Error deleting user: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Error deleting user")
		return
	}

	log.Printf("User with UserID: %s successfully deleted", userID)
	respondWithJSON(w, http.StatusOK, map[string]string{"message": "User deleted successfully"})
}

func initFirebaseApp() *firebase.App {
	opt := option.WithCredentialsFile("appsoftdev-d4314-firebase-adminsdk-vyv9x-3327a98e65.json")
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("error initializing firebase app: %v", err)
	}
	return app
}
