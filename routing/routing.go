package routing

import (
	"bytes"
	"cane-project/account"
	"cane-project/api"
	"cane-project/auth"
	"cane-project/database"
	"cane-project/model"
	"cane-project/util"
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

	resp := api.CallAPI(callAPI)

	// json.NewDecoder(res.Body).Decode(&respBody)

	defer resp.Body.Close()
	respBody, _ := ioutil.ReadAll(resp.Body)

	// fmt.Println(string(respBody))

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
	var x model.XMLNode
	var j model.JSONNode

	bodyReader, _ := ioutil.ReadAll(r.Body)

	buf := bytes.NewBuffer(bodyReader)
	dec := xml.NewDecoder(buf)
	xmlErr := dec.Decode(&x)

	if xmlErr != nil {
		fmt.Println(xmlErr)
		util.RespondWithError(w, http.StatusBadRequest, "invalid xml")
		return
	}

	x.ScrubXML()

	mapString := x.XMLtoJSON()

	fmt.Println(x.XMLtoJSON())
	// fmt.Println("-------------------------------")

	jBytes, _ := json.MarshalIndent(mapString, "", "  ")

	jString := string(jBytes)

	fmt.Println(jString)

	fmt.Println("-------------------------------")

	json.Unmarshal([]byte(jString), &j)

	fmt.Println(j.ToXML())

	// testMap := map[string]interface{}{
	// 	"cuicoperationRequest": map[string]interface{}{
	// 		"payload": map[string]interface{}{
	// 			"cdata": map[string]interface{}{
	// 				"ucsUuidPool": map[string]interface{}{
	// 					"name": map[string]interface{}{
	// 						"data": "Test_UUID_Pool",
	// 					},
	// 					"descr": map[string]interface{}{
	// 						"data": "Test_UUID_Pool",
	// 					},
	// 					"prefix": map[string]interface{}{
	// 						"data": "other",
	// 					},
	// 					"otherPrefix": map[string]interface{}{
	// 						"data": "00000000-0000-0000",
	// 					},
	// 					"accountName": map[string]interface{}{
	// 						"data": "ucsm-248",
	// 					},
	// 					"org": map[string]interface{}{
	// 						"data": "org-root",
	// 					},
	// 					"firstMACAddress": map[string]interface{}{
	// 						"data": "0000-000000000001",
	// 					},
	// 					"size": map[string]interface{}{
	// 						"data": 1,
	// 					},
	// 				},
	// 			},
	// 		},
	// 	},
	// }

	// jsonMap, _ := json.Marshal(testMap)

	// fmt.Println(string(jsonMap))

	util.RespondwithJSON(w, http.StatusCreated, mapString)
}
