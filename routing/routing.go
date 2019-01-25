package routing

import (
	"bytes"
	"cane-project/account"
	"cane-project/api"
	"cane-project/database"
	"cane-project/jwt"
	"cane-project/model"
	"cane-project/util"
	"cane-project/workflow"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
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
	Router = chi.NewRouter()

	filter := primitive.M{}
	foundVals, _ := database.FindAll("routing", "routes", filter)
	mapstructure.Decode(foundVals, &iterVals)

	fmt.Println("Updating routes...")

	cors := cors.New(cors.Options{
		// AllowedOrigins: []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})

	Router.Use(cors.Handler)

	// Public Default Routes
	Router.Post("/login", account.Login)
	// Router.Post("/addRoute", AddRoutes)
	// Router.Post("/parseVars", ParseVars)
	// Router.Post("/addUser", account.AddUser)
	// Router.Post("/validateToken", account.ValidateUserToken)
	// Router.Patch("/updateToken/{user}", account.RefreshToken)
	// Router.Post("/addDevice", account.AddDevice)
	// Router.Get("/loadDevice/{name}", account.LoadDevice)
	// Router.Patch("/updateDevice/{name}", account.UpdateDevice)
	// Router.Get("/listDevice", account.ListDevices)
	// Router.Get("/deviceApis/{device}", account.ListDeviceAPIs)
	// Router.Post("/addApi", api.AddAPI)
	Router.Post("/apiTest", TestCallAPI)
	Router.Get("/testPath/*", TestPath)
	Router.Post("/testJSON", JSONTest)
	Router.Post("/testXML", XMLTest)
	Router.Post("/testGJSON", TestGJSON)
	// Router.Post("/testAPIAuth", TestAPIAuth)
	// Router.Post("/addWorkflow", workflow.AddWorkflow)
	// Router.Get("/listWorkflow", workflow.ListWorkflows)
	// Router.Get("/listWorkflow/{name}", workflow.LoadWorkflow)
	Router.Post("/callWorkflow/{name}", workflow.ExecuteWorkflow)
	Router.Get("/loadAPI/{account}/{name}", api.LoadAPI)
	Router.Get("/claimTest", ClaimTest)

	// Private Default Routes
	Router.Group(func(r chi.Router) {
		r.Use(cors.Handler)
		r.Use(jwtauth.Verifier(jwt.TokenAuth))
		r.Use(jwtauth.Authenticator)
		r.Post("/addRoute", AddRoutes)
		r.Post("/parseVars", ParseVars)
		r.Post("/addUser", account.AddUser)
		r.Post("/validateToken", account.ValidateUserToken)
		r.Patch("/updateToken/{user}", account.RefreshToken)
		r.Post("/addDevice", account.AddDevice)
		r.Get("/loadDevice/{name}", account.LoadDevice)
		r.Patch("/updateDevice/{name}", account.UpdateDevice)
		r.Get("/listDevice", account.ListDevices)
		r.Get("/deviceApis/{device}", account.ListDeviceAPIs)
		r.Post("/addApi", api.AddAPI)
		r.Post("/addWorkflow", workflow.AddWorkflow)
		r.Get("/listWorkflow", workflow.ListWorkflows)
		r.Get("/listWorkflow/{name}", workflow.LoadWorkflow)
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

// ClaimTest Function
func ClaimTest(w http.ResponseWriter, r *http.Request) {
	var testResult model.StepResult

	stepResults := make(map[string]model.StepResult)

	fmt.Println("Generating new claim...")
	claim := workflow.GenerateClaim()

	fmt.Println("Saving new claim...")
	claim.Save()

	fmt.Println("Loading fake step data...")

	testResult.APIAccount = "testaccount"
	testResult.APICall = "testcall"
	testResult.Error = errors.New("")
	testResult.ReqBody = "{req_body}"
	testResult.ResBody = "{res_body}"
	testResult.Status = 2

	fmt.Println("Assigning fake step data to fake results...")
	stepResults["1"] = testResult

	fmt.Println("Assigning fake results to claim...")
	claim.WorkflowResults = stepResults

	fmt.Println("Saving updated claim...")
	claim.Save()

	util.RespondwithJSON(w, http.StatusCreated, map[string]interface{}{"claim": claim})
}
