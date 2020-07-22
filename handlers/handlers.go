package handlers

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"

	"github.com/gofiber/fiber"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/jarota/toodle-backup/auth"
	"github.com/jarota/toodle-backup/user"
	"github.com/jarota/toodle-backup/db"
)

type credentials struct {
	Username	string 	`json:"username"`
	Password	string 	`json:"password"`
}

const (
	DB string = "ToodleBackup"
	USERS string = "Users"
)

var client = db.ConnectToMongoDB()

func HelloWorld(c *fiber.Ctx) {
	c.Send("Hello, World!\n")
}


func Register(c *fiber.Ctx) {
	
	var creds credentials
	err := json.Unmarshal([]byte(c.Body()), &creds)

	if err != nil {
		c.SendStatus(fiber.StatusInternalServerError)
		return
	}

	userCollection, err := db.GetCollection(client, DB, USERS)

	if err != nil {
		c.SendStatus(fiber.StatusInternalServerError)
		return
	}
	
	// Make sure there are no existing users with creds.username
	filter := bson.D{{"username", creds.Username}}

	var possibleUser user.User
	err = userCollection.FindOne(context.TODO(), filter).Decode(&possibleUser)
	
	if err != mongo.ErrNoDocuments {
		c.Status(409).Send("Username taken")
		return
	} 
	
	// Otherwise store new user in DB
	hash, err := auth.HashPassword(creds.Password)
	if err != nil {
		c.SendStatus(fiber.StatusInternalServerError)
		return
	}
		
	u := user.New(creds.Username, hash)

	_, err = userCollection.InsertOne(context.TODO(), u)
	if err != nil {
		c.SendStatus(fiber.StatusInternalServerError)
		return
	}

	// Create token
	token := jwt.New(jwt.SigningMethodHS256)

	// Set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["name"] = u.Username
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	// Generate encoded token and send it as response
	t, err := token.SignedString([]byte(os.Getenv("SECRET")))

	if err != nil {
		c.SendStatus(fiber.StatusInternalServerError)
		return
	}

	c.JSON(fiber.Map{"token": t})
	
}


func Login(c *fiber.Ctx) {
	
	var creds credentials
	err := json.Unmarshal([]byte(c.Body()), &creds)

	if err != nil {
		c.SendStatus(fiber.StatusInternalServerError)
		return
	}

	userCollection, err := db.GetCollection(client, DB, USERS)

	if err != nil {
		c.SendStatus(fiber.StatusInternalServerError)
		return
	}
	
	// Check the user exists
	filter := bson.D{{"username", creds.Username}}

	var u user.User
	err = userCollection.FindOne(context.TODO(), filter).Decode(&u)

	if err != nil {
		c.SendStatus(fiber.StatusUnauthorized)
		return
	}

	// Check the passwords match
	match, err := auth.ComparePassAndHash(creds.Password, u.Password)

	if err != nil || match == false {
		c.SendStatus(fiber.StatusUnauthorized)
		return
	}

	// Create token
	token := jwt.New(jwt.SigningMethodHS256)

	// Set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["name"] = u.Username
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	// Generate encoded token and send it as response
	t, err := token.SignedString([]byte(os.Getenv("SECRET")))

	if err != nil {
		c.SendStatus(fiber.StatusInternalServerError)
		return
	}

	c.JSON(fiber.Map{"token": t})

}


func Logout(c *fiber.Ctx) {
	c.Send("Logout a user\n")
	/*
		JWT will expire after one day, if this is not enough
		of a logout solution, a 'Blacklist' of tokens must be
		kept track of
	*/
	
	// TODO
}


func GetUser(c *fiber.Ctx) {

	name := getAuthenticatedUsername(c)

	userCollection, err := db.GetCollection(client, DB, USERS)

	if err != nil {
		c.SendStatus(fiber.StatusInternalServerError)
		return
	}

	filter := bson.D{{"username", name}}
	var u user.User
	err = userCollection.FindOne(context.TODO(), filter).Decode(&u)

	if err != nil {
		c.SendStatus(fiber.StatusInternalServerError)
		log.Printf("Error finding user in database")
		return
	}

	b, err := json.Marshal(u)

	if err != nil {
		c.SendStatus(fiber.StatusInternalServerError)
		return
	}

	c.Send(b)

}


func ConnToodledo(c *fiber.Ctx) {

	var toodleInfo user.ToodleInfo
	json.Unmarshal([]byte(c.Body()), &toodleInfo)

	userCollection, err := db.GetCollection(client, DB, USERS)

	if err != nil {
		c.SendStatus(fiber.StatusInternalServerError)
		return
	}
	
	name := getAuthenticatedUsername(c)
	filter := bson.D{{"username", name}}
	update := bson.D{
		{"$set", bson.D{
			{"toodledo", toodleInfo},
		}},
	}

	_, err = userCollection.UpdateOne(context.TODO(), filter, update)

	if err != nil {
		c.SendStatus(fiber.StatusInternalServerError)
		return
	}

	c.SendStatus(201)
	
}


func ConnCloudStorage(c *fiber.Ctx) {
	
	var cloud user.Cloud
	json.Unmarshal([]byte(c.Body()), &cloud)

	userCollection, err := db.GetCollection(client, DB, USERS)
	
	if err != nil {
		c.SendStatus(fiber.StatusInternalServerError)
		return
	}

	name := getAuthenticatedUsername(c)
	filter := bson.D{{"username", name}}
	update := bson.D{
		{"$push", bson.D{
			{"clouds", cloud},
		}},
	}

	_, err = userCollection.UpdateOne(context.TODO(), filter, update)

	if err != nil {
		c.SendStatus(fiber.StatusInternalServerError)
		return
	}

	c.SendStatus(201) // Cloud service successfully added

}


func SetBackupFrequency(c *fiber.Ctx) {

	var freq string
	json.Unmarshal([]byte(c.Body()), &freq)

	userCollection, err := db.GetCollection(client, DB, USERS)
	
	if err != nil {
		c.SendStatus(fiber.StatusInternalServerError)
		return
	}

	name := getAuthenticatedUsername(c)
	filter := bson.D{{"username", name}}
	update := bson.D{
		{"$set", bson.D{
			{"frequency", freq},
		}},
	}

	_, err = userCollection.UpdateOne(context.TODO(), filter, update)

	if err != nil {
		c.SendStatus(fiber.StatusInternalServerError)
		return
	}

	c.SendStatus(201) // Backup frequency successfully set
}


func getAuthenticatedUsername(c *fiber.Ctx) string{
	
	userInfo := c.Locals("userInfo").(*jwt.Token)
	claims := userInfo.Claims.(jwt.MapClaims)
	name := claims["name"].(string)
	return name

}