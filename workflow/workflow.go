package workflow

import (
	"cane-project/database"
	"cane-project/util"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
)

// Workflow Struct
type Workflow struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name        string             `json:"name" bson:"name"`
	Description string             `json:"description" bson:"description"`
	Type        string             `json:"type" bson:"type"`
	Steps       []Step             `json:"steps" bson:"steps"`
	ClaimCode   int                `json:"claimCode" bson:"claimCode"`
}

// Step Struct
type Step struct {
	ID            primitive.ObjectID  `json:"id,omitempty" bson:"_id,omitempty"`
	StepNum       int                 `json:"stepNum" bson:"stepNum"`
	APICall       string              `json:"apiCall" bson:"apiCall"`
	DeviceAccount string              `json:"deviceAccount" bson:"deviceAccount"`
	VarMap        []map[string]string `json:"varMap" bson:"varMap"`
	Status        int                 `json:"status" bson:"status"`
}

// AddWorkflow Function
func AddWorkflow(w http.ResponseWriter, r *http.Request) {
	var target Workflow

	jsonErr := json.NewDecoder(r.Body).Decode(&target)

	if jsonErr != nil {
		fmt.Println(jsonErr)
		util.RespondWithError(w, http.StatusBadRequest, "error decoding json")
		return
	}

	filter := primitive.M{
		"name": target.Name,
	}

	_, findErr := database.FindOne("workflow", "workflows", filter)

	if findErr == nil {
		fmt.Println(findErr)
		util.RespondWithError(w, http.StatusBadRequest, "existing workflow")
		return
	}

	deviceID, _ := database.Save("workflows", "workflow", target)
	target.ID = deviceID.(primitive.ObjectID)

	fmt.Print("Inserted ID: ")
	fmt.Println(deviceID.(primitive.ObjectID).Hex())

	foundVal, _ := database.FindOne("workflow", "workflows", filter)

	util.RespondwithJSON(w, http.StatusCreated, foundVal)
}

// LoadWorkflow Function
func LoadWorkflow(w http.ResponseWriter, r *http.Request) {
	filter := primitive.M{
		"name": chi.URLParam(r, "name"),
	}

	foundVal, foundErr := database.FindOne("workflow", "workflows", filter)

	if foundErr != nil {
		fmt.Println(foundErr)
		util.RespondWithError(w, http.StatusBadRequest, "workflow not found")
		return
	}

	util.RespondwithJSON(w, http.StatusOK, foundVal)
}

// ListWorkflows Function
func ListWorkflows(w http.ResponseWriter, r *http.Request) {
	var workflows []string

	foundVal, foundErr := database.FindAll("workflow", "workflows", primitive.M{})

	if foundErr == nil {
		fmt.Println(foundErr)
		util.RespondWithError(w, http.StatusBadRequest, "no workflows found")
		return
	}

	if len(foundVal) == 0 {
		util.RespondWithError(w, http.StatusBadRequest, "no workflows found")
		return
	}

	for key := range foundVal {
		workflows = append(workflows, foundVal[key]["name"].(string))
	}

	util.RespondwithJSON(w, http.StatusOK, map[string][]string{"workflows": workflows})
}
