package db

import (
	"context"
	"errors"
	"log"

	// "go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client

var ErrNotConnectedToMongoDB = errors.New("Error: Not connected to MongoDB")

func ConnectToMongoDB() {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	Client, err := mongo.Connect(context.Background(), clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	err = Client.Ping(context.TODO(), nil)

	if err != nil {
		log.Fatal(err)
	}
}

func GetCollection(db, c string) (*mongo.Collection, error) {

	if Client == nil {
		return nil, ErrNotConnectedToMongoDB
	}

	collection := Client.Database(db).Collection(c)
	return collection, nil
}