package workflow

import (
	"cane-project/api"
	"cane-project/database"
	"cane-project/model"
	"cane-project/util"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"

	"github.com/tidwall/sjson"

	"github.com/mitchellh/mapstructure"
	"github.com/tidwall/gjson"

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

	_, findErr := database.FindOne("workflows", "workflow", filter)

	if findErr == nil {
		fmt.Println(findErr)
		util.RespondWithError(w, http.StatusBadRequest, "existing workflow")
		return
	}

	deviceID, _ := database.Save("workflows", "workflow", target)
	target.ID = deviceID.(primitive.ObjectID)

	fmt.Print("Inserted ID: ")
	fmt.Println(deviceID.(primitive.ObjectID).Hex())

	foundVal, _ := database.FindOne("workflows", "workflow", filter)

	util.RespondwithJSON(w, http.StatusCreated, foundVal)
}

// LoadWorkflow Function
func LoadWorkflow(w http.ResponseWriter, r *http.Request) {
	filter := primitive.M{
		"name": chi.URLParam(r, "name"),
	}

	foundVal, foundErr := database.FindOne("workflows", "workflow", filter)

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

	foundVal, foundErr := database.FindAll("workflows", "workflow", primitive.M{})

	if foundErr != nil {
		fmt.Println(foundErr)
		util.RespondWithError(w, http.StatusBadRequest, "no workflows found")
		return
	}

	if len(foundVal) == 0 {
		fmt.Println(foundVal)
		util.RespondWithError(w, http.StatusBadRequest, "empty workflows list")
		return
	}

	for key := range foundVal {
		workflows = append(workflows, foundVal[key]["name"].(string))
	}

	util.RespondwithJSON(w, http.StatusOK, map[string][]string{"workflows": workflows})
}

// ExecuteWorkflow Function
func ExecuteWorkflow(w http.ResponseWriter, r *http.Request) {
	var targetWorkflow Workflow
	var stepAPI model.API
	// var stepDevice model.DeviceAccount
	var stepAPIErr error
	// var stepDeviceErr error

	apiResults := map[string]interface{}{}

	bodyBytes, bodyErr := ioutil.ReadAll(r.Body)
	bodyString := string(bodyBytes)

	if bodyErr != nil {
		fmt.Println(bodyErr)
		util.RespondWithError(w, http.StatusBadRequest, "error reading body")
		return
	}

	filter := primitive.M{
		"name": chi.URLParam(r, "name"),
	}

	foundVal, foundErr := database.FindOne("workflows", "workflow", filter)

	if foundErr != nil {
		fmt.Println(foundErr)
		util.RespondWithError(w, http.StatusBadRequest, "workflow not found")
		return
	}

	mapErr := mapstructure.Decode(foundVal, &targetWorkflow)

	if mapErr != nil {
		fmt.Println(mapErr)
		util.RespondWithError(w, http.StatusBadRequest, "error parsing workflow")
		return
	}

	// For each step in "STEPS"
	for i := 0; i < len(targetWorkflow.Steps); i++ {
		stepAPI, stepAPIErr = api.GetAPIFromDB(targetWorkflow.Steps[i].DeviceAccount, targetWorkflow.Steps[i].APICall)

		if stepAPIErr != nil {
			fmt.Println(stepAPIErr)
			util.RespondWithError(w, http.StatusBadRequest, "error loading target API")
			return
		}

		// stepDevice, stepDeviceErr = account.GetDeviceFromDB(targetWorkflow.Steps[i].DeviceAccount)

		// if stepDeviceErr != nil {
		// 	fmt.Println(stepDeviceErr)
		// 	util.RespondWithError(w, http.StatusBadRequest, "error loading target device")
		// 	return
		// }

		// For each Variable Map in "VARMAP"
		for j := 0; j < len(targetWorkflow.Steps[i].VarMap); j++ {
			for key, val := range targetWorkflow.Steps[i].VarMap[j] {
				left := strings.Index(key, "{")
				right := strings.Index(key, "}")

				stepFrom := key[(left + 1):right]
				fromMap := key[(right + 1):]

				setData := gjson.Get(bodyString, fromMap)

				fmt.Println("From Step: ", stepFrom)

				var typedData interface{}

				switch dataKind := reflect.TypeOf(gjson.Get(stepAPI.Body, val).Value()).Kind(); dataKind {
				case reflect.Int:
					// fmt.Println("Value: ", val)
					// fmt.Println("Kind: ", dataKind)
					typedData = setData.Int()
				case reflect.Float64:
					// fmt.Println("Value: ", val)
					// fmt.Println("Kind: ", dataKind)
					typedData = setData.Float()
				case reflect.String:
					// fmt.Println("Value: ", val)
					// fmt.Println("Kind: ", dataKind)
					typedData = setData.String()
				default:
					fmt.Println("Value: ", val)
					fmt.Println("Unidentified Kind: ", dataKind)
				}

				stepAPI.Body, _ = sjson.Set(stepAPI.Body, val, typedData)
				fmt.Println("Updated Body: ", stepAPI.Body)

				apiResp, _ := api.CallAPI(stepAPI)
				fmt.Println("API Response:")
				fmt.Println(apiResp)

				defer apiResp.Body.Close()

				respBody, _ := ioutil.ReadAll(apiResp.Body)
				// bodyObject := make(map[string]interface{})

				marshalErr := json.Unmarshal(respBody, &apiResults)

				if marshalErr != nil {
					fmt.Println(marshalErr)
					util.RespondWithError(w, http.StatusBadRequest, "error parsing response body")
					return
				}
			}
		}
	}

	util.RespondwithJSON(w, http.StatusOK, apiResults)
}
