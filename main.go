package main

import (
	// "context"
	"fmt"
	// "log"
	"os"

	"github.com/gofiber/fiber"
	"github.com/gofiber/jwt"

	// "github.com/jarota/toodle-backup/db"
	"github.com/jarota/ToodleBackupBackend/handlers"
)

func main() {
	fmt.Println("Starting Toodle Backup Backend...")

	fmt.Println("Connected to MongoDB")

	app := fiber.New()

	app.Get("/", handlers.HelloWorld)
	app.Post("/register", handlers.Register)
	app.Post("/login", handlers.Login)

	app.Use(jwtware.New(jwtware.Config{
		SigningKey: []byte(os.Getenv("SECRET")),
		ContextKey: "userInfo",
	}))

	app.Post("/logout", handlers.Logout)
	app.Put("/connToodledo", handlers.ConnToodledo)
	app.Put("/connCloudStorage", handlers.ConnCloudStorage)
	app.Put("/setBackupFrequency", handlers.SetBackupFrequency)

	app.Listen(3000)

	// err := db.Client.Disconnect(context.TODO())

	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println("Disconnected from MongoDB")

}
