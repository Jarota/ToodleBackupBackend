package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"

	"github.com/gofiber/fiber/v2"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/jarota/ToodleBackupBackend/auth"
	"github.com/jarota/ToodleBackupBackend/db"
	"github.com/jarota/ToodleBackupBackend/dropbox"
	"github.com/jarota/ToodleBackupBackend/random"
	"github.com/jarota/ToodleBackupBackend/toodledo"
	"github.com/jarota/ToodleBackupBackend/user"
)

type credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type code struct {
	Value string `json:"value"`
}

const (
	dbName string = "ToodleBackup"
	users  string = "Users"
)

var client = db.ConnectToMongoDB()

// HelloWorld handler for basic testing
func HelloWorld(c *fiber.Ctx) error {
	c.Send([]byte("Hello, World!\n"))
	return nil
}

// Register handler for registering new users
func Register(c *fiber.Ctx) error {

	var creds credentials
	err := json.Unmarshal([]byte(c.Body()), &creds)

	if err != nil {
		c.SendStatus(fiber.StatusInternalServerError)
		return err
	}

	userCollection, err := db.GetCollection(client, dbName, users)

	if err != nil {
		c.SendStatus(fiber.StatusInternalServerError)
		return err
	}

	// Make sure there are no existing users with creds.username
	filter := bson.D{{Key: "username", Value: creds.Username}}

	var possibleUser user.User
	err = userCollection.FindOne(context.TODO(), filter).Decode(&possibleUser)

	if err != mongo.ErrNoDocuments {
		c.Status(409).Send([]byte("Username taken"))
		return err
	}

	// Otherwise store new user in DB
	hash, err := auth.HashPassword(creds.Password)
	if err != nil {
		c.SendStatus(fiber.StatusInternalServerError)
		return err
	}

	u := user.New(creds.Username, hash)

	_, err = userCollection.InsertOne(context.TODO(), u)
	if err != nil {
		c.SendStatus(fiber.StatusInternalServerError)
		return err
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
		return err
	}

	c.JSON(fiber.Map{"token": t})
	return nil
}

// Login handler for logging in an existing user and returning a JWT
func Login(c *fiber.Ctx) error {

	var creds credentials
	err := json.Unmarshal([]byte(c.Body()), &creds)

	if err != nil {
		c.SendStatus(fiber.StatusInternalServerError)
		return err
	}

	userCollection, err := db.GetCollection(client, dbName, users)

	if err != nil {
		c.SendStatus(fiber.StatusInternalServerError)
		return err
	}

	// Check the user exists
	filter := bson.D{{Key: "username", Value: creds.Username}}

	var u user.User
	err = userCollection.FindOne(context.TODO(), filter).Decode(&u)

	if err != nil {
		c.SendStatus(fiber.StatusUnauthorized)
		return err
	}

	// Check the passwords match
	match, err := auth.ComparePassAndHash(creds.Password, u.Password)

	if err != nil || match == false {
		c.SendStatus(fiber.StatusUnauthorized)
		return err
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
		return err
	}

	c.JSON(fiber.Map{"token": t})
	return nil
}

// Logout handler, currently a placeholder
func Logout(c *fiber.Ctx) error {
	c.Send([]byte("Logout a user\n"))
	/*
		JWT will expire after one day, if this is not enough
		of a logout solution, a 'Blacklist' of tokens must be
		kept track of
	*/

	// TODO
	return nil
}

// GetUser handler for returning the logged in user's info
func GetUser(c *fiber.Ctx) error {

	name := getAuthenticatedUsername(c)

	userCollection, err := db.GetCollection(client, dbName, users)

	if err != nil {
		c.SendStatus(fiber.StatusInternalServerError)
		return err
	}

	filter := bson.D{{Key: "username", Value: name}}
	var u user.User
	err = userCollection.FindOne(context.TODO(), filter).Decode(&u)

	if err != nil {
		c.SendStatus(fiber.StatusInternalServerError)
		log.Printf("Error finding user in database")
		return err
	}

	b, err := json.Marshal(u)

	if err != nil {
		c.SendStatus(fiber.StatusInternalServerError)
		return err
	}

	c.Send(b)
	return nil
}

// ConnToodledo handler for putting access token in the db
func ConnToodledo(c *fiber.Ctx) error {

	var code code
	json.Unmarshal([]byte(c.Body()), &code)
	fmt.Println(code.Value)

	toodleInfo, err := toodledo.GetToodledoTokens(code.Value)

	if err != nil {
		c.SendStatus(fiber.StatusInternalServerError)
		return err
	}

	userCollection, err := db.GetCollection(client, dbName, users)

	if err != nil {
		c.SendStatus(fiber.StatusInternalServerError)
		return err
	}

	name := getAuthenticatedUsername(c)
	filter := bson.D{{Key: "username", Value: name}}
	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "toodledo", Value: *toodleInfo},
		}},
	}

	_, err = userCollection.UpdateOne(context.TODO(), filter, update)

	if err != nil {
		c.SendStatus(fiber.StatusInternalServerError)
		return err
	}

	c.SendStatus(201)
	return nil
}

// ConnDropbox handler for putting access token in the db
func ConnDropbox(c *fiber.Ctx) error {

	var code code
	json.Unmarshal([]byte(c.Body()), &code)

	dropboxInfo, err := dropbox.GetDropboxTokens(code.Value)

	if err != nil {
		c.SendStatus(401)
		return err
	}

	userCollection, err := db.GetCollection(client, dbName, users)

	if err != nil {
		c.SendStatus(402)
		return err
	}

	name := getAuthenticatedUsername(c)
	filter := bson.D{{Key: "username", Value: name}}
	update := bson.D{
		{Key: "$push", Value: bson.D{
			{Key: "clouds", Value: dropboxInfo},
		}},
	}

	_, err = userCollection.UpdateOne(context.TODO(), filter, update)

	if err != nil {
		c.SendStatus(403)
		return err
	}

	c.SendStatus(201) // Cloud service successfully added
	return nil
}

// SetBackupFrequency handler for setting/updating user's frequency in the db
func SetBackupFrequency(c *fiber.Ctx) error {

	var freq string
	json.Unmarshal([]byte(c.Body()), &freq)

	userCollection, err := db.GetCollection(client, dbName, users)

	if err != nil {
		c.SendStatus(fiber.StatusInternalServerError)
		return err
	}

	name := getAuthenticatedUsername(c)
	filter := bson.D{{Key: "username", Value: name}}
	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "frequency", Value: freq},
		}},
	}

	_, err = userCollection.UpdateOne(context.TODO(), filter, update)

	if err != nil {
		c.SendStatus(fiber.StatusInternalServerError)
		return err
	}

	c.SendStatus(201) // Backup frequency successfully set
	return nil
}

// RandomString gets a string from random.org for state paramter in toodledo api redirect url
func RandomString(c *fiber.Ctx) error {

	rand, err := random.GetRandomString()

	if err != nil {
		c.SendStatus(fiber.StatusInternalServerError)
		return err
	}

	c.JSON(fiber.Map{"state": rand})
	return nil
}

func getAuthenticatedUsername(c *fiber.Ctx) string {

	userInfo := c.Locals("userInfo").(*jwt.Token)
	claims := userInfo.Claims.(jwt.MapClaims)
	name := claims["name"].(string)
	return name

}
