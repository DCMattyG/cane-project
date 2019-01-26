package account

import (
	"cane-project/database"
	"cane-project/model"
	"cane-project/util"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/fatih/structs"

	"github.com/mongodb/mongo-go-driver/mongo/options"

	"github.com/mitchellh/mapstructure"

	"github.com/go-chi/chi"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"github.com/tidwall/gjson"
)

// CreateDevice Function
func CreateDevice(w http.ResponseWriter, r *http.Request) {
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

	authType := gjson.Get(target, "device.authType").String()
	authInfo := gjson.Get(target, "auth").Value()
	deviceInfo := gjson.Get(target, "device").Value()

	switch authType {
	case "none":
		authInfo = primitive.M{}
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

	fmt.Print("Inserted Auth ID: ")
	fmt.Println(authID.(primitive.ObjectID).Hex())

	deviceID, _ := database.Save("accounts", "devices", device)
	device.ID = deviceID.(primitive.ObjectID)

	fmt.Print("Inserted Device ID: ")
	fmt.Println(deviceID.(primitive.ObjectID).Hex())

	_, reloadErr := database.FindOne("accounts", "devices", filter)

	if reloadErr != nil {
		fmt.Println(reloadErr)
		util.RespondWithError(w, http.StatusBadRequest, "error reloading account")
		return
	}

	util.RespondwithString(w, http.StatusCreated, "")
}

// UpdateDevice Function
func UpdateDevice(w http.ResponseWriter, r *http.Request) {
	var device model.DeviceAccount
	var newAuth map[string]interface{}
	var authErr error

	bodyBytes, bodyErr := ioutil.ReadAll(r.Body)
	target := string(bodyBytes)

	if bodyErr != nil {
		fmt.Println(bodyErr)
		util.RespondWithError(w, http.StatusBadRequest, "error decoding body")
		return
	}

	if !gjson.Get(target, "device.name").Exists() {
		fmt.Println("invalid device name")
		util.RespondWithError(w, http.StatusBadRequest, "invalid device name")
		return
	}

	deviceFilter := primitive.M{
		"name": gjson.Get(target, "device.name").String(),
	}

	loadDevice, loadErr := database.FindOne("accounts", "devices", deviceFilter)

	if loadErr != nil {
		fmt.Println(loadErr)
		util.RespondWithError(w, http.StatusBadRequest, "no such device")
		return
	}

	mapstructure.Decode(loadDevice, &device)

	device.IP = gjson.Get(target, "device.ip").String()

	authType := gjson.Get(target, "device.authType").String()
	authInfo := gjson.Get(target, "auth").Value()
	updatedAuth := gjson.Get(target, "auth").Map()

	if device.AuthType != authType {
		if !gjson.Get(target, "auth").Exists() {
			fmt.Println("cannot change authType without providing new auth info")
			util.RespondWithError(w, http.StatusBadRequest, "authType change without new auth body")
			return
		}

		deleteFilter := primitive.M{
			"_id": device.AuthObj,
		}

		deleteErr := database.Delete("auth", device.AuthType, deleteFilter)

		if deleteErr != nil {
			fmt.Println(deleteErr)
			util.RespondWithError(w, http.StatusBadRequest, "error deleting old auth")
			return
		}

		// deviceInfo := gjson.Get(target, "device").Value()

		switch authType {
		case "none":
			authInfo = primitive.M{}
		case "basic":
			var basicAuth model.BasicAuth
			authErr = mapstructure.Decode(authInfo, &basicAuth)
			newAuth = structs.Map(basicAuth)
		case "session":
			var sessionAuth model.SessionAuth
			authErr = mapstructure.Decode(authInfo, &sessionAuth)
			newAuth = structs.Map(sessionAuth)
		case "apikey":
			var apiKeyAuth model.APIKeyAuth
			authErr = mapstructure.Decode(authInfo, &apiKeyAuth)
			newAuth = structs.Map(apiKeyAuth)
		case "rfc3447":
			var rfc3447Auth model.Rfc3447Auth
			authErr = mapstructure.Decode(authInfo, &rfc3447Auth)
			newAuth = structs.Map(rfc3447Auth)
		default:
			util.RespondWithError(w, http.StatusBadRequest, "invalid auth type")
			return
		}

		fmt.Println("NEWAUTH: ", newAuth)

		if authErr != nil {
			fmt.Println(authErr)
			util.RespondWithError(w, http.StatusBadRequest, "invalid auth details")
			return
		}

		authID, authIDErr := database.Save("auth", authType, newAuth)

		if authIDErr != nil {
			fmt.Println(authIDErr)
			util.RespondWithError(w, http.StatusBadRequest, "error saving replaced auth")
			return
		}

		device.AuthObj = authID.(primitive.ObjectID)
		device.AuthType = authType
	} else if gjson.Get(target, "auth").Exists() {
		authFilter := primitive.M{
			"_id": device.AuthObj,
		}

		loadAuth, loadAuthErr := database.FindOne("auth", device.AuthType, authFilter)

		if loadAuthErr != nil {
			fmt.Println(loadAuthErr)
			util.RespondWithError(w, http.StatusBadRequest, "no such auth")
			return
		}

		for k := range loadAuth {
			loadAuth[k] = updatedAuth[k]
		}

		delete(loadAuth, "_id")

		_, replaceAuthErr := database.FindAndReplace("auth", device.AuthType, authFilter, loadAuth)

		if replaceAuthErr != nil {
			fmt.Println(replaceAuthErr)
			util.RespondWithError(w, http.StatusBadRequest, "error saving updated auth")
			return
		}
	}

	_, updatedErr := database.FindAndReplace("accounts", "devices", deviceFilter, structs.Map(device))

	if updatedErr != nil {
		fmt.Println(updatedErr)
		util.RespondWithError(w, http.StatusBadRequest, "error saving updated device")
		return
	}

	// foundVal, _ := database.FindOne("accounts", "devices", filter)

	util.RespondwithString(w, http.StatusOK, "")
}

// DeleteDevice Function
func DeleteDevice(w http.ResponseWriter, r *http.Request) {
	deviceFilter := primitive.M{
		"name": chi.URLParam(r, "devicename"),
	}

	findVal, findErr := database.FindOne("accounts", "devices", deviceFilter)

	if findErr != nil {
		fmt.Println(findErr)
		util.RespondWithError(w, http.StatusBadRequest, "device not found")
		return
	}

	deleteDeviceErr := database.Delete("accounts", "devices", deviceFilter)

	if deleteDeviceErr != nil {
		fmt.Println(deleteDeviceErr)
		util.RespondWithError(w, http.StatusBadRequest, "user not found")
		return
	}

	authFilter := primitive.M{
		"_id": findVal["authObj"].(primitive.ObjectID),
	}

	deleteAuthErr := database.Delete("auth", findVal["authType"].(string), authFilter)

	if deleteAuthErr != nil {
		fmt.Println(deleteAuthErr)
		util.RespondWithError(w, http.StatusBadRequest, "error deleting device auth")
		return
	}

	util.RespondwithString(w, http.StatusOK, "")
}

// ValidateAuthObj
// func ValidateAuthObj(authObj map[string]interface{}) (string, map[string]interface{}) {
// 	var authErr error
// 	var authObj map[string]interface{}

// 	authType := gjson.Get(authInfo, "device.authtype").String()

// 	switch authType {
// 	case "none":
// 		authInfo = make(map[string]interface{})
// 	case "basic":
// 		var basicAuth model.BasicAuth
// 		authErr = mapstructure.Decode(authInfo, &basicAuth)
// 		authObj = structs.Map(basicAuth)
// 	case "session":
// 		var sessionAuth model.SessionAuth
// 		authErr = mapstructure.Decode(authInfo, &sessionAuth)
// 		authObj = structs.Map(sessionAuth)
// 	case "apikey":
// 		var apiKeyAuth model.APIKeyAuth
// 		authErr = mapstructure.Decode(authInfo, &apiKeyAuth)
// 		authObj = structs.Map(apiKeyAuth)
// 	case "rfc3447":
// 		var rfc3447Auth model.BasicAuth
// 		authErr = mapstructure.Decode(authInfo, &rfc3447Auth)
// 		authObj = structs.Map(rfc3447Auth)
// 	default:
// 		util.RespondWithError(w, http.StatusBadRequest, "invalid auth type")
// 		return nil, nil
// 	}
// }

// GetDevice Function
func GetDevice(w http.ResponseWriter, r *http.Request) {
	var authType string

	filter := primitive.M{
		"name": chi.URLParam(r, "devicename"),
	}

	findDeviceVal, findDeviceErr := database.FindOne("accounts", "devices", filter)

	if findDeviceErr != nil {
		fmt.Println(findDeviceErr)
		util.RespondWithError(w, http.StatusBadRequest, "device not found")
		return
	}

	authType = findDeviceVal["authType"].(string)

	filter = primitive.M{
		"_id": findDeviceVal["authObj"],
	}

	findAuthVal, findAuthErr := database.FindOne("auth", authType, filter)

	if findAuthErr != nil {
		fmt.Println(findAuthErr)
		util.RespondWithError(w, http.StatusBadRequest, "auth object not found")
		return
	}

	delete(findAuthVal, "_id")
	delete(findDeviceVal, "_id")
	delete(findDeviceVal, "authObj")

	response := map[string]interface{}{
		"device": findDeviceVal,
		"auth":   findAuthVal,
	}

	util.RespondwithJSON(w, http.StatusOK, response)
}

// GetDevices Function
func GetDevices(w http.ResponseWriter, r *http.Request) {
	var opts options.FindOptions
	var deviceList []string

	projection := primitive.M{
		"_id":  0,
		"name": 1,
	}

	opts.SetProjection(projection)

	findVals, findErr := database.FindAll("accounts", "devices", primitive.M{}, opts)

	if findErr != nil {
		fmt.Println(findErr)
		util.RespondWithError(w, http.StatusBadRequest, "no devices found")
		return
	}

	for _, device := range findVals {
		deviceList = append(deviceList, device["name"].(string))
	}

	util.RespondwithJSON(w, http.StatusOK, map[string][]string{"devices": deviceList})
}

// GetDeviceID Function
func GetDeviceID(deviceName string) (primitive.ObjectID, error) {
	var deviceID primitive.ObjectID

	filter := primitive.M{
		"name": deviceName,
	}

	findVal, findErr := database.FindOne("accounts", "devices", filter)

	if findErr != nil {
		fmt.Println(findErr)
		return deviceID, findErr
	}

	deviceID = findVal["_id"].(primitive.ObjectID)

	return deviceID, nil
}

// UpdateDevice Function
// func UpdateDevice(w http.ResponseWriter, r *http.Request) {
// 	var deviceDetails map[string]interface{}
// 	var updatedDevice model.DeviceAccount

// 	json.NewDecoder(r.Body).Decode(&deviceDetails)

// 	filter := primitive.M{
// 		"name": chi.URLParam(r, "devicename"),
// 	}

// 	findVal, findErr := database.FindOne("accounts", "devices", filter)

// 	if findErr != nil {
// 		fmt.Println(findErr)
// 		util.RespondWithError(w, http.StatusBadRequest, "device not found")
// 		return
// 	}

// 	mapstructure.Decode(findVal, &updatedDevice)

// 	updatedDevice.Name = deviceDetails["name"].(string)
// 	updatedDevice.IP = deviceDetails["ip"].(string)
// 	updatedDevice.AuthType = deviceDetails["authType"].(string)

// 	_, replaceErr := database.ReplaceOne("accounts", "users", filter, structs.Map(updatedUser))

// 	if replaceErr != nil {
// 		fmt.Println(replaceErr)
// 		util.RespondWithError(w, http.StatusBadRequest, "error updating user")
// 		return
// 	}

// 	util.RespondwithJSON(w, http.StatusOK, updatedUser)
// }

// GetDevices Function
// func GetDevices(w http.ResponseWriter, r *http.Request) {
// 	var opts options.FindOptions
// 	var deviceList []string

// 	findVal, findErr := database.FindAll("accounts", "devices", primitive.M{}, opts)

// 	if findErr != nil {
// 		fmt.Println(findErr)
// 		util.RespondWithError(w, http.StatusBadRequest, "device not found")
// 		return
// 	}

// 	for key := range findVal {
// 		deviceList = append(deviceList, findVal[key]["name"].(string))
// 	}

// 	util.RespondwithJSON(w, http.StatusOK, map[string][]string{"devices": deviceList})
// }

// GetDeviceFromDB Function
func GetDeviceFromDB(deviceName string) (model.DeviceAccount, error) {
	var device model.DeviceAccount

	filter := primitive.M{
		"name": deviceName,
	}

	findVal, findErr := database.FindOne("accounts", "devices", filter)

	if findErr != nil {
		fmt.Println(findErr)
		return device, findErr
	}

	mapErr := mapstructure.Decode(findVal, &device)

	if mapErr != nil {
		fmt.Println(mapErr)
		return device, mapErr
	}

	return device, nil
}
