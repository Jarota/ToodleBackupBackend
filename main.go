package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"

	jwtware "github.com/gofiber/jwt/v2"

	"github.com/jarota/ToodleBackupBackend/handlers"
	"github.com/jarota/ToodleBackupBackend/scheduler"
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
	app.Static("/dropboxredirect", "./frontend")

	app.Post("/api/register", handlers.Register)
	app.Post("/api/login", handlers.Login)

	app.Use(jwtware.New(jwtware.Config{
		SigningKey: []byte(os.Getenv("SECRET")),
		ContextKey: "userInfo",
	}))

	app.Get("/api/getUser", handlers.GetUser)
	app.Post("/api/logout", handlers.Logout)
	app.Put("/api/connToodledo", handlers.ConnToodledo)
	app.Put("/api/connDropbox", handlers.ConnDropbox)
	app.Put("/api/setBackupFrequency", handlers.SetBackupFrequency)
	app.Put("/api/setBackupTime", handlers.SetBackupTime)
	app.Get("/api/backupUser", handlers.BackupUser)

	app.Get("/api/randomString", handlers.RandomString)

	cert, err := tls.LoadX509KeyPair( // "./certs/server.crt", "./certs/server.key")
		"/etc/letsencrypt/live/toodlebackup.com/cert.pem",
		"/etc/letsencrypt/live/toodlebackup.com/privkey.pem",
	)
	if err != nil {
		log.Fatal(err)
	}

	config := &tls.Config{Certificates: []tls.Certificate{cert}}

	ln, err := tls.Listen("tcp", ":443", config)

	// Spin up scheduler
	go scheduler.PollForPendingBackups()

	// Start webserver
	log.Fatal(app.Listener(ln))

	// err := db.Client.Disconnect(context.TODO())

	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println("Disconnected from MongoDB")

}
