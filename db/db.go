package db

import (
	"context"
	"errors"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ErrNotConnectedToMongoDB - Error to throw when not connected to mongodb
var ErrNotConnectedToMongoDB = errors.New("error: not connected to mongodb")

// ConnectToMongoDB - function for connecting to mongodb
func ConnectToMongoDB(ctx context.Context) *mongo.Client {

	// pwd := os.Getenv("MONGOPASS")

	clientOptions := options.Client().ApplyURI("mongodb://127.0.0.1:27017") // .SetAuth(options.Credential{
	// AuthSource: "admin", Username: "toodle", Password: pwd,
	// })

	client, err := mongo.Connect(ctx, clientOptions)

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

// GetCollection by name
func GetCollection(client *mongo.Client, db, coll string) (*mongo.Collection, error) {

	if client == nil {
		return nil, ErrNotConnectedToMongoDB
	}

	collection := client.Database(db).Collection(coll)
	return collection, nil
}
