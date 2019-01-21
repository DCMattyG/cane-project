package account

import (
	"cane-project/database"
	"cane-project/model"
	"cane-project/util"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
)

// AddDevice Function
func AddDevice(w http.ResponseWriter, r *http.Request) {
	var target model.DeviceAccount

	json.NewDecoder(r.Body).Decode(&target)

	filter := primitive.M{
		"name": target.Name,
	}

	_, findErr := database.FindOne("accounts", "devices", filter)

	if findErr == nil {
		fmt.Println(findErr)
		util.RespondWithError(w, http.StatusBadRequest, "existing account")
		return
	}

	deviceID, _ := database.Save("accounts", "devices", target)
	target.ID = deviceID.(primitive.ObjectID)

	fmt.Print("Inserted ID: ")
	fmt.Println(deviceID.(primitive.ObjectID).Hex())

	foundVal, _ := database.FindOne("accounts", "devices", filter)

	util.RespondwithJSON(w, http.StatusCreated, foundVal)
}

// LoadDevice Function
func LoadDevice(w http.ResponseWriter, r *http.Request) {
	filter := primitive.M{
		"name": chi.URLParam(r, "name"),
	}

	foundVal, foundErr := database.FindOne("accounts", "devices", filter)

	if foundErr != nil {
		fmt.Println(foundErr)
		util.RespondWithError(w, http.StatusBadRequest, "device not found")
		return
	}

	util.RespondwithJSON(w, http.StatusOK, foundVal)
}

// UpdateDevice Function
func UpdateDevice(w http.ResponseWriter, r *http.Request) {
	var details map[string]interface{}

	json.NewDecoder(r.Body).Decode(&details)

	fmt.Println(details)

	filter := primitive.M{
		"name": chi.URLParam(r, "name"),
	}

	for key := range details {
		if key != "name" {
			filter[key] = primitive.M{"$exists": true}
		}
	}

	update := primitive.M{
		"$set": details,
	}

	updateVal, updateErr := database.FindAndUpdate("accounts", "devices", filter, update)

	if updateErr != nil {
		fmt.Println(updateErr)
		util.RespondWithError(w, http.StatusBadRequest, "device not found")
		return
	}

	util.RespondwithJSON(w, http.StatusOK, updateVal)
}

// ListDevices Function
func ListDevices(w http.ResponseWriter, r *http.Request) {
	var devices []string

	foundVal, foundErr := database.FindAll("accounts", "devices", primitive.M{})

	if foundErr != nil {
		fmt.Println(foundErr)
		util.RespondWithError(w, http.StatusBadRequest, "device not found")
		return
	}

	for key := range foundVal {
		devices = append(devices, foundVal[key]["name"].(string))
	}

	util.RespondwithJSON(w, http.StatusOK, map[string][]string{"devices": devices})
}

// ListDeviceAPIs Function
func ListDeviceAPIs(w http.ResponseWriter, r *http.Request) {
	device := chi.URLParam(r, "device")

	var apis []string

	foundVal, foundErr := database.FindAll("apis", device, primitive.M{})

	if foundErr != nil {
		fmt.Println(foundErr)
		util.RespondWithError(w, http.StatusBadRequest, "device not found")
		return
	}

	for key := range foundVal {
		apis = append(apis, foundVal[key]["name"].(string))
	}

	util.RespondwithJSON(w, http.StatusOK, map[string][]string{"apis": apis})
}
