package workflow

import (
	"github.com/mongodb/mongo-go-driver/bson/primitive"
)

// Workflow Struct
type Workflow struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Steps     []Step             `json:"steps" bson:"steps"`
	Claimcode int                `json:"claimCode" bson:"claimCode"`
}

// Step Struct
type Step struct {
	ID            primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	StepNum       int                `json:"stepNum" bson:"stepNum"`
	APICall       string             `json:"apiCall" bson:"apiCall"`
	DeviceAccount string             `json:"deviceAccount" bson:"deviceAccount"`
	VarMap        map[string]string  `json:"varMap" bson:"varMap"`
	Status        int                `json:"status" bson:"status"`
}
