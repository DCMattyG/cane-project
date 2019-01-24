package model

import (
	"crypto/rsa"

	"github.com/mongodb/mongo-go-driver/bson/primitive"
)

// BasicAuth Type
type BasicAuth struct {
	UserName string `json:"userName" bson:"userName" mapstructure:"userName"`
	Password string `json:"password" bson:"password" mapstructure:"password"`
}

// SessionAuth Type
type SessionAuth struct {
	UserName       string `json:"userName" bson:"userName" mapstructure:"userName"`
	Password       string `json:"password" bson:"password" mapstructure:"password"`
	CookieLifetime int32  `json:"cookieLifetime" bson:"cookieLifetime" mapstructure:"cookieLifetime"`
}

// APIKeyAuth Type
type APIKeyAuth struct {
	Header string `json:"header" bson:"header" mapstructure:"header"`
	Key    string `json:"key" bson:"key" mapstructure:"key"`
}

// Rfc3447Auth Type
type Rfc3447Auth struct {
	PublicKey  string          `json:"publicKey" bson:"publicKey" mapstructure:"publicKey"`
	PrivateKey *rsa.PrivateKey `json:"privateKey" bson:"privateKey" mapstructure:"privateKey"`
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

// Workflow Struct
type Workflow struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty" mapstructure:"_id"`
	Name        string             `json:"name" bson:"name" mapstructure:"name"`
	Description string             `json:"description" bson:"description" mapstructure:"description"`
	Type        string             `json:"type" bson:"type" mapstructure:"type"`
	Steps       []Step             `json:"steps" bson:"steps" mapstructure:"steps"`
	ClaimCode   int                `json:"claimCode" bson:"claimCode" mapstructure:"claimCode"`
	// Note, add OutputMap []map[string]string
}

// Step Struct
type Step struct {
	ID            primitive.ObjectID  `json:"id,omitempty" bson:"_id,omitempty" mapstructure:"_id"`
	StepNum       int                 `json:"stepNum" bson:"stepNum" mapstructure:"stepNum"`
	APICall       string              `json:"apiCall" bson:"apiCall" mapstructure:"apiCall"`
	DeviceAccount string              `json:"deviceAccount" bson:"deviceAccount" mapstructure:"deviceAccount"`
	VarMap        []map[string]string `json:"varMap" bson:"varMap" mapstructure:"varMap"`
	Status        int                 `json:"status" bson:"status" mapstructure:"status"`
}

// StepResult Struct
type StepResult struct {
	APICall    string `json:"apiCall" bson:"apiCall" mapstructure:"apiCall"`
	APIAccount string `json:"apiAccount" bson:"apiAccount" mapstructure:"apiAccount"`
	ReqBody    string `json:"reqBody" bson:"reqBody" mapstructure:"reqBody"`
	ResBody    string `json:"resBody" bson:"resBody" mapstructure:"resBody"`
	Error      error  `json:"error" bson:"error" mapstructure:"error"`
	Status     int    `json:"status" bson:"status" mapstructure:"status"`
}
