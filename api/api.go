package api

import (
	"cane-project/database"
	"cane-project/model"
	"cane-project/util"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

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

// CallAPI Function
func CallAPI(req *http.Request, proxy string) *http.Response {
	transport := &http.Transport{}
	client := &http.Client{}

	// transport := &http.Transport{
	// 	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	// }
	// client := &http.Client{Transport: transport}

	// Verify the proxyURL is properly formatted
	proxyURL, err := url.Parse(proxy)
	if err != nil {
		fmt.Println("Invalid proxy URL format!")
	}

	// Ignore self-signed certificates
	// transport.TLSClientConfig.InsecureSkipVerify = true

	// client = &http.Client{
	// 	Transport: transport,
	// }

	// Add proxy settings to the HTTP Transport object
	if len(proxyURL.RawPath) > 0 {
		transport.Proxy = http.ProxyURL(proxyURL)

		// transport = &http.Transport{
		// 	Proxy: http.ProxyURL(proxyURL),
		// }

		client = &http.Client{
			Transport: transport,
		}
	}

	resp, err := client.Do(req)

	if err != nil {
		log.Print(err)
		fmt.Println("Errored when sending request to the server!")
		return nil
	}

	return resp
}
