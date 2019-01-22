package routing

import (
	"bytes"
	"cane-project/account"
	"cane-project/api"
	"cane-project/auth"
	"cane-project/database"
	"cane-project/model"
	"cane-project/util"
	"cane-project/workflow"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/jwtauth"
	"github.com/mitchellh/mapstructure"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"github.com/tidwall/gjson"
)

// Router Variable
var Router *chi.Mux

func init() {
	Router = chi.NewRouter()
}

// Routers Function
func Routers() {
	var iterVals []model.RouteValue
	Router = chi.NewMux()

	filter := primitive.M{}
	foundVals, _ := database.FindAll("routing", "routes", filter)
	mapstructure.Decode(foundVals, &iterVals)

	fmt.Println("Updating routes...")

	// Public Default Routes
	Router.Post("/addRoute", AddRoutes)
	Router.Post("/parseVars", ParseVars)
	Router.Post("/login", account.Login)
	Router.Post("/addUser", account.AddUser)
	Router.Post("/validateToken", account.ValidateUserToken)
	Router.Patch("/updateToken/{user}", account.RefreshToken)
	Router.Post("/apiTest", TestCallAPI)
	Router.Post("/addDevice", account.AddDevice)
	Router.Get("/loadDevice/{name}", account.LoadDevice)
	Router.Patch("/updateDevice/{name}", account.UpdateDevice)
	Router.Get("/listDevice", account.ListDevices)
	Router.Get("/deviceApis/{device}", account.ListDeviceAPIs)
	Router.Post("/addApi", api.AddAPI)
	Router.Get("/testPath/*", TestPath)
	// Router.Get("/aplTest", APLTest)
	Router.Post("/testJSON", JSONTest)
	Router.Post("/testXML", XMLTest)
	Router.Post("/testGJSON", TestGJSON)
	Router.Post("/testAPIAuth", TestAPIAuth)
	Router.Post("/addWorkflow", workflow.AddWorkflow)
	Router.Get("/listWorkflow", workflow.ListWorkflows)

	// Private Default Routes
	Router.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(auth.TokenAuth))
		r.Use(jwtauth.Authenticator)
		r.Patch("/changePassword/{user}", account.ChangePassword)
		r.Get("/test", TestPost)
	})

	// Dynamic Routes
	for i := range iterVals {
		routeVal := iterVals[i]

		if routeVal.Enable {
			newRoute := "/v" + strconv.Itoa(routeVal.Version) + "/" + routeVal.Category + "/" + routeVal.Route
			newMessage := routeVal.Message

			if routeVal.Verb == "get" {
				Router.Get(newRoute, func(w http.ResponseWriter, r *http.Request) {
					util.RespondwithJSON(w, http.StatusCreated, newMessage)
				})
			} else if routeVal.Verb == "post" {
				Router.Post(newRoute, func(w http.ResponseWriter, r *http.Request) {
					postJSON := make(map[string]interface{})
					err := json.NewDecoder(r.Body).Decode(&postJSON)

					fmt.Println(postJSON)

					if err != nil {
						panic(err)
					}

					util.RespondwithJSON(w, http.StatusCreated, postJSON)
				})
			}
		}
	}
}

// ParseVars function
func ParseVars(w http.ResponseWriter, r *http.Request) {
	bodyReader, err := ioutil.ReadAll(r.Body)

	// fmt.Println(string(bodyReader))

	if err != nil {
		fmt.Println(err)
		util.RespondWithError(w, http.StatusBadRequest, "invalid data")
		return
	}

	if model.IsJSON(string(bodyReader)) {
		var j model.JSONNode

		jsonErr := json.Unmarshal(bodyReader, &j)

		if jsonErr != nil {
			fmt.Println(jsonErr)
			util.RespondWithError(w, http.StatusBadRequest, "invalid json")
			return
		}

		j.JSONVars()

		jsonAPI := map[string]string{
			"parsedAPI": j.Marshal(),
			"type":      "json",
		}

		util.RespondwithJSON(w, http.StatusOK, jsonAPI)
	}

	if model.IsXML(string(bodyReader)) {
		var x model.XMLNode

		buf := bytes.NewBuffer(bodyReader)
		dec := xml.NewDecoder(buf)
		xmlErr := dec.Decode(&x)

		if xmlErr != nil {
			fmt.Println(xmlErr)
			util.RespondWithError(w, http.StatusBadRequest, "invalid xml")
			return
		}

		x.ScrubXML()
		// x.XMLVars()

		xmlAPI := map[string]string{
			"parsedAPI": x.Marshal(),
			"type":      "xml",
		}

		util.RespondwithJSON(w, http.StatusOK, xmlAPI)
		//util.RespondwithXML(w, http.StatusOK, x)
	}
}

