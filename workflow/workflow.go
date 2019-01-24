package workflow

import (
	"cane-project/api"
	"cane-project/database"
	"cane-project/model"
	"cane-project/util"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/tidwall/sjson"

	"github.com/mitchellh/mapstructure"
	"github.com/tidwall/gjson"

	"github.com/go-chi/chi"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
)

// AddWorkflow Function
func AddWorkflow(w http.ResponseWriter, r *http.Request) {
	var target model.Workflow

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
	var targetWorkflow model.Workflow
	var setData gjson.Result
	var stepAPI model.API
	var stepAPIErr error

	// apiResults := make(map[string]interface{})
	apiResults := make(map[string]*model.StepResult)

	bodyBytes, bodyErr := ioutil.ReadAll(r.Body)
	stepZero := string(bodyBytes)

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

	fmt.Println("Initializing Step Results...")

	for i := 0; i < len(targetWorkflow.Steps); i++ {
		var step model.StepResult

		step.APICall = targetWorkflow.Steps[i].APICall
		step.APIAccount = targetWorkflow.Steps[i].DeviceAccount
		step.Status = 0

		apiResults[strconv.Itoa(i+1)] = &step
	}

	fmt.Println("Beginning Step Loop...")

	// For each step in "STEPS"
	for i := 0; i < len(targetWorkflow.Steps); i++ {
		fmt.Println("Setting API Status to 1...")

		apiResults[strconv.Itoa(i+1)].Status = 1

		fmt.Println("Loading Step API...")

		stepAPI, stepAPIErr = api.GetAPIFromDB(targetWorkflow.Steps[i].DeviceAccount, targetWorkflow.Steps[i].APICall)

		if stepAPIErr != nil {
			fmt.Println(stepAPIErr)
			util.RespondWithError(w, http.StatusBadRequest, "error loading target API")
			return
		}

		apiResults[strconv.Itoa(i+1)].APICall = stepAPI.Name
		apiResults[strconv.Itoa(i+1)].APIAccount = stepAPI.DeviceAccount

		fmt.Println("Beginning VarMap Loop...")

		// For each Variable Map in "VARMAP"
		for j := 0; j < len(targetWorkflow.Steps[i].VarMap); j++ {
			for key, val := range targetWorkflow.Steps[i].VarMap[j] {
				left := strings.Index(key, "{")
				right := strings.Index(key, "}")

				stepFrom := key[(left + 1):right]
				fromMap := key[(right + 1):]

				fmt.Println("From Step: ", stepFrom)
				fmt.Println("From Map: ", fromMap)
				fmt.Println("APIResults:")
				fmt.Println(apiResults)

				if stepFrom == "0" {
					setData = gjson.Get(stepZero, fromMap)
				} else {
					fmt.Println("Res Body: ", &apiResults[stepFrom].ResBody)
					setData = gjson.Get(apiResults[stepFrom].ResBody, fromMap)
				}

				var typedData interface{}

				fmt.Println("Determining TypeData...")
				fmt.Println("GJSON Results: ", gjson.Get(stepAPI.Body, val))

				if gjson.Get(stepAPI.Body, val).Exists() {
					switch dataKind := reflect.TypeOf(gjson.Get(stepAPI.Body, val).Value()).Kind(); dataKind {
					case reflect.Int:
						fmt.Println("Value: ", val)
						fmt.Println("Kind: ", dataKind)
						typedData = setData.Int()
					case reflect.Float64:
						fmt.Println("Value: ", val)
						fmt.Println("Kind: ", dataKind)
						typedData = setData.Float()
					case reflect.String:
						fmt.Println("Value: ", val)
						fmt.Println("Kind: ", dataKind)
						typedData = setData.String()
					default:
						fmt.Println("Value: ", val)
						fmt.Println("Unidentified Kind: ", dataKind)
					}
				} else {
					util.RespondWithError(w, http.StatusBadRequest, "Invalid mapping data")
					apiResults[strconv.Itoa(i+1)].Error = errors.New("Invalid mapping data")
					apiResults[strconv.Itoa(i+1)].Status = -1
					return
				}

				fmt.Println("Setting StepAPI Body...")

				stepAPI.Body, _ = sjson.Set(stepAPI.Body, val, typedData)
			}
		}

		fmt.Println("Updated Body: ", stepAPI.Body)
		apiResults[strconv.Itoa(i+1)].ReqBody = stepAPI.Body

		apiResp, apiErr := api.CallAPI(stepAPI)

		if apiErr != nil {
			fmt.Println(apiErr)
			util.RespondWithError(w, http.StatusBadRequest, "error executing API")
			apiResults[strconv.Itoa(i+1)].Error = apiErr
			apiResults[strconv.Itoa(i+1)].Status = -1
			return
		}

		fmt.Println("API Response:")
		fmt.Println(apiResp)

		defer apiResp.Body.Close()

		respBody, _ := ioutil.ReadAll(apiResp.Body)

		fmt.Println("API Response Body:")
		fmt.Println(string(respBody))

		apiResults[strconv.Itoa(i+1)].ResBody = string(respBody)

		bodyObject := make(map[string]interface{})
		marshalErr := json.Unmarshal(respBody, &bodyObject)

		if marshalErr != nil {
			fmt.Println(marshalErr)
			util.RespondWithError(w, http.StatusBadRequest, "error parsing response body")
			return
		}

		apiResults[strconv.Itoa(i+1)].Status = 2
	}

	util.RespondwithJSON(w, http.StatusOK, apiResults)
}
