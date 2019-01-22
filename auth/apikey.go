package auth

import (
	"cane-project/database"
	"cane-project/model"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/mongodb/mongo-go-driver/bson/primitive"
)

// APIKeyAuth Function
func APIKeyAuth(account model.DeviceAccount, api model.API) (*http.Response, error) {
	host, err := url.Parse(account.IP)
	if err != nil {
		panic("Cannot parse *host*!")
	}

	fmt.Println("SCHEME: ", host.Scheme)
	fmt.Println("HOSTNAME: ", host.Hostname())
	fmt.Println("ENDPOINT: ", api.URL)

	targetMethod := strings.ToUpper(api.Method)

	fmt.Println("METHOD: ", targetMethod)

	targetURL := host.Scheme + "://" + host.Hostname() + api.URL

	fmt.Println("TARGETURL: ", targetURL)

	// Create HTTP request
	req, err := http.NewRequest(targetMethod, targetURL, strings.NewReader(""))

	if err != nil {
		log.Print(err)
		fmt.Println("Errored when creating the HTTP request!")
	}

	fmt.Println("REQ: ", req)

	filter := primitive.M{
		"_id": primitive.ObjectID(account.AuthObj),
	}

	foundVal, foundErr := database.FindOne("auth", account.AuthType, filter)

	if foundErr != nil {
		fmt.Println(foundErr)
		return nil, foundErr
	}

	apiHeader := foundVal["header"].(string)
	apiKey := foundVal["key"].(string)

	fmt.Println("APIHEADER: ", apiHeader)
	fmt.Println("APIKEY: ", apiKey)

	// Append headers to HTTP request
	req.Header.Add(apiHeader, apiKey)

	// client := &http.Client{}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: transport}

	resp, err := client.Do(req)

	if err != nil {
		log.Print(err)
		fmt.Println("Errored when sending request to the server!")
		return nil, err
	}

	return resp, nil
}
