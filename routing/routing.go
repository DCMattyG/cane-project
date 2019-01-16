package routing

import (
	"bytes"
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
	"github.com/mitchellh/mapstructure"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
)

// RouteValue Struct
type RouteValue struct {
	ID       primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Enable   bool               `json:"enable" bson:"enable"`
	Verb     string             `json:"verb" bson:"verb"`
	Version  int                `json:"version" bson:"version"`
	Category string             `json:"category" bson:"category"`
	Route    string             `json:"route" bson:"route"`
	Message  map[string]string  `json:"message" bson:"message"`
}

// Router Variable
var Router *chi.Mux

func init() {
	Router = chi.NewRouter()

	// catch(err)
}

// Routers Function
func Routers() {
	var iterVals []RouteValue
	Router = chi.NewMux()

	database.SelectDatabase("routing", "routes")

	filter := primitive.M{}
	foundVals := database.FindAllInDB(filter)
	mapstructure.Decode(foundVals, &iterVals)

	fmt.Println("Updating routes...")

	// Built-In Default Routes
	Router.Post("/addRoute", AddRoutes)
	Router.Post("/parseVars", ParseVars)

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
					testJSON := make(map[string]interface{})
					err := json.NewDecoder(r.Body).Decode(&testJSON)

					fmt.Println(testJSON)

					if err != nil {
						panic(err)
					}

					util.RespondwithJSON(w, http.StatusCreated, testJSON)
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
	var target RouteValue

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

	// database.SelectDatabase("routing", "routes")

	postID := database.SaveToDB("routing", "routes", target)

	fmt.Print("Inserted ID: ")
	fmt.Println(postID)

	Routers()

	util.RespondwithJSON(w, http.StatusCreated, map[string]string{"message": "routes added"})
}

// ValidateRoute Function
func ValidateRoute(route RouteValue) bool {
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
