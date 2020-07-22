package main

import (
	"fmt"
	"os"

	"github.com/gofiber/cors"
	"github.com/gofiber/fiber"
	jwtware "github.com/gofiber/jwt"

	// "github.com/jarota/toodle-backup/db"
	"github.com/jarota/ToodleBackupBackend/handlers"
)

func main() {
	fmt.Println("Starting Toodle Backup Backend...")

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		ExposeHeaders: []string{"Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With"},
	}))

	app.Get("/", handlers.HelloWorld)
	app.Post("/register", handlers.Register)
	app.Post("/login", handlers.Login)

	app.Use(jwtware.New(jwtware.Config{
		SigningKey: []byte(os.Getenv("SECRET")),
		ContextKey: "userInfo",
	}))

	app.Get("/getUser", handlers.GetUser)
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
