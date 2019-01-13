package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mongodb/mongo-go-driver/bson/primitive"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/readpref"
)

var client *mongo.Client
var db *mongo.Collection

// SelectDatabase Function
func SelectDatabase(dbase string, coll string) {
	db = client.Database(dbase).Collection(coll)

	return
}

// SaveToDB Function
func SaveToDB(database string, collection string, insertVal interface{}) interface{} {
	SelectDatabase(database, collection)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := db.InsertOne(ctx, &insertVal)

	if err != nil {
		log.Fatal(err)
	}

	id := res.InsertedID

	return id
}

// FindDB Function
func FindDB(filter bson.M) []string {
	var result []string

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := client.ListDatabaseNames(ctx, filter)

	if err != nil {
		log.Fatal(err)
	}

	return result
}

// FindOneInDB Function
func FindOneInDB(filter bson.M) primitive.M {
	var result primitive.M

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := db.FindOne(ctx, filter).Decode(&result)

	if err != nil {
		log.Fatal(err)
	}

	return result
}

// FindAllInDB Function
func FindAllInDB(filter bson.M) []primitive.M {
	var results []primitive.M
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cur, err := db.Find(ctx, filter)

	if err != nil {
		log.Fatal(err)
	}

	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var result primitive.M

		err := cur.Decode(&result)

		if err != nil {
			log.Fatal(err)
		}

		results = append(results, result)
	}

	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}

	return results
}

func init() {
	var err error

	fmt.Println("Connecting to database...")

	client, err = mongo.NewClient("mongodb://localhost:27017")

	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	err = client.Connect(ctx)

	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, readpref.Primary())

	if err != nil {
		log.Fatal(err)
	}

	// dbs := FindDB(bson.M{"name": "routing"})

	// fmt.Println(dbs)
}
