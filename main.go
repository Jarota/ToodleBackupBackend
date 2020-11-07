package main

import (
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	jwtware "github.com/gofiber/jwt/v2"

	"github.com/jarota/ToodleBackupBackend/handlers"
)

func main() {
	fmt.Println("Starting Toodle Backup Backend...")

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		ExposeHeaders: "Access-Control-Allow-Headers, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With",
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

	app.Get("/randomString", handlers.RandomString)

	app.Listen(":3000")

	// err := db.Client.Disconnect(context.TODO())

	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println("Disconnected from MongoDB")
	fmt.Println("Fin")

}
