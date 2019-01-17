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
	Router.Post("/updateToken", account.RefreshToken)
	Router.Get("/apiTest", TestCallAPI)

	// Private Default Routes
	Router.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(auth.TokenAuth))
		r.Use(jwtauth.Authenticator)
		r.Post("/changePassword", account.ChangePassword)
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

	if err != nil {
		fmt.Println(err)
		util.RespondWithError(w, http.StatusBadRequest, "invalid data")
		return
	}

	if model.IsJSON(string(bodyReader)) {
		var j model.JSONNode

		jsonErr := json.Unmarshal(bodyReader, &j)

		if jsonErr != nil {
			panic(jsonErr)
		}

		j.JSONVars()

		util.RespondwithJSON(w, http.StatusCreated, j)
	}

	if model.IsXML(string(bodyReader)) {
		var x model.XMLNode

		buf := bytes.NewBuffer(bodyReader)
		dec := xml.NewDecoder(buf)
		xmlErr := dec.Decode(&x)

		if xmlErr != nil {
			panic(xmlErr)
		}

		x.ScrubXML()
		x.XMLVars()

		util.RespondwithXML(w, http.StatusCreated, x)
	}
}

// TestPost function
func TestPost(w http.ResponseWriter, r *http.Request) {
	util.RespondwithJSON(w, http.StatusCreated, map[string]string{"message": "test post"})
}

// AddRoutes function
func AddRoutes(w http.ResponseWriter, r *http.Request) {
	var target model.RouteValue

	bodyReader, err := ioutil.ReadAll(r.Body)

	if err != nil {
		fmt.Println(err)
		util.RespondWithError(w, http.StatusBadRequest, "invalid data")
		return
	}

	err = util.UnmarshalJSON(bodyReader, &target)

	if err != nil {
		fmt.Println(err)
		util.RespondWithError(w, http.StatusBadRequest, "unmarshall failed")
		return
	}

	if !(ValidateRoute(target)) {
		util.RespondWithError(w, http.StatusBadRequest, "invalid route")
		return
	}

	fmt.Println("Adding routes to database...")

	postID, _ := database.Save("routing", "routes", target)

	fmt.Print("Inserted ID: ")
	fmt.Println(postID)

	Routers()

	util.RespondwithJSON(w, http.StatusCreated, map[string]string{"message": "routes added"})
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
	var respBody map[string]interface{}

	res := api.CallAPI()

	json.NewDecoder(res.Body).Decode(&respBody)
	fmt.Println(respBody)

	util.RespondwithJSON(w, http.StatusCreated, respBody)
}