// TestPost function
func TestPost(w http.ResponseWriter, r *http.Request) {
	util.RespondwithJSON(w, http.StatusCreated, map[string]string{"message": "test post"})
}

// TestPath function
func TestPath(w http.ResponseWriter, r *http.Request) {
	path := chi.URLParam(r, "*")
	routeContext := chi.RouteContext(r.Context())
	routePattern := routeContext.RoutePattern()

	account := strings.Replace(routePattern, "/", "", -1)
	account = strings.Replace(account, "*", "", -1)

	mapResponse := map[string]string{
		"path":        path,
		"routePatten": routePattern,
		"account":     account,
	}

	util.RespondwithJSON(w, http.StatusCreated, mapResponse)
}

// AddRoutes function
func AddRoutes(w http.ResponseWriter, r *http.Request) {
	var target model.RouteValue

	json.NewDecoder(r.Body).Decode(&target)

	if !(ValidateRoute(target)) {
		util.RespondWithError(w, http.StatusBadRequest, "invalid route")
		return
	}

	fmt.Println("Adding routes to database...")

	postID, postErr := database.Save("routing", "routes", target)

	if postErr != nil {
		fmt.Println(postErr)
		util.RespondWithError(w, http.StatusBadRequest, "failed saving route")
		return
	}

	Routers()

	util.RespondwithJSON(w, http.StatusCreated, postID)
}

// ValidateRoute Function
func ValidateRoute(route model.RouteValue) bool {
	verbs := []string{"get", "post", "patch", "delete"}
	categories := []string{"network", "compute", "storage", "security", "virtualization", "cloud"}

	if !(util.StringInSlice(verbs, route.Verb)) {
		return false
	}

	if !(util.StringInSlice(categories, route.Category)) {
		return false
	}

	return true
}

// TestCallAPI Function
func TestCallAPI(w http.ResponseWriter, r *http.Request) {
	// var respBody map[string]interface{}
	var apiInput map[string]interface{}
	var callAPI model.API

	json.NewDecoder(r.Body).Decode(&apiInput)

	targetFilter := primitive.M{
		"name": apiInput["apiName"],
	}

	targetAPI, targetErr := database.FindOne("apis", apiInput["deviceAccount"].(string), targetFilter)

	if targetErr != nil {
		fmt.Println(targetErr)
		util.RespondWithError(w, http.StatusBadRequest, "no such api")
		return
	}

	mapstructure.Decode(targetAPI, &callAPI)

	apiMap := apiInput["apiMap"].(map[string]interface{})

	// fmt.Println(apiMap)

	tempAPI := targetAPI["body"].(string)

	// fmt.Println(tempAPI)

	for key, val := range apiMap {
		tempAPI = strings.Replace(tempAPI, key, val.(string), 1)
	}

	callAPI.Body = tempAPI

	// resp := api.CallAPI(callAPI)

	// defer resp.Body.Close()
	// respBody, _ := ioutil.ReadAll(resp.Body)

	// fmt.Println(string(respBody))

	respBody := ""

	util.RespondwithJSON(w, http.StatusCreated, string(respBody))
}

// APLTest Function
// func APLTest(w http.ResponseWriter, r *http.Request) {
// 	model.APLtoJSON()
// }

// JSONTest Function
func JSONTest(w http.ResponseWriter, r *http.Request) {
	test := map[string]interface{}{
		"results": []map[string]interface{}{
			{
				"moid": "65f345ff345sss24dd",
				"type": "compute",
				"tags": []string{"hot", "cold"},
				"parent": map[string]interface{}{
					"moid": "43r34r4h743834",
					"obj":  "chassis",
				},
			},
		},
	}

	fixed := model.JSONNode(test).StripJSON()

	// model.JSONNode(test).StripJSON()

	util.RespondwithJSON(w, http.StatusCreated, fixed)
}

