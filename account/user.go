package account

import (
	"cane-project/database"
	"cane-project/util"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
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

// UserAccount Struct
type UserAccount struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	UserName  string             `json:"username" bson:"username"`
	Password  string             `json:"password" bson:"password"`
	Privilege int                `json:"privilege" bson:"privilege"`
	Enable    bool               `json:"enable" bson:"enable"`
	Token     string             `json:"token" bson:"token"`
}

// GenerateJWT Function
func GenerateJWT() (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)

	claims["authorized"] = true
	claims["client"] = "Matthew Garrett"
	claims["exp"] = time.Now().Add(time.Minute * 30).Unix()

	tokenString, err := token.SignedString(mySigningKey)

	if err != nil {
		fmt.Println("Something Went Wrong: ", err.Error())
		return "", err
	}

	return tokenString, nil
}

// ValidateJWT Function
func ValidateJWT(t string) {
	token, err := jwt.Parse(t, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("There was an error")
		}
		return mySigningKey, nil
	})

	if err != nil {
		fmt.Println(err.Error())
	}

	if token.Valid {
		fmt.Println("Valid Token!")
	} else {

		fmt.Println("Not Authorized!")
	}
}

// Login Function
func Login(w http.ResponseWriter, r *http.Request) {
	var account UserAccount
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
	var target UserAccount

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

	target.Token, err = GenerateJWT()

	userID, _ := database.Save("accounts", "users", target)

	fmt.Print("Inserted ID: ")
	fmt.Println(userID)

	util.RespondwithJSON(w, http.StatusCreated, map[string]string{"message": "user added"})
}

// ValidateUserToken Function
func ValidateUserToken(w http.ResponseWriter, r *http.Request) {
	var account UserAccount

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

	ValidateJWT(account.Token)
}

// ChangePassword Function
func ChangePassword(w http.ResponseWriter, r *http.Request) {
	var account UserAccount
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
	var account UserAccount
	var filter primitive.M

	newToken, _ := GenerateJWT()

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

	update := primitive.M{
		"$set": primitive.M{
			"token": newToken,
		},
	}

	updateVal, _ := database.FindAndUpdate("accounts", "users", filter, update)
	mapstructure.Decode(updateVal, &account)

	util.RespondwithJSON(w, http.StatusCreated, map[string]string{"message": "token updated", "token": newToken})
}
