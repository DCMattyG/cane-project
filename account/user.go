package account

import (
	"cane-project/auth"
	"cane-project/database"
	"cane-project/model"
	"cane-project/util"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/fatih/structs"
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

	bodyReader, err := ioutil.ReadAll(r.Body)

	if err != nil {
		fmt.Println(err)
		util.RespondWithError(w, http.StatusBadRequest, "invalid data")
		return
	}

	err = util.UnmarshalJSON(bodyReader, &login)

	if err != nil {
		fmt.Println(err)
		util.RespondWithError(w, http.StatusBadRequest, "unmarshall failed")
		return
	}

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

// AddUser Function
func AddUser(w http.ResponseWriter, r *http.Request) {
	var target model.UserAccount

	bodyReader, err := ioutil.ReadAll(r.Body)

	if err != nil {
		fmt.Println(err)
		util.RespondWithError(w, http.StatusBadRequest, "invalid data")
		return
	}

	err = util.UnmarshalJSON(bodyReader, &target)

	if err != nil {
		fmt.Println(err)
		util.RespondWithError(w, http.StatusBadRequest, "unmarshall failed")
		return
	}

	filter := primitive.M{
		"username": target.UserName,
	}

	_, findErr := database.FindOne("accounts", "users", filter)

	if findErr == nil {
		fmt.Println(findErr)
		util.RespondWithError(w, http.StatusBadRequest, "existing account")
		return
	}

	target.Token, err = auth.GenerateJWT(target)

	userID, _ := database.Save("accounts", "users", target)

	fmt.Print("Inserted ID: ")
	fmt.Println(userID)

	util.RespondwithJSON(w, http.StatusCreated, structs.Map(target))
}

// ValidateUserToken Function
func ValidateUserToken(w http.ResponseWriter, r *http.Request) {
	var account model.UserAccount

	bodyReader, err := ioutil.ReadAll(r.Body)

	if err != nil {
		fmt.Println(err)
		util.RespondWithError(w, http.StatusBadRequest, "invalid data")
		return
	}

	err = util.UnmarshalJSON(bodyReader, &account)

	if err != nil {
		fmt.Println(err)
		util.RespondWithError(w, http.StatusBadRequest, "unmarshall failed")
		return
	}

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

	auth.ValidateJWT(account.Token)
}

// ChangePassword Function
func ChangePassword(w http.ResponseWriter, r *http.Request) {
	var account model.UserAccount
	var details map[string]interface{}

	bodyReader, err := ioutil.ReadAll(r.Body)

	if err != nil {
		fmt.Println(err)
		util.RespondWithError(w, http.StatusBadRequest, "invalid data")
		return
	}

	err = util.UnmarshalJSON(bodyReader, &details)

	if err != nil {
		fmt.Println(err)
		util.RespondWithError(w, http.StatusBadRequest, "unmarshall failed")
		return
	}

	filter := primitive.M{
		"username": details["username"],
	}

	update := primitive.M{
		"$set": primitive.M{
			"password": details["password"],
		},
	}

	updateVal, _ := database.FindAndUpdate("accounts", "users", filter, update)
	mapstructure.Decode(updateVal, &account)

	util.RespondwithJSON(w, http.StatusCreated, map[string]string{"message": "password changed"})
}

// RefreshToken Function
func RefreshToken(w http.ResponseWriter, r *http.Request) {
	var account model.UserAccount
	var filter primitive.M

	bodyReader, err := ioutil.ReadAll(r.Body)

	if err != nil {
		fmt.Println(err)
		util.RespondWithError(w, http.StatusBadRequest, "invalid data")
		return
	}

	err = util.UnmarshalJSON(bodyReader, &filter)

	if err != nil {
		fmt.Println(err)
		util.RespondWithError(w, http.StatusBadRequest, "unmarshall failed")
		return
	}

	findVal, findErr := database.FindOne("accounts", "users", filter)

	if findErr != nil {
		fmt.Println(findErr)
		util.RespondWithError(w, http.StatusBadRequest, "invalid account")
		return
	}

	mapstructure.Decode(findVal, &account)

	newToken, _ := auth.GenerateJWT(account)

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

	mapstructure.Decode(updateVal, &account)

	util.RespondwithJSON(w, http.StatusCreated, map[string]string{"message": "token updated", "token": newToken})
}
