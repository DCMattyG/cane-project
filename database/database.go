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

// Save Function
func Save(database string, collection string, insertVal interface{}) (interface{}, error) {
	var id interface{}

	SelectDatabase(database, collection)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := db.InsertOne(ctx, &insertVal)

	if err != nil {
		return id, err
	}

	id = res.InsertedID

	return id, nil
}

// FindDB Function
func FindDB(filter bson.M) ([]string, error) {
	var result []string

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := client.ListDatabaseNames(ctx, filter)

	if err != nil {
		return result, err
	}

	return result, nil
}

// FindOne Function
func FindOne(database string, collection string, filter bson.M) (primitive.M, error) {
	var result primitive.M

	SelectDatabase(database, collection)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := db.FindOne(ctx, filter).Decode(&result)

	return result, err
}

// FindAll Function
func FindAll(database string, collection string, filter bson.M) ([]primitive.M, error) {
	var results []primitive.M

	SelectDatabase(database, collection)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cur, err := db.Find(ctx, filter)

	if err != nil {
		return results, err
	}

	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var result primitive.M

		err := cur.Decode(&result)

		if err != nil {
			return results, err
		}

		results = append(results, result)
	}

	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}

	return results, nil
}

// FindAndUpdate Function
func FindAndUpdate(database string, collection string, filter bson.M, update bson.M) (primitive.M, error) {
	var result primitive.M

	SelectDatabase(database, collection)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := db.FindOneAndUpdate(ctx, filter, update).Decode(&result)

	if err != nil {
		return result, err
	}

	return result, nil
}

func init() {
	var err error

	fmt.Print("Creating database connection...")

	client, err = mongo.NewClient("mongodb://localhost:27017")

	if err != nil {
		fmt.Println("[FAIL]")
		log.Fatal(err)
	} else {
		fmt.Println("[SUCCESS]")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	fmt.Print("Connecting to database...")

	err = client.Connect(ctx)

	if err != nil {
		fmt.Println("[FAIL]")
		log.Fatal(err)
	} else {
		fmt.Println("[SUCCESS]")
	}

	fmt.Print("Pinging database...")

	err = client.Ping(ctx, readpref.Primary())

	if err != nil {
		fmt.Println("[FAIL]")
		log.Fatal(err)
	} else {
		fmt.Println("[SUCCESS]")
	}
}
