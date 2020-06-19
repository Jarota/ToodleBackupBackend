package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gofiber/fiber"

	"github.com/jarota/toodle-backup/db"
	"github.com/jarota/toodle-backup/handlers"
	"github.com/jarota/toodle-backup/user"
)

func main() {
	fmt.Println("Starting Toodle Backup Backend...")

	db.ConnectToMongoDB()
	fmt.Println("Connected to MongoDB")

	u := user.New("james", "pass")
	u.Print()
	

	app := fiber.New()
	app.Get("/", handlers.HelloWorld)
	app.Post("/register", handlers.Register)
	app.Post("/login", handlers.Login)
	app.Post("/logout", handlers.Logout)
	app.Put("/connToodledo", handlers.ConnToodledo)
	app.Put("/connCloudStorage", handlers.ConnCloudStorage)

	app.Listen(3000)

	err := db.Client.Disconnect(context.TODO())

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Disconnected from MongoDB")

}