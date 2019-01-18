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
		"name": api.Account,
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

	_, existErr := database.FindOne("apis", api.Account, existFilter)

	if existErr == nil {
		fmt.Println(existErr)
		util.RespondWithError(w, http.StatusBadRequest, "api already exists")
		return
	}

	saveID, saveErr := database.Save("apis", api.Account, api)

	if saveErr != nil {
		fmt.Println(saveErr)
		util.RespondWithError(w, http.StatusBadRequest, "error saving api")
		return
	}

	api.ID = saveID.(primitive.ObjectID)

	fmt.Print("Inserted ID: ")
	fmt.Println(saveID.(primitive.ObjectID).Hex())

	foundVal, _ := database.FindOne("apis", api.Account, existFilter)

	util.RespondwithJSON(w, http.StatusCreated, foundVal)
}

// CallAPI Function
func CallAPI() *http.Response {
	// transport := &http.Transport{}
	client := &http.Client{}

	/* Temp Variables */
	// host, err := url.Parse("https://intersight.com/api/v1")
	host, err := url.Parse("https://deckofcardsapi.com/api/deck/new/")

	if err != nil {
		panic("Cannot parse *host*!")
	}

	method := "GET"
	// resourcePath := "/ntp/Policies"
	resourcePath := ""
	targetURL := host.String() + resourcePath
	// var bodyString string
	// proxy := "http://proxy.esl.cisco.com"
	// proxy := ""
	// proxyURL, err := url.Parse(proxy)
	// queryPath := ""
	// var requestHeader map[string]string
	/* End Temp Variables */

	// Create HTTP request
	fmt.Println("Method: ", method)
	fmt.Println("TargetURL: ", targetURL)
	// fmt.Println("Body: ", strings.NewReader(bodyString))
	req, err := http.NewRequest(method, host.String(), nil)

	if err != nil {
		log.Print(err)
		fmt.Println("Errored when creating the HTTP request!")
		return nil
	}

	// Append headers to HTTP request
	// for key, value := range requestHeader {
	// 	req.Header.Add(key, value)
	// }

	// Add proxy settings to the HTTP Transport object
	// if len(proxyURL.RawPath) > 0 {
	// 	transport = &http.Transport{
	// 		// Proxy: http.ProxyURL(proxyURL),
	// 	}

	// 	client = &http.Client{
	// 		Transport: transport,
	// 	}
	// }

	// Add query params and call HTTP request
	// req.URL.RawQuery = queryPath
	resp, err := client.Do(req)

	if err != nil {
		log.Print(err)
		fmt.Println("Errored when sending request to the server!")
		return nil
	}

	fmt.Println("Response:")
	fmt.Println(resp)

	return resp
}
