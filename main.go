package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"

	jwtware "github.com/gofiber/jwt/v3"

	"github.com/jarota/ToodleBackupBackend/db"
	"github.com/jarota/ToodleBackupBackend/handlers"
	"github.com/jarota/ToodleBackupBackend/scheduler"
)

func main() {
	fmt.Println("Starting Toodle Backup Backend...")
	ctx := context.Background()
	dbc := db.ConnectToMongoDB(ctx)
	defer dbc.Disconnect(ctx)

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		ExposeHeaders: "Access-Control-Allow-Headers, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With",
	}))

	app.Use(logger.New())

	app.Static("/", "./frontend")
	app.Static("/toodleredirect", "./frontend")
	app.Static("/dropboxredirect", "./frontend")

	app.Post("/api/register", handlers.Register(dbc))
	app.Post("/api/login", handlers.Login(dbc))

	app.Use(jwtware.New(jwtware.Config{
		SigningKey: []byte(os.Getenv("SECRET")),
		ContextKey: "userInfo",
	}))

	app.Get("/api/getUser", handlers.GetUser(dbc))
	app.Post("/api/logout", handlers.Logout(dbc))
	app.Put("/api/connToodledo", handlers.ConnToodledo(dbc))
	app.Put("/api/connDropbox", handlers.ConnDropbox(dbc))
	app.Put("/api/setBackupFrequency", handlers.SetBackupFrequency(dbc))
	app.Put("/api/setBackupTime", handlers.SetBackupTime(dbc))
	app.Get("/api/backupUser", handlers.BackupUser(dbc))

	app.Get("/api/randomString", handlers.RandomString(dbc))

	cert, err := tls.LoadX509KeyPair("./certs/server.crt", "./certs/server.key")
	// 	"/etc/letsencrypt/live/toodlebackup.com/cert.pem",
	// 	"/etc/letsencrypt/live/toodlebackup.com/privkey.pem",
	// )
	if err != nil {
		log.Fatal(err)
	}

	config := &tls.Config{Certificates: []tls.Certificate{cert}}

	ln, err := tls.Listen("tcp", ":443", config)
	if err != nil {
		log.Fatal(err)
	}

	// Spin up scheduler
	go scheduler.PollForPendingBackups(ctx, dbc)

	// Start webserver
	log.Fatal(app.Listener(ln))
}
