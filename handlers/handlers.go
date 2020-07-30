package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"

	"github.com/gofiber/fiber"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/jarota/ToodleBackupBackend/auth"
	"github.com/jarota/ToodleBackupBackend/db"
	"github.com/jarota/ToodleBackupBackend/user"
)

type credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

const (
	dbName string = "ToodleBackup"
	users  string = "Users"
)

var client = db.ConnectToMongoDB()

// HelloWorld handler for basic testing
func HelloWorld(c *fiber.Ctx) {
	c.Send("Hello, World!\n")
}

// Register handler for registering new users
func Register(c *fiber.Ctx) {

	var creds credentials
	err := json.Unmarshal([]byte(c.Body()), &creds)

	if err != nil {
		c.SendStatus(fiber.StatusInternalServerError)
		return
	}

	userCollection, err := db.GetCollection(client, dbName, users)

	if err != nil {
		c.SendStatus(fiber.StatusInternalServerError)
		return
	}

	// Make sure there are no existing users with creds.username
	filter := bson.D{{Key: "username", Value: creds.Username}}

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

// Login handler for logging in an existing user and returning a JWT
func Login(c *fiber.Ctx) {

	var creds credentials
	err := json.Unmarshal([]byte(c.Body()), &creds)

	if err != nil {
		c.SendStatus(fiber.StatusInternalServerError)
		return
	}

	userCollection, err := db.GetCollection(client, dbName, users)

	if err != nil {
		c.SendStatus(fiber.StatusInternalServerError)
		return
	}

	// Check the user exists
	filter := bson.D{{Key: "username", Value: creds.Username}}

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

// Logout handler, currently a placeholder
func Logout(c *fiber.Ctx) {
	c.Send("Logout a user\n")
	/*
		JWT will expire after one day, if this is not enough
		of a logout solution, a 'Blacklist' of tokens must be
		kept track of
	*/

	// TODO
}

// GetUser handler for returning the logged in user's info
func GetUser(c *fiber.Ctx) {

	name := getAuthenticatedUsername(c)

	userCollection, err := db.GetCollection(client, dbName, users)

	if err != nil {
		c.SendStatus(fiber.StatusInternalServerError)
		return
	}

	filter := bson.D{{Key: "username", Value: name}}
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

// ConnToodledo handler for putting access token in the db
func ConnToodledo(c *fiber.Ctx) {

	var toodleInfo user.ToodleInfo
	json.Unmarshal([]byte(c.Body()), &toodleInfo)

	userCollection, err := db.GetCollection(client, dbName, users)

	if err != nil {
		c.SendStatus(fiber.StatusInternalServerError)
		return
	}

	name := getAuthenticatedUsername(c)
	filter := bson.D{{Key: "username", Value: name}}
	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "toodledo", Value: toodleInfo},
		}},
	}

	_, err = userCollection.UpdateOne(context.TODO(), filter, update)

	if err != nil {
		c.SendStatus(fiber.StatusInternalServerError)
		return
	}

	c.SendStatus(201)

}

// ConnCloudStorage handler for putting access token in the db
func ConnCloudStorage(c *fiber.Ctx) {

	var cloud user.Cloud
	json.Unmarshal([]byte(c.Body()), &cloud)

	userCollection, err := db.GetCollection(client, dbName, users)

	if err != nil {
		c.SendStatus(fiber.StatusInternalServerError)
		return
	}

	name := getAuthenticatedUsername(c)
	filter := bson.D{{Key: "username", Value: name}}
	update := bson.D{
		{Key: "$push", Value: bson.D{
			{Key: "clouds", Value: cloud},
		}},
	}

	_, err = userCollection.UpdateOne(context.TODO(), filter, update)

	if err != nil {
		c.SendStatus(fiber.StatusInternalServerError)
		return
	}

	c.SendStatus(201) // Cloud service successfully added

}

// SetBackupFrequency handler for setting/updating user's frequency in the db
func SetBackupFrequency(c *fiber.Ctx) {

	var freq string
	json.Unmarshal([]byte(c.Body()), &freq)

	userCollection, err := db.GetCollection(client, dbName, users)

	if err != nil {
		c.SendStatus(fiber.StatusInternalServerError)
		return
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
		return
	}

	c.SendStatus(201) // Backup frequency successfully set
}

type randomParams struct {
	APIKey     string `json:"apiKey"`
	N          int    `json:"n"`
	Length     int    `json:"length"`
	Characters string `json:"characters"`
}

type randomBody struct {
	JSONrpc string       `json:"jsonrpc"`
	Method  string       `json:"method"`
	Params  randomParams `json:"params"`
	ID      int          `json:"id"`
}

type randomData struct {
	Data           []string `json:"data"`
	CompletionTime string   `json:"completionTime"`
}

type generateStringsResult struct {
	Random        randomData `json:"random"`
	BitsUsed      int        `json:"bitsUsed"`
	BitsLeft      int        `json:"bitsLeft"`
	RequestsLeft  int        `json:"requestsLeft"`
	AdvisoryDelay int        `json:"advisoryDelay"`
}

type randomResponse struct {
	JSONrpc string                `json:"jsonrpc"`
	Result  generateStringsResult `json:"result"`
	ID      int                   `json:"id"`
}

// RandomString gets a string from random.org for state paramter in toodledo api redirect url
func RandomString(c *fiber.Ctx) {

	apiKey := os.Getenv("RANDOMAPI")

	params := &randomParams{
		APIKey:     apiKey,
		N:          1,
		Length:     10,
		Characters: "qwertyuiopasdfghjklzxcvbnm",
	}

	body := &randomBody{
		JSONrpc: "2.0",
		Method:  "generateStrings",
		Params:  *params,
		ID:      42,
	}

	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(body)

	contentType := "application/json"
	url := "https://api.random.org/json-rpc/2/invoke"
	resp, err := http.Post(url, contentType, buf)

	if err != nil {
		c.SendStatus(fiber.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		c.SendStatus(fiber.StatusInternalServerError)
		return
	}

	var randResp randomResponse
	json.Unmarshal(bytes, &randResp)

	c.Send(randResp.Result.Random.Data[0])

}

func getAuthenticatedUsername(c *fiber.Ctx) string {

	userInfo := c.Locals("userInfo").(*jwt.Token)
	claims := userInfo.Claims.(jwt.MapClaims)
	name := claims["name"].(string)
	return name

}
