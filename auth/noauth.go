package auth

import (
	"cane-project/model"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// NoAuth Function
func NoAuth(account model.DeviceAccount, api model.API) *http.Response {
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

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		log.Print(err)
		fmt.Println("Errored when sending request to the server!")
	}

	return resp
}