// XMLTest Function
func XMLTest(w http.ResponseWriter, r *http.Request) {
	bodyReader, _ := ioutil.ReadAll(r.Body)

	x, xmlErr := model.XMLfromBytes(bodyReader)

	if xmlErr != nil {
		fmt.Println(xmlErr)
		util.RespondWithError(w, http.StatusBadRequest, "invalid xml")
	}

	mapString := x.XMLtoJSON()

	fmt.Println(x.XMLtoJSON())
	fmt.Println("-------------------------------")

	jBytes, _ := json.MarshalIndent(mapString, "", "  ")

	jString := string(jBytes)

	fmt.Println(jString)
	fmt.Println("-------------------------------")

	j, jsonErr := model.JSONfromBytes(jBytes)

	if jsonErr != nil {
		fmt.Println(jsonErr)
		util.RespondWithError(w, http.StatusBadRequest, "invalid json")
	}

	fmt.Println(j.ToXML())

	util.RespondwithJSON(w, http.StatusCreated, mapString)
}

// TestGJSON Function
func TestGJSON(w http.ResponseWriter, r *http.Request) {
	var output map[string]interface{}

	// json := `{"name":{"first":"Janet","last":"Prichard"},"age":47}`

	bodyBytes, _ := ioutil.ReadAll(r.Body)
	bodyString := string(bodyBytes)

	grab := "cuicoperationRequest.payload.cdata.ucsUuidPool.accountName"
	value := gjson.Get(bodyString, grab)

	gJSON := value.String()

	json.Unmarshal([]byte(gJSON), &output)

	util.RespondwithJSON(w, http.StatusCreated, output)
}

// TestAPIAuth Function
func TestAPIAuth(w http.ResponseWriter, r *http.Request) {
	var authDetails struct {
		DeviceAccount string `json:"deviceAccount"`
		APICall       string `json:"apiCall"`
	}
	var targetDevice model.DeviceAccount
	var targetAPI model.API
	var resp *http.Response

	jsonErr := json.NewDecoder(r.Body).Decode(&authDetails)

	if jsonErr != nil {
		fmt.Println(jsonErr)
		util.RespondWithError(w, http.StatusBadRequest, "invalid json request")
	}

	targetFilter := primitive.M{
		"name": authDetails.DeviceAccount,
	}

	apiFilter := primitive.M{
		"name": authDetails.APICall,
	}

	deviceResult, deviceDBErr := database.FindOne("accounts", "devices", targetFilter)

	if deviceDBErr != nil {
		fmt.Println(deviceDBErr)
		util.RespondWithError(w, http.StatusBadRequest, "no such device")
		return
	}

	deviceDecodeErr := mapstructure.Decode(deviceResult, &targetDevice)

	if deviceDecodeErr != nil {
		fmt.Println(deviceDecodeErr)
		util.RespondWithError(w, http.StatusBadRequest, "error decoding device")
	}

	apiResult, apiDBErr := database.FindOne("apis", authDetails.DeviceAccount, apiFilter)

	if apiDBErr != nil {
		fmt.Println(apiDBErr)
		util.RespondWithError(w, http.StatusBadRequest, "no such api")
		return
	}

	apiDecodeErr := mapstructure.Decode(apiResult, &targetAPI)

	if apiDecodeErr != nil {
		fmt.Println(apiDecodeErr)
		util.RespondWithError(w, http.StatusBadRequest, "error decoding api")
	}

	switch targetDevice.AuthType {
	case "none":
		resp = auth.NoAuth(targetDevice, targetAPI)
	case "basic":
		resp = auth.BasicAuth(targetDevice, targetAPI)
	default:
		fmt.Println("Invalid AuthType!")
		return
	}

	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)
	bodyObject := make(map[string]interface{})

	respErr := json.Unmarshal(respBody, &bodyObject)

	if respErr != nil {
		fmt.Println(respErr)
		util.RespondWithError(w, http.StatusBadRequest, "error parsing response body")
	}

	util.RespondwithJSON(w, http.StatusOK, bodyObject)
}
