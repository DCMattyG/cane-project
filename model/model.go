package model

import (
	"crypto/rsa"

	"github.com/mongodb/mongo-go-driver/bson/primitive"
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

// UserAccount Struct
type UserAccount struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	FirstName string             `json:"fname" bson:"fname"`
	LastName  string             `json:"lname" bson:"lname"`
	UserName  string             `json:"username" bson:"username"`
	Password  string             `json:"password" bson:"password"`
	Privilege int                `json:"privilege" bson:"privilege"`
	Enable    bool               `json:"enable" bson:"enable"`
	Token     string             `json:"token" bson:"token"`
}

// RouteValue Struct
type RouteValue struct {
	ID       primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Enable   bool               `json:"enable" bson:"enable"`
	Verb     string             `json:"verb" bson:"verb"`
	Version  int                `json:"version" bson:"version"`
	Category string             `json:"category" bson:"category"`
	Route    string             `json:"route" bson:"route"`
	Message  map[string]string  `json:"message" bson:"message"`
}
