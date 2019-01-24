package api

import (
	"cane-project/auth"
	"cane-project/database"
	"cane-project/model"
	"cane-project/util"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-chi/chi"
	"github.com/mitchellh/mapstructure"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
)

// APITypes Variable
var APITypes []string

func init() {
	APITypes = []string{
		"xml",
		"json",
	}
}

// AddAPI Function
func AddAPI(w http.ResponseWriter, r *http.Request) {
	var api model.API

	json.NewDecoder(r.Body).Decode(&api)

	accountFilter := primitive.M{
		"name": api.DeviceAccount,
	}

	_, accountErr := database.FindOne("accounts", "devices", accountFilter)

	if accountErr != nil {
		fmt.Println(accountErr)
		util.RespondWithError(w, http.StatusBadRequest, "no such account")
		return
	}

	existFilter := primitive.M{
		"name": api.Name,
	}

	_, existErr := database.FindOne("apis", api.DeviceAccount, existFilter)

	if existErr == nil {
		fmt.Println(existErr)
		util.RespondWithError(w, http.StatusBadRequest, "api already exists")
		return
	}

	saveID, saveErr := database.Save("apis", api.DeviceAccount, api)

	if saveErr != nil {
		fmt.Println(saveErr)
		util.RespondWithError(w, http.StatusBadRequest, "error saving api")
		return
	}

	api.ID = saveID.(primitive.ObjectID)

	fmt.Print("Inserted ID: ")
	fmt.Println(saveID.(primitive.ObjectID).Hex())

	foundVal, _ := database.FindOne("apis", api.DeviceAccount, existFilter)

	util.RespondwithJSON(w, http.StatusCreated, foundVal)
}

// LoadAPI Function
func LoadAPI(w http.ResponseWriter, r *http.Request) {
	apiAccount := chi.URLParam(r, "account")
	apiName := chi.URLParam(r, "name")

	getAPI, getErr := GetAPIFromDB(apiAccount, apiName)

	if getErr != nil {
		fmt.Println(getErr)
		util.RespondWithError(w, http.StatusBadRequest, "api not found")
		return
	}

	util.RespondwithJSON(w, http.StatusOK, getAPI)
}

// GetAPIFromDB Function
func GetAPIFromDB(apiAccount string, apiName string) (model.API, error) {
	var api model.API

	filter := primitive.M{
		"name": apiName,
	}

	foundVal, foundErr := database.FindOne("apis", apiAccount, filter)

	if foundErr != nil {
		fmt.Println(foundErr)
		return api, foundErr
	}

	mapErr := mapstructure.Decode(foundVal, &api)

	if mapErr != nil {
		fmt.Println(mapErr)
		return api, mapErr
	}

	return api, nil
}

// CallAPI Function
func CallAPI(targetAPI model.API) (*http.Response, error) {
	transport := &http.Transport{}
	client := &http.Client{}

	var targetDevice model.DeviceAccount
	var req *http.Request
	var reqErr error

	proxyURL, err := url.Parse(util.ProxyURL)
	if err != nil {
		fmt.Println("Invalid proxy URL format!")
	}

	if util.IgnoreSSL {
		transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	// Add proxy settings to the HTTP Transport object
	if len(proxyURL.RawPath) > 0 {
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	client = &http.Client{Transport: transport}

	deviceFilter := primitive.M{
		"name": targetAPI.DeviceAccount,
	}

	deviceResult, deviceDBErr := database.FindOne("accounts", "devices", deviceFilter)

	if deviceDBErr != nil {
		fmt.Println(deviceDBErr)
		fmt.Println("Error loading device for CallAPI")
		return nil, deviceDBErr
	}

	deviceDecodeErr := mapstructure.Decode(deviceResult, &targetDevice)

	if deviceDecodeErr != nil {
		fmt.Println(deviceDecodeErr)
		fmt.Println("Error decoding device for CallAPI")
		return nil, deviceDecodeErr
	}

	switch targetDevice.AuthType {
	case "none":
		req, reqErr = auth.NoAuth(targetAPI)
	case "basic":
		req, reqErr = auth.BasicAuth(targetAPI)
	case "apikey":
		req, reqErr = auth.APIKeyAuth(targetAPI)
	default:
		fmt.Println("Invalid AuthType!")
		return nil, errors.New("Invalid AuthType")
	}

	if reqErr != nil {
		fmt.Println(reqErr)
		fmt.Println("Error getting request for CallAPI")
		return nil, reqErr
	}

	resp, respErr := client.Do(req)

	if respErr != nil {
		fmt.Println("Errored when sending request to the server!")
		return nil, respErr
	}

	return resp, nil
}
