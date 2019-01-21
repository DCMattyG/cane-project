package model

import (
	"github.com/mongodb/mongo-go-driver/bson/primitive"
)

// API Struct
type API struct {
	ID            primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name          string             `json:"name" bson:"name"`
	DeviceAccount string             `json:"deviceAccount" bson:"deviceAccount"`
	Method        string             `json:"method" bson:"method"`
	URL           string             `json:"url" bson:"url"`
	Body          string             `json:"body" bson:"body"`
	Type          string             `json:"type" bson:"type"`
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

// DeviceAccount Struct
type DeviceAccount struct {
	ID       primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name     string             `json:"name" bson:"name"`
	IP       string             `json:"ip" bson:"ip"`
	UserName string             `json:"username" bson:"username"`
	Password string             `json:"password" bson:"password"`
	AuthType string             `json:"authtype" bson:"authtype"`
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
