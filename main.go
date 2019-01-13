package main

import (
	"cane/routing"
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
	// fmt.Println("Selecting database...")
	// database.SelectDatabase("testing", "numbers")

	// fmt.Println("Adding test post to database...")

	// toInsert := database.InsertValue{
	// 	Name:  "pi",
	// 	Value: 3.1459,
	// }

	// id := database.InsertToDB(toInsert)

	// fmt.Print("Inserted ID: ")
	// fmt.Println(id)

	routing.Routers()

	fmt.Println("Starting router...")
	http.ListenAndServe(":8005", logger())
}
