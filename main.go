package main

import (
	"cane-project/database"
	"cane-project/routing"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/mongodb/mongo-go-driver/bson/primitive"
)

// LogMessage Struct
type LogMessage struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Timestamp time.Time          `json:"timestamp" bson:"timestamp"`
	Method    string             `json:"method" bson:"method"`
	URL       *url.URL           `json:"url" bson:"url"`
}

// Catch Function
func catch(err error) {
	if err != nil {
		panic(err)
	}
}

// Logger return log message
func logger() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		newLog := LogMessage{
			Timestamp: time.Now(),
			Method:    r.Method,
			URL:       r.URL,
		}

		database.SelectDatabase("logging", "logs")
		logID := database.InsertToDB(newLog)

		fmt.Print("Inserted Log: ")
		fmt.Println(logID)

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
