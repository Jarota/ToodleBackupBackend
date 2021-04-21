package db

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ErrNotConnectedToMongoDB - Error to throw when not connected to mongodb
var ErrNotConnectedToMongoDB = errors.New("Error: Not connected to MongoDB")

// ConnectToMongoDB - function for connecting to mongodb
func ConnectToMongoDB() *mongo.Client {

	pwd := os.Getenv("MONGOPASS")

	clientOptions := options.Client().ApplyURI("mongodb://127.0.0.1:27017").
		SetAuth(options.Credential{
			AuthSource: "admin", Username: "toodle", Password: pwd,
		})

	client, err := mongo.Connect(context.Background(), clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.TODO(), nil)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB")
	return client
}

// GetCollection by name c
func GetCollection(client *mongo.Client, db, c string) (*mongo.Collection, error) {

	if client == nil {
		return nil, ErrNotConnectedToMongoDB
	}

	collection := client.Database(db).Collection(c)
	return collection, nil
}
