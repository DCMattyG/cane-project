package routing

import (
	"cane/database"
	"cane/utils"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/mitchellh/mapstructure"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
)

// Router Variable
var Router *chi.Mux

// Routers Function
func Routers() {
	var iterVals []database.RouteValue
	Router = chi.NewMux()

	database.SelectDatabase("routing", "routes")

	filter := primitive.M{}
	foundVals := database.FindAllInDB(filter)
	mapstructure.Decode(foundVals, &iterVals)

	fmt.Println("Updating routes...")

	Router.Post("/add", AddRoutes)

	for i := range iterVals {
		routeVal := iterVals[i]

		if routeVal.Enable {
			newRoute := "/v" + strconv.Itoa(routeVal.Version) + "/" + routeVal.Category + "/" + routeVal.Route
			newMessage := routeVal.Message

			if routeVal.Verb == "get" {
				Router.Get(newRoute, func(w http.ResponseWriter, r *http.Request) {
					utils.RespondwithJSON(w, http.StatusCreated, newMessage)
				})
			} else if routeVal.Verb == "post" {
				Router.Post(newRoute, func(w http.ResponseWriter, r *http.Request) {
					testJSON := make(map[string]interface{})
					err := json.NewDecoder(r.Body).Decode(&testJSON)

					fmt.Println(testJSON)

					if err != nil {
						panic(err)
					}

					utils.RespondwithJSON(w, http.StatusCreated, testJSON)
				})
			}
		}
	}
}

// TestPost function
func TestPost(w http.ResponseWriter, r *http.Request) {
	utils.RespondwithJSON(w, http.StatusCreated, map[string]string{"message": "test post"})
}

// AddRoutes function
func AddRoutes(w http.ResponseWriter, r *http.Request) {
	var target database.RouteValue

	bodyReader, err := ioutil.ReadAll(r.Body)

	if err != nil {
		fmt.Println(err)
		utils.RespondWithError(w, http.StatusBadRequest, "invalid data")
		return
	}

	err = utils.UnmarshalJSON(bodyReader, &target)

	if err != nil {
		fmt.Println(err)
		utils.RespondWithError(w, http.StatusBadRequest, "unmarshall failed")
		return
	}

	fmt.Println("Adding routes to database...")

	database.SelectDatabase("routing", "routes")

	postID := database.InsertToDB(target)

	fmt.Print("Inserted ID: ")
	fmt.Println(postID)

	Routers()

	utils.RespondwithJSON(w, http.StatusCreated, map[string]string{"message": "routes added"})
}

func init() {
	Router = chi.NewRouter()

	// catch(err)
}
