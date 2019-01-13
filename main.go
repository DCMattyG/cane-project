package main

import (
	"cane-project/routing"
	"fmt"
	"net/http"
	"time"
)

// Catch Function
func catch(err error) {
	if err != nil {
		panic(err)
	}
}

// Logger return log message
func logger() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(time.Now(), r.Method, r.URL)
		routing.Router.ServeHTTP(w, r)
	})
}

// Main Function
func main() {
	routing.Routers()

	fmt.Println("Starting router...")
	http.ListenAndServe(":8005", logger())
}
