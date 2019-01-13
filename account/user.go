package account

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
)

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
	Token     jwt.Token          `json:"token" bson:"token"`
}
