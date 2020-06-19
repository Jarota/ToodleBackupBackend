package handlers

import (
	"context"
	"encoding/json"
	"log"

	"github.com/gofiber/fiber"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/jarota/toodle-backup/auth"
	"github.com/jarota/toodle-backup/user"
	"github.com/jarota/toodle-backup/db"
)

type credentials struct {
	username	string	`json:"username"`
	password	string	`json:"password"`
}

const (
	DB string = "ToodleBackup"
	USERS string = "Users"
)

func HelloWorld(c *fiber.Ctx) {
	c.Send("Hello, World!\n")
}

// Should be passed a credentials object in the request body
func Register(c *fiber.Ctx) {
	
	var creds credentials
	json.Unmarshal([]byte(c.Body()), &creds)

	userCollection, err := db.GetCollection(DB, USERS)

	if err != nil {
		log.Fatal(err)
	}
	
	// Make sure there are no existing users with creds.username
	filter := bson.D{{"Username", creds.username}}

	var existingUser user.User
	err = userCollection.FindOne(context.TODO(), filter).Decode(&existingUser)
	if err == nil || err != mongo.ErrNoDocuments {
		c.Status(409).Send("Username taken")
	} else { // Otherwise store new user in DB
		hash, err := auth.HashPassword(creds.password)
		if err != nil {
			panic(err)
		}
		
		u := user.New(creds.username, hash)

		_, err = userCollection.InsertOne(context.TODO(), u)
		if err != nil {
			log.Fatal(err)
		}

		c.Status(201).Send("User successfully registered")
	}

}

func Login(c *fiber.Ctx) {
	c.Send("Login an existing user\n")

}

func Logout(c *fiber.Ctx) {
	c.Send("Logout a user\n")
}

func ConnToodledo(c *fiber.Ctx) {
	
}

func ConnCloudStorage(c *fiber.Ctx) {
	
	var cloud user.Cloud
	json.Unmarshal([]byte(c.Body()), &cloud)


}