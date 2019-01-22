package model

import (
	"crypto/rsa"

	"github.com/mongodb/mongo-go-driver/bson/primitive"
)

// BasicAuth Type
type BasicAuth struct {
	userName string
	password string
}

// SessionAuth Type
type SessionAuth struct {
	userName       string
	password       string
	cookieLifetime int32
}

// APIKeyAuth Type
type APIKeyAuth struct {
	key string
}

// Rfc3447Auth Type
type Rfc3447Auth struct {
	publicKey  string
	privateKey *rsa.PrivateKey
}

// API Struct
type API struct {
	ID            primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty" mapstructure:"_id"`
	Name          string             `json:"name" bson:"name"`
	DeviceAccount string             `json:"deviceAccount" bson:"deviceAccount"`
	Method        string             `json:"method" bson:"method"`
	URL           string             `json:"url" bson:"url"`
	Body          string             `json:"body" bson:"body"`
	Type          string             `json:"type" bson:"type"`
}

// UserAccount Struct
type UserAccount struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty" mapstructure:"_id"`
	FirstName string             `json:"fname" bson:"fname" mapstructure:"fname"`
	LastName  string             `json:"lname" bson:"lname" mapstructure:"lname"`
	UserName  string             `json:"username" bson:"username" mapstructure:"username"`
	Password  string             `json:"password" bson:"password" mapstructure:"password"`
	Privilege int                `json:"privilege" bson:"privilege" mapstructure:"privilege"`
	Enable    bool               `json:"enable" bson:"enable" mapstructure:"enable"`
	Token     string             `json:"token" bson:"token" mapstructure:"token"`
}

// DeviceAccount Struct
type DeviceAccount struct {
	ID       primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty" mapstructure:"_id"`
	Name     string             `json:"name" bson:"name" mapstructure:"name"`
	IP       string             `json:"ip" bson:"ip" mapstructure:"ip"`
	AuthType string             `json:"authtype" bson:"authtype" mapstructure:"authtype"`
	AuthObj  primitive.ObjectID `json:"authobj" bson:"authobj" mapstructure:"authobj"`
}

// RouteValue Struct
type RouteValue struct {
	ID       primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty" mapstructure:"_id"`
	Enable   bool               `json:"enable" bson:"enable" mapstructure:"enable"`
	Verb     string             `json:"verb" bson:"verb" mapstructure:"verb"`
	Version  int                `json:"version" bson:"version" mapstructure:"version"`
	Category string             `json:"category" bson:"category" mapstructure:"category"`
	Route    string             `json:"route" bson:"route" mapstructure:"route"`
	Message  map[string]string  `json:"message" bson:"message" mapstructure:"message"`
}
