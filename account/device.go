package account

import (
	"cane-project/database"
	"cane-project/model"
	"cane-project/util"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/mitchellh/mapstructure"

	"github.com/go-chi/chi"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"github.com/tidwall/gjson"
)

// AddDevice Function
func AddDevice(w http.ResponseWriter, r *http.Request) {
	var device model.DeviceAccount
	var authErr error
	// var authSave map[string]interface{}

	bodyBytes, bodyErr := ioutil.ReadAll(r.Body)
	target := string(bodyBytes)

	// jsonErr := json.NewDecoder(r.Body).Decode(&target)

	if bodyErr != nil {
		fmt.Println(bodyErr)
		util.RespondWithError(w, http.StatusBadRequest, "error decoding body")
		return
	}

	authType := gjson.Get(target, "device.authtype").String()
	authInfo := gjson.Get(target, "auth").Value()
	deviceInfo := gjson.Get(target, "device").Value()

	switch authType {
	case "none":
		//No Auth
	case "basic":
		var basicAuth model.BasicAuth
		authErr = mapstructure.Decode(authInfo, &basicAuth)
	case "session":
		var sessionAuth model.SessionAuth
		authErr = mapstructure.Decode(authInfo, &sessionAuth)
	case "apikey":
		var apiKeyAuth model.APIKeyAuth
		authErr = mapstructure.Decode(authInfo, &apiKeyAuth)
	case "rfc3447":
		var rfc3447Auth model.BasicAuth
		authErr = mapstructure.Decode(authInfo, &rfc3447Auth)
	default:
		util.RespondWithError(w, http.StatusBadRequest, "invalid auth type")
		return
	}

	if authErr != nil {
		fmt.Println(authErr)
		util.RespondWithError(w, http.StatusBadRequest, "invalid auth details")
		return
	}

	marshalErr := mapstructure.Decode(deviceInfo, &device)

	if marshalErr != nil {
		fmt.Println(marshalErr)
		util.RespondWithError(w, http.StatusBadRequest, "invalid device details")
		return
	}

	filter := primitive.M{
		"name": device.Name,
	}

	_, findErr := database.FindOne("accounts", "devices", filter)

	if findErr == nil {
		fmt.Println(findErr)
		util.RespondWithError(w, http.StatusBadRequest, "existing account")
		return
	}

	authID, _ := database.Save("auth", authType, authInfo)
	device.AuthObj = authID.(primitive.ObjectID)

	deviceID, _ := database.Save("accounts", "devices", device)
	device.ID = deviceID.(primitive.ObjectID)

	fmt.Print("Inserted Auth ID: ")
	fmt.Println(authID.(primitive.ObjectID).Hex())
	fmt.Print("Inserted Device ID: ")
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
