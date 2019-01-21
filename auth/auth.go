package auth

import (
	"cane-project/model"
	"crypto/rsa"
	"fmt"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-chi/jwtauth"
)

// Basic Auth Type
type Basic struct {
	userName string
	password string
}

// Session Auth Type
type Session struct {
	userName       string
	password       string
	cookieLifetime int32
}

// APIKey Auth Type
type APIKey struct {
	key string
}

// Rfc3447 Auth Type
type Rfc3447 struct {
	publicKey  string
	privateKey *rsa.PrivateKey
}

// MySigningKey Variable
var MySigningKey = []byte("secret")

// TokenAuth Variable
var TokenAuth *jwtauth.JWTAuth

// AuthTypes Variable
var AuthTypes = map[string]string{
	"basic":   "Basic",
	"session": "Session",
	"apikey":  "APIKey",
	"rfc3447": "Rfc3447",
}

func init() {
	TokenAuth = jwtauth.New("HS256", MySigningKey, nil)
}

// GenerateJWT Function
func GenerateJWT(account model.UserAccount) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)

	claims["authorized"] = true
	claims["client"] = account.FirstName + " " + account.LastName
	claims["time"] = time.Now().Unix()
	// claims["exp"] = time.Now().Add(time.Minute * 30).Unix()

	tokenString, err := token.SignedString(MySigningKey)

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
		return MySigningKey, nil
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
