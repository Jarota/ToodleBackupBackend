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

var ErrNotConnectedToMongoDB = errors.New("Error: Not connected to MongoDB")

func ConnectToMongoDB() *mongo.Client {

	pwd := os.Getenv("MONGOP")

	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017").
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

func GetCollection(client *mongo.Client, db, c string) (*mongo.Collection, error) {

	if client == nil {
		return nil, ErrNotConnectedToMongoDB
	}

	collection := client.Database(db).Collection(c)
	return collection, nil
}
