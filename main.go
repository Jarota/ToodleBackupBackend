package main

import (
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"

	jwtware "github.com/gofiber/jwt/v2"

	"github.com/jarota/ToodleBackupBackend/handlers"
)

func main() {
	fmt.Println("Starting Toodle Backup Backend...")

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		ExposeHeaders: "Access-Control-Allow-Headers, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With",
	}))

	app.Use(logger.New())

	app.Static("/", "./frontend")
	app.Static("/toodleredirect", "./frontend")

	app.Post("/api/register", handlers.Register)
	app.Post("/api/login", handlers.Login)

	app.Use(jwtware.New(jwtware.Config{
		SigningKey: []byte(os.Getenv("SECRET")),
		ContextKey: "userInfo",
	}))

	app.Get("/api/getUser", handlers.GetUser)
	app.Post("/api/logout", handlers.Logout)
	app.Put("/api/connToodledo", handlers.ConnToodledo)
	app.Put("/api/connCloudStorage", handlers.ConnCloudStorage)
	app.Put("/api/setBackupFrequency", handlers.SetBackupFrequency)

	app.Get("/api/randomString", handlers.RandomString)

	err := app.Listen(":80")

	if err != nil {
		fmt.Println(err)
	}

	// err := db.Client.Disconnect(context.TODO())

	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println("Disconnected from MongoDB")
	fmt.Println("Fin")

}
