package account

import (
	"cane-project/database"
	"cane-project/jwt"
	"cane-project/model"
	"cane-project/util"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mongodb/mongo-go-driver/mongo/options"

	"github.com/fatih/structs"
	"github.com/go-chi/chi"
	"github.com/mitchellh/mapstructure"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
)

var mySigningKey = []byte("secret")

var privilegeLevel = map[string]int{
	"admin":    1,
	"user":     2,
	"readonly": 3,
}

// Login Function
func Login(w http.ResponseWriter, r *http.Request) {
	var account model.UserAccount
	var login map[string]interface{}

	json.NewDecoder(r.Body).Decode(&login)

	filter := primitive.M{
		"username": login["username"],
	}

	foundVal, _ := database.FindOne("accounts", "users", filter)
	mapstructure.Decode(foundVal, &account)

	if account.Password == login["password"] {
		util.RespondwithJSON(w, http.StatusOK, structs.Map(account))
	} else {
		util.RespondWithError(w, http.StatusBadRequest, "invalid login")
	}
}

// GetUser Function
func GetUser(w http.ResponseWriter, r *http.Request) {
	var account model.UserAccount
	// var opts options.FindOptions

	filter := primitive.M{
		"username": chi.URLParam(r, "username"),
	}

	// projection := primitive.M{
	// 	"_id":      0,
	// 	"username": 1,
	// }

	// opts.SetProjection(projection)

	foundVal, _ := database.FindOne("accounts", "users", filter)

	// for _, user := range accounts {
	// 	accountList = append(accountList, user["username"].(string))
	// }

	mapstructure.Decode(foundVal, &account)

	util.RespondwithJSON(w, http.StatusOK, account)
}

// GetUsers Function
func GetUsers(w http.ResponseWriter, r *http.Request) {
	// var accounts []primitive.M
	var opts options.FindOptions
	var accountList []string

	projection := primitive.M{
		"_id":      0,
		"username": 1,
	}

	opts.SetProjection(projection)

	foundVals, _ := database.FindAll("accounts", "users", primitive.M{}, opts)

	for _, user := range foundVals {
		accountList = append(accountList, user["username"].(string))
	}

	util.RespondwithJSON(w, http.StatusOK, map[string][]string{"users": accountList})
}

// CreateUser Function
func CreateUser(w http.ResponseWriter, r *http.Request) {
	var target model.UserAccount

	json.NewDecoder(r.Body).Decode(&target)

	if UserExists(target.UserName) {
		util.RespondWithError(w, http.StatusBadRequest, "username already exists")
		return
	}

	target.Token, _ = jwt.GenerateJWT(target)

	userID, saveErr := database.Save("accounts", "users", target)

	if saveErr != nil {
		fmt.Println(saveErr)
		util.RespondWithError(w, http.StatusBadRequest, "error saving account")
		return
	}

	filter := primitive.M{
		"_id": userID.(primitive.ObjectID),
	}

	userVal, _ := database.FindOne("accounts", "users", filter)
	mapstructure.Decode(userVal, &target)

	util.RespondwithJSON(w, http.StatusCreated, target)
}

// DeleteUser Function
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	filter := primitive.M{
		"username": chi.URLParam(r, "username"),
	}

	deleteErr := database.Delete("accounts", "users", filter)

	if deleteErr != nil {
		fmt.Println(deleteErr)
		util.RespondWithError(w, http.StatusBadRequest, "user not found")
		return
	}

	util.RespondwithJSON(w, http.StatusOK, nil)
}

// UserExists Function
func UserExists(username string) bool {
	filter := primitive.M{
		"username": username,
	}

	_, findErr := database.FindOne("accounts", "users", filter)

	if findErr == nil {
		return true
	}

	return false
}

// UpdateUser Function
func UpdateUser(w http.ResponseWriter, r *http.Request) {
	var userDetails map[string]interface{}
	var updatedUser model.UserAccount

	json.NewDecoder(r.Body).Decode(&userDetails)

	filter := primitive.M{
		"username": chi.URLParam(r, "username"),
	}

	findVal, findErr := database.FindOne("accounts", "users", filter)

	if findErr != nil {
		fmt.Println(findErr)
		util.RespondWithError(w, http.StatusBadRequest, "user not found")
		return
	}

	mapstructure.Decode(findVal, &updatedUser)

	updatedUser.FirstName = userDetails["fname"].(string)
	updatedUser.LastName = userDetails["lname"].(string)
	updatedUser.Password = userDetails["password"].(string)
	updatedUser.Privilege = int(userDetails["privilege"].(float64))
	updatedUser.Enable = userDetails["enable"].(bool)

	_, replaceErr := database.ReplaceOne("accounts", "users", filter, structs.Map(updatedUser))

	if replaceErr != nil {
		fmt.Println(replaceErr)
		util.RespondWithError(w, http.StatusBadRequest, "error updating user")
		return
	}

	util.RespondwithJSON(w, http.StatusOK, updatedUser)
}

// ValidateUserToken Function
func ValidateUserToken(w http.ResponseWriter, r *http.Request) {
	var account model.UserAccount

	json.NewDecoder(r.Body).Decode(&account)

	filter := primitive.M{
		"username": account.UserName,
	}

	foundVal, findErr := database.FindOne("accounts", "users", filter)

	if findErr != nil {
		fmt.Println(findErr)
		util.RespondWithError(w, http.StatusBadRequest, "invalid username")
		return
	}

	mapstructure.Decode(foundVal, &account)

	jwt.ValidateJWT(account.Token)
}

// RefreshToken Function
func RefreshToken(w http.ResponseWriter, r *http.Request) {
	var account model.UserAccount

	filter := primitive.M{
		"username": chi.URLParam(r, "user"),
	}

	findVal, findErr := database.FindOne("accounts", "users", filter)

	if findErr != nil {
		fmt.Println(findErr)
		util.RespondWithError(w, http.StatusBadRequest, "invalid account")
		return
	}

	mapstructure.Decode(findVal, &account)

	newToken, _ := jwt.GenerateJWT(account)

	update := primitive.M{
		"$set": primitive.M{
			"token": newToken,
		},
	}

	updateVal, updateErr := database.FindAndUpdate("accounts", "users", filter, update)

	if updateErr != nil {
		fmt.Println(updateErr)
		util.RespondWithError(w, http.StatusBadRequest, "token refresh failed")
		return
	}

	util.RespondwithJSON(w, http.StatusCreated, updateVal)
}
