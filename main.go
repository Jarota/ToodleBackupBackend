package main

import (
	"context"
	"fmt"
	"log"

	// "go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/gofiber/fiber"

	"github.com/jarota/toodle-backup/user"
	"github.com/jarota/toodle-backup/handlers"
)

func main() {
	fmt.Println("Starting Toodle Backup Backend...")

	toBackup := []string{"tasks"}
	u := user.NewUser("james", "pass", "token", toBackup)
	u.Print()

	// Set client options
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")
	fmt.Println(client)
	
	// collection := client.Database("toodleBackup").Collection("users")

	app := fiber.New()
	app.Get("/", handlers.HelloWorld)
	app.Post("/register", handlers.Register)
	app.Post("/login", handlers.Login)
	app.Post("/logout", handlers.Logout)
	app.Put("/connCloudStorage/:name", handlers.ConnCloudStorage)

	app.Listen(3000)

}