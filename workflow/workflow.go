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
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/mongodb/mongo-go-driver/mongo/options"

	"github.com/tidwall/sjson"

	"github.com/mitchellh/mapstructure"
	"github.com/tidwall/gjson"

	"github.com/go-chi/chi"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
)

// CreateWorkflow Function
func CreateWorkflow(w http.ResponseWriter, r *http.Request) {
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

	_, saveErr := database.Save("workflows", "workflow", target)

	if saveErr != nil {
		fmt.Println(saveErr)
		util.RespondWithError(w, http.StatusBadRequest, "error saving workflow")
		return
	}

	util.RespondwithString(w, http.StatusCreated, "")
}

// DeleteWorkflow Function
func DeleteWorkflow(w http.ResponseWriter, r *http.Request) {
	filter := primitive.M{
		"name": chi.URLParam(r, "workflowname"),
	}

	deleteErr := database.Delete("workflows", "workflow", filter)

	if deleteErr != nil {
		fmt.Println(deleteErr)
		util.RespondWithError(w, http.StatusBadRequest, "workflow not found")
		return
	}

	util.RespondwithString(w, http.StatusOK, "")
}

// GetWorkflow Function
func GetWorkflow(w http.ResponseWriter, r *http.Request) {
	filter := primitive.M{
		"name": chi.URLParam(r, "workflowname"),
	}

	foundVal, foundErr := database.FindOne("workflows", "workflow", filter)

	if foundErr != nil {
		fmt.Println(foundErr)
		util.RespondWithError(w, http.StatusBadRequest, "workflow not found")
		return
	}

	util.RespondwithJSON(w, http.StatusOK, foundVal)
}

// GetWorkflows Function
func GetWorkflows(w http.ResponseWriter, r *http.Request) {
	var opts options.FindOptions
	var workflows []string

	foundVal, foundErr := database.FindAll("workflows", "workflow", primitive.M{}, opts)

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

// CallWorkflow Function
func CallWorkflow(w http.ResponseWriter, r *http.Request) {
	var targetWorkflow model.Workflow

	bodyBytes, bodyErr := ioutil.ReadAll(r.Body)
	stepZero := string(bodyBytes)

	if bodyErr != nil {
		fmt.Println(bodyErr)
		util.RespondWithError(w, http.StatusBadRequest, "error reading body")
		return
	}

	filter := primitive.M{
		"name": chi.URLParam(r, "workflowname"),
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

	fmt.Println("Generating Claim Code...")
	workflowClaim := GenerateClaim()
	workflowClaim.Save()
	util.RespondwithJSON(w, http.StatusCreated, map[string]interface{}{"claimCode": workflowClaim.ClaimCode})

	go ExecuteWorkflow(stepZero, targetWorkflow, workflowClaim)
}

// ExecuteWorkflow Function
func ExecuteWorkflow(stepZero string, targetWorkflow model.Workflow, workflowClaim Claim) {
	// var setData gjson.Result
	var stepAPI model.API
	var stepAPIErr error

	apiResults := make(map[string]model.StepResult)
	varPool := make(map[string]map[string]string)
	zeroMap := make(map[string]interface{})

	bodyBuilder := ""
	stepHeader := make(map[string]string)
	var stepQuery url.Values

	fmt.Println("Beginning Step Loop...")

	fmt.Println("Step Zero Body:")
	fmt.Println(stepZero)
	fmt.Println("Step Zero Body Length:")
	fmt.Println(len(stepZero))

	fmt.Println("Unmarshal Step Zero BODY to MAP")

	if len(stepZero) != 0 {
		zeroErr := json.Unmarshal([]byte(stepZero), &zeroMap)

		if zeroErr != nil {
			fmt.Println("Error unmarshalling Zero BODY to MAP!")
			fmt.Println(zeroErr)
			return
		}
	} else {
		fmt.Println("Empty BODY, setting to {}")
	}

	fmt.Println("Zero MAP:")
	fmt.Println(zeroMap)

	// Add Variables from ZeroMap to VarPool
	fmt.Println("Adding Variables from ZeroMap...")

	for key, val := range zeroMap {
		fmt.Println("Mapping Zero Variable: " + val.(string))

		switch val.(type) {
		case int, int32, int64:
			fmt.Println("Mapping: " + val.(string) + " as (int)")
			varPool[key] = map[string]string{val.(string): "int"}
		case float32, float64:
			fmt.Println("Mapping: " + val.(string) + " as (float)")
			varPool[key] = map[string]string{val.(string): "float"}
		case string:
			fmt.Println("Mapping: " + val.(string) + " as (string)")
			varPool[key] = map[string]string{val.(string): "string"}
		default:
			fmt.Println("Unknown: " + val.(string) + " type (" + reflect.TypeOf(val).String() + ")")
		}
	}

	// For each step in "STEPS"
	for i := 0; i < len(targetWorkflow.Steps); i++ {
		var step model.StepResult

		fmt.Println("Setting API Status to 1...")

		step.Status = 1
		apiResults[strconv.Itoa(i+1)] = step
		workflowClaim.WorkflowResults = apiResults
		workflowClaim.CurrentStatus = 1
		workflowClaim.Save()

		fmt.Println("Loading Step API...")

		stepAPI, stepAPIErr = api.GetAPIFromDB(targetWorkflow.Steps[i].DeviceAccount, targetWorkflow.Steps[i].APICall)

		if stepAPIErr != nil {
			fmt.Println("Error getting API from DB...")
			fmt.Println(stepAPIErr)
			step.Error = stepAPIErr.Error()
			step.Status = -1
			apiResults[strconv.Itoa(i+1)] = step
			workflowClaim.WorkflowResults = apiResults
			workflowClaim.CurrentStatus = -1
			workflowClaim.Save()
			return
		}

		step.APICall = stepAPI.Name
		step.APIAccount = stepAPI.DeviceAccount

		// Build BODY string from Map
		fmt.Println("Building BODY from Map...")

		bodyBuilder = ""
		// stepBody = make(map[string]string)

		for bodyCount := 0; bodyCount < len(targetWorkflow.Steps[i].Body); bodyCount++ {
			for key, val := range targetWorkflow.Steps[i].Body[bodyCount] {
				fmt.Println("Processing BODY Var: (" + val + ")")
				if util.IsVar(val) {
					fmt.Println("Found variable: " + val)
					val = strings.Replace(val, "{{", "", 1)
					val = strings.Replace(val, "}}", "", 1)
					fmt.Println("Stripped variable: (" + val + ")")
				}

				if poolVal, ok := varPool[val]; ok {
					for replaceVar, replaceType := range poolVal {
						switch replaceType {
						case "int":
							fmt.Println("(" + replaceVar + ") is an INT")
							bodyVal, _ := strconv.ParseInt(replaceVar, 10, 64)
							bodyBuilder, _ = sjson.Set(bodyBuilder, key, bodyVal)
						case "float":
							fmt.Println("(" + replaceVar + ") is a FLOAT")
							bodyVal, _ := strconv.ParseFloat(replaceVar, 64)
							bodyBuilder, _ = sjson.Set(bodyBuilder, key, bodyVal)
						case "string":
							fmt.Println("(" + replaceVar + ") is a STRING")
							bodyBuilder, _ = sjson.Set(bodyBuilder, key, replaceVar)
						}
					}
				} else {
					bodyBuilder, _ = sjson.Set(bodyBuilder, key, val)
				}
			}
		}

		fmt.Println("BodyBuider:")
		fmt.Println(bodyBuilder)
		fmt.Println("BodyBuider Length:")
		fmt.Println(len(bodyBuilder))

		fmt.Println("Parsing BODY to Map...")

		var stepBody = make(map[string]interface{})

		if len(bodyBuilder) != 0 {
			decoder := json.NewDecoder(strings.NewReader(bodyBuilder))
			decoder.UseNumber()
			bodyErr := decoder.Decode(&stepBody)

			if bodyErr != nil {
				fmt.Println("Error parsing BODY to Map!")
				return
			}
		} else {
			fmt.Println("Empty BodyBuilder, setting StepBody to {}")
		}

		fmt.Println("BODY:")
		fmt.Println(stepBody)

		// Build HEADER from Map
		fmt.Println("Building HEADER from Map...")

		stepHeader = make(map[string]string)

		for headerCount := 0; headerCount < len(targetWorkflow.Steps[i].Headers); headerCount++ {
			for key, val := range targetWorkflow.Steps[i].Headers[headerCount] {
				if util.IsVar(val) {
					fmt.Println("Found variable: " + val)
				}

				if poolVal, ok := varPool[val]; ok {
					for replaceVar := range poolVal {
						stepHeader[key] = replaceVar
					}
				} else {
					stepHeader[key] = val
				}
			}
		}

		fmt.Println("HEADER:")
		fmt.Println(stepHeader)

		step.ReqHeaders = stepHeader

		// Build QUERY from Map
		fmt.Println("Building QUERY from Map...")

		stepQuery = make(map[string][]string)

		for queryCount := 0; queryCount < len(targetWorkflow.Steps[i].Query); queryCount++ {
			for key, val := range targetWorkflow.Steps[i].Query[queryCount] {
				if util.IsVar(val) {
					fmt.Println("Found variable: " + val)
				}

				if poolVal, ok := varPool[val]; ok {
					for replaceVar := range poolVal {
						stepQuery.Add(key, replaceVar)
					}
				} else {
					stepQuery.Add(key, val)
				}
			}
		}

		fmt.Println("QUERY:")
		fmt.Println(stepQuery)

		step.ReqQuery = stepQuery

		fmt.Println("Updated Body: ", stepAPI.Body)
		step.ReqBody = stepBody

		varMatch := regexp.MustCompile(`([{]{2}[a-zA-Z]*[}]{2}){1}`)
		searchPath := varMatch.FindString(stepAPI.Path)

		for searchPath != "" {
			fmt.Println("SearchPath: " + searchPath)

			val := searchPath
			val = strings.Replace(val, "{{", "", 1)
			val = strings.Replace(val, "}}", "", 1)

			fmt.Println("Variable to Replace: " + val)
			fmt.Println("Current Variable Pool:")
			fmt.Println(varPool)

			if poolVal, ok := varPool[val]; ok {
				for replaceVar := range poolVal {
					fmt.Println("Replace Variable Value: " + replaceVar)
					stepAPI.Path = strings.Replace(stepAPI.Path, searchPath, replaceVar, 1)
				}
			} else {
				fmt.Println("Replace Variable Not Found!")
				stepAPI.Path = strings.Replace(stepAPI.Path, searchPath, "<error>", 1)
			}

			searchPath = varMatch.FindString(stepAPI.Path)
		}

		fmt.Println("Updated API Path:")
		fmt.Println(stepAPI.Path)

		apiResp, apiErr := api.CallAPI(stepAPI, stepQuery, stepHeader)

		if apiErr != nil {
			fmt.Println(apiErr)
			step.Error = apiErr.Error()
			step.Status = -1
			apiResults[strconv.Itoa(i+1)] = step
			workflowClaim.WorkflowResults = apiResults
			workflowClaim.CurrentStatus = -1
			workflowClaim.Save()
			return
		}

		fmt.Println("API Response:")
		fmt.Println(apiResp)

		defer apiResp.Body.Close()

		respBody, respErr := ioutil.ReadAll(apiResp.Body)

		if respErr != nil {
			fmt.Println(respErr)
			step.Error = respErr.Error()
			step.Status = -1
			apiResults[strconv.Itoa(i+1)] = step
			workflowClaim.WorkflowResults = apiResults
			workflowClaim.CurrentStatus = -1
			workflowClaim.Save()
			return
		}

		fmt.Println("API Response Body:")
		fmt.Println(string(respBody))

		step.ResBody = string(respBody)

		if !gjson.Valid(step.ResBody) {
			fmt.Println("GJSON Reports ResBody is invalid JSON!")
		}

		// Extract VARIABLES from Response BODY
		fmt.Println("Extracting VARIABLES from Reponse Body...")

		for varCount := 0; varCount < len(targetWorkflow.Steps[i].Variables); varCount++ {
			for key, val := range targetWorkflow.Steps[i].Variables[varCount] {
				fmt.Println("Extracting Variable: " + val)

				varValue := gjson.Get(step.ResBody, val)

				if varValue.Exists() {
					fmt.Println("(" + val + ") Found! Value: " + varValue.String())
					fmt.Println("GJSON Type:" + string(varValue.Type.String()))

					switch varKind := reflect.TypeOf(varValue.Value()).Kind(); varKind {
					case reflect.Int:
						fmt.Println("Value: ", val)
						fmt.Println("Kind: ", varKind)
						varPool[key] = map[string]string{varValue.String(): "int"}
					case reflect.Float64:
						fmt.Println("Value: ", val)
						fmt.Println("Kind: ", varKind)
						if strings.ContainsAny(varValue.String(), ".") {
							varPool[key] = map[string]string{varValue.String(): "float"}
						} else {
							fmt.Println("No decimal, storing as INT...")
							varPool[key] = map[string]string{varValue.String(): "int"}
						}
					case reflect.String:
						fmt.Println("Value: ", val)
						fmt.Println("Kind: ", varKind)
						varPool[key] = map[string]string{varValue.String(): "string"}
					default:
						fmt.Println("Value: ", val)
						fmt.Println("Unidentified Kind: ", varKind)
					}
				} else {
					fmt.Println("(" + val + ") not found in Response Body!")
				}
			}
		}

		step.Status = 2
		apiResults[strconv.Itoa(i+1)] = step
		workflowClaim.WorkflowResults = apiResults
		// workflowClaim.CurrentStatus = 2
		workflowClaim.Save()
	}

	workflowClaim.CurrentStatus = 2
	workflowClaim.Save()
}

// ExecuteWorkflow Function
// func ExecuteWorkflow(stepZero string, targetWorkflow model.Workflow, workflowClaim Claim) {
// 	var setData gjson.Result
// 	var stepAPI model.API
// 	var stepAPIErr error

// 	apiResults := make(map[string]model.StepResult)

// 	fmt.Println("Beginning Step Loop...")

// 	fmt.Println("Step Zero Body:")
// 	fmt.Println(stepZero)

// 	// For each step in "STEPS"
// 	for i := 0; i < len(targetWorkflow.Steps); i++ {
// 		var step model.StepResult

// 		fmt.Println("Setting API Status to 1...")

// 		step.Status = 1
// 		apiResults[strconv.Itoa(i+1)] = step
// 		workflowClaim.WorkflowResults = apiResults
// 		workflowClaim.CurrentStatus = 1
// 		workflowClaim.Save()

// 		fmt.Println("Loading Step API...")

// 		stepAPI, stepAPIErr = api.GetAPIFromDB(targetWorkflow.Steps[i].DeviceAccount, targetWorkflow.Steps[i].APICall)

// 		if stepAPIErr != nil {
// 			fmt.Println(stepAPIErr)
// 			step.Error = stepAPIErr.Error()
// 			step.Status = -1
// 			apiResults[strconv.Itoa(i+1)] = step
// 			workflowClaim.WorkflowResults = apiResults
// 			workflowClaim.CurrentStatus = -1
// 			workflowClaim.Save()
// 			return
// 		}

// 		step.APICall = stepAPI.Name
// 		step.APIAccount = stepAPI.DeviceAccount

// 		fmt.Println("Beginning VarMap Loop...")

// 		// For each Variable Map in "VARMAP"
// 		for j := 0; j < len(targetWorkflow.Steps[i].VarMap); j++ {
// 			for key, val := range targetWorkflow.Steps[i].VarMap[j] {
// 				left := strings.Index(key, "{")
// 				right := strings.Index(key, "}")

// 				stepFrom := key[(left + 1):right]
// 				fromMap := key[(right + 1):]

// 				fmt.Println("From Step: ", stepFrom)
// 				fmt.Println("From Map: ", fromMap)
// 				fmt.Println("APIResults:")
// 				fmt.Println(apiResults)

// 				if stepFrom == "0" {
// 					setData = gjson.Get(stepZero, fromMap)
// 				} else if stepFrom == "s" {
// 					var gString gjson.Result
// 					gString.Str = fromMap
// 					gString.Type = gjson.String
// 					setData = gString
// 				} else if stepFrom == "n" {
// 					var gString gjson.Result
// 					gString.Num, _ = strconv.ParseFloat(fromMap, 64)
// 					gString.Type = gjson.Number
// 					setData = gString
// 				} else {
// 					fmt.Println("Res Body: ", apiResults[stepFrom].ResBody)
// 					setData = gjson.Get(apiResults[stepFrom].ResBody, fromMap)
// 				}

// 				var typedData interface{}

// 				stepAPI.Body = strings.Replace(stepAPI.Body, "\n", "", -1)
// 				stepAPI.Body = strings.Replace(stepAPI.Body, "\t", "", -1)
// 				stepAPI.Body = strings.Replace(stepAPI.Body, "\r", "", -1)
// 				stepAPI.Body = strings.Replace(stepAPI.Body, "\\", "", -1)

// 				fmt.Println("STRIPPED API BODY:")
// 				fmt.Println(stepAPI.Body)

// 				fmt.Println("Determining TypeData...")
// 				fmt.Println("GJSON Results: ", gjson.Get(stepAPI.Body, val))

// 				if gjson.Get(stepAPI.Body, val).Exists() {
// 					switch dataKind := reflect.TypeOf(gjson.Get(stepAPI.Body, val).Value()).Kind(); dataKind {
// 					case reflect.Int:
// 						fmt.Println("Value: ", val)
// 						fmt.Println("Kind: ", dataKind)
// 						typedData = setData.Int()
// 					case reflect.Float64:
// 						fmt.Println("Value: ", val)
// 						fmt.Println("Kind: ", dataKind)
// 						typedData = setData.Float()
// 					case reflect.String:
// 						fmt.Println("Value: ", val)
// 						fmt.Println("Kind: ", dataKind)
// 						typedData = setData.String()
// 					default:
// 						fmt.Println("Value: ", val)
// 						fmt.Println("Unidentified Kind: ", dataKind)
// 					}
// 				} else {
// 					fmt.Println("Mapping Error!")
// 					fmt.Println("Step Body:")
// 					fmt.Println(stepAPI.Body)
// 					output := fmt.Sprintf("Map Value [%s]", val)
// 					fmt.Println(output)
// 					step.Error = "Invalid mapping data, target value does not exist"
// 					step.Status = -1
// 					apiResults[strconv.Itoa(i+1)] = step
// 					workflowClaim.WorkflowResults = apiResults
// 					workflowClaim.CurrentStatus = -1
// 					workflowClaim.Save()
// 					return
// 				}

// 				fmt.Println("Setting StepAPI Body...")

// 				var sjsonSetErr error
// 				stepAPI.Body, sjsonSetErr = sjson.Set(stepAPI.Body, val, typedData)

// 				if sjsonSetErr != nil {
// 					fmt.Println(sjsonSetErr)
// 					step.Error = sjsonSetErr.Error()
// 					step.Status = -1
// 					apiResults[strconv.Itoa(i+1)] = step
// 					workflowClaim.WorkflowResults = apiResults
// 					workflowClaim.CurrentStatus = -1
// 					workflowClaim.Save()
// 					return
// 				}
// 			}
// 		}

// 		fmt.Println("Updated Body: ", stepAPI.Body)
// 		step.ReqBody = stepAPI.Body

// 		apiResp, apiErr := api.CallAPI(stepAPI, nil)

// 		if apiErr != nil {
// 			fmt.Println(apiErr)
// 			step.Error = apiErr.Error()
// 			step.Status = -1
// 			apiResults[strconv.Itoa(i+1)] = step
// 			workflowClaim.WorkflowResults = apiResults
// 			workflowClaim.CurrentStatus = -1
// 			workflowClaim.Save()
// 			return
// 		}

// 		fmt.Println("API Response:")
// 		fmt.Println(apiResp)

// 		defer apiResp.Body.Close()

// 		respBody, respErr := ioutil.ReadAll(apiResp.Body)

// 		if respErr != nil {
// 			fmt.Println(respErr)
// 			step.Error = respErr.Error()
// 			step.Status = -1
// 			apiResults[strconv.Itoa(i+1)] = step
// 			workflowClaim.WorkflowResults = apiResults
// 			workflowClaim.CurrentStatus = -1
// 			workflowClaim.Save()
// 			return
// 		}

// 		fmt.Println("API Response Body:")
// 		fmt.Println(string(respBody))

// 		step.ResBody = string(respBody)

// 		step.Status = 2
// 		apiResults[strconv.Itoa(i+1)] = step
// 		workflowClaim.WorkflowResults = apiResults
// 		workflowClaim.CurrentStatus = 2
// 		workflowClaim.Save()
// 	}
// }
