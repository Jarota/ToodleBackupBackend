package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/jarota/ToodleBackupBackend/auth"
	"github.com/jarota/ToodleBackupBackend/db"
	"github.com/jarota/ToodleBackupBackend/dropbox"
	"github.com/jarota/ToodleBackupBackend/random"
	"github.com/jarota/ToodleBackupBackend/scheduler"
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

// HelloWorld handler for basic testing
func HelloWorld(_ mongo.Client) handler {
	return func(c *fiber.Ctx) error {
		c.Send([]byte("Hello, World!\n"))
		return nil
	}
}

// Register handler for registering new users
func Register(dbc *mongo.Client) handler {
	ctx := context.Background()
	return func(c *fiber.Ctx) error {
		var creds credentials
		err := json.Unmarshal([]byte(c.Body()), &creds)

		if err != nil {
			c.SendStatus(fiber.StatusInternalServerError)
			return err
		}

		userCollection, err := db.GetCollection(dbc, dbName, users)

		if err != nil {
			c.SendStatus(fiber.StatusInternalServerError)
			return err
		}

		// Make sure there are no existing users with creds.username
		filter := bson.D{{Key: "username", Value: creds.Username}}

		var possibleUser user.User
		err = userCollection.FindOne(ctx, filter).Decode(&possibleUser)

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

		_, err = userCollection.InsertOne(ctx, u)
		if err != nil {
			c.SendStatus(fiber.StatusInternalServerError)
			return err
		}

		t, err := auth.CreateJWT(u.Username)
		if err != nil {
			c.SendStatus(fiber.StatusInternalServerError)
			return err
		}

		c.JSON(fiber.Map{"token": t})
		return nil
	}
}

// Login handler for logging in an existing user and returning a JWT
func Login(dbc *mongo.Client) handler {
	ctx := context.Background()
	return func(c *fiber.Ctx) error {
		var creds credentials
		err := json.Unmarshal([]byte(c.Body()), &creds)
		if err != nil {
			c.SendStatus(fiber.StatusInternalServerError)
			return err
		}

		userCollection, err := db.GetCollection(dbc, dbName, users)
		if err != nil {
			c.SendStatus(fiber.StatusInternalServerError)
			return err
		}

		// Check the user exists
		filter := bson.D{{Key: "username", Value: creds.Username}}

		var u user.User
		err = userCollection.FindOne(ctx, filter).Decode(&u)
		if err != nil {
			c.SendStatus(fiber.StatusUnauthorized)
			return err
		}

		// Check the passwords match
		match, err := auth.ComparePassAndHash(creds.Password, u.Password)
		if !match || err != nil {
			c.SendStatus(fiber.StatusUnauthorized)
			return err
		}

		t, err := auth.CreateJWT(u.Username)
		if err != nil {
			c.SendStatus(fiber.StatusInternalServerError)
			return err
		}

		c.JSON(fiber.Map{"token": t})
		return nil
	}
}

// Logout handler, currently a placeholder
func Logout(dbc *mongo.Client) handler {
	return func(c *fiber.Ctx) error {
		c.Send([]byte("Logout a user\n"))
		/*
			JWT will expire after one day, if this is not enough
			of a logout solution, a 'Blacklist' of tokens must be
			kept track of
		*/

		// TODO
		return nil
	}
}

type handler = func(c *fiber.Ctx) error

// GetUser handler for returning the logged in user's info
func GetUser(dbc *mongo.Client) handler {
	ctx := context.Background()
	return func(c *fiber.Ctx) error {
		name := getAuthenticatedUsername(c)

		userCollection, err := db.GetCollection(dbc, dbName, users)
		if err != nil {
			c.SendStatus(fiber.StatusInternalServerError)
			return err
		}

		filter := bson.D{{Key: "username", Value: name}}
		var u user.User
		err = userCollection.FindOne(ctx, filter).Decode(&u)
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
}

// ConnToodledo handler for putting access token in the db
func ConnToodledo(dbc *mongo.Client) handler {
	ctx := context.Background()
	return func(c *fiber.Ctx) error {
		var code code
		json.Unmarshal([]byte(c.Body()), &code)
		fmt.Println(code.Value)

		toodleInfo, err := toodledo.GetToodledoTokens(code.Value, "authorization_code")
		if err != nil {
			c.SendStatus(fiber.StatusInternalServerError)
			return err
		}

		userCollection, err := db.GetCollection(dbc, dbName, users)
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
		_, err = userCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.SendStatus(fiber.StatusInternalServerError)
			return err
		}

		c.SendStatus(201)
		return nil
	}
}

// ConnDropbox handler for putting access token in the db
func ConnDropbox(dbc *mongo.Client) handler {
	ctx := context.Background()
	return func(c *fiber.Ctx) error {
		var code code
		json.Unmarshal([]byte(c.Body()), &code)

		_, dropboxInfo, err := dropbox.GetDropboxTokens(code.Value, "authorization_code")
		if err != nil {
			c.SendStatus(401)
			return err
		}

		userCollection, err := db.GetCollection(dbc, dbName, users)
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
		_, err = userCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.SendStatus(403)
			return err
		}

		c.SendStatus(201) // Cloud service successfully added
		return nil
	}
}

// SetBackupFrequency handler for setting/updating user's frequency in the db
func SetBackupFrequency(dbc *mongo.Client) handler {
	ctx := context.Background()
	return func(c *fiber.Ctx) error {
		var freq string
		json.Unmarshal([]byte(c.Body()), &freq)

		userCollection, err := db.GetCollection(dbc, dbName, users)
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

		_, err = userCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.SendStatus(fiber.StatusInternalServerError)
			return err
		}

		c.SendStatus(201) // Backup frequency successfully set
		return nil
	}
}

// SetBackupTime sets the backuptime for the authenticated user
func SetBackupTime(dbc *mongo.Client) handler {
	ctx := context.Background()
	return func(c *fiber.Ctx) error {
		var t user.BackupTime
		json.Unmarshal([]byte(c.Body()), &t)

		userCollection, err := db.GetCollection(dbc, dbName, users)
		if err != nil {
			c.SendStatus(fiber.StatusBadRequest)
			return err
		}

		name := getAuthenticatedUsername(c)
		filter := bson.D{{Key: "username", Value: name}}
		update := bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "time", Value: t},
			}},
		}
		_, err = userCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.SendStatus(fiber.StatusInternalServerError)
			return err
		}

		c.SendStatus(201) // Backup Time successfully set
		return nil
	}
}

// BackupUser is an explicit call to the backup function
func BackupUser(dbc *mongo.Client) handler {
	ctx := context.Background()
	return func(c *fiber.Ctx) error {
		userCollection, err := db.GetCollection(dbc, dbName, users)
		if err != nil {
			c.SendStatus(fiber.StatusBadRequest)
			return err
		}

		name := getAuthenticatedUsername(c)
		filter := bson.D{{Key: "username", Value: name}}

		var u user.User
		err = userCollection.FindOne(ctx, filter).Decode(&u)
		if err != nil {
			c.SendStatus(fiber.StatusInternalServerError)
			return err
		}

		scheduler.BackupUserData(ctx, dbc, &u)

		c.SendStatus(200) // User backup complete
		return nil
	}
}

// RandomString gets a string from random.org for state paramter in toodledo api redirect url
func RandomString(_ *mongo.Client) handler {
	return func(c *fiber.Ctx) error {
		rand, err := random.GetRandomString()
		if err != nil {
			c.SendStatus(fiber.StatusInternalServerError)
			return err
		}

		c.JSON(fiber.Map{"state": rand})
		return nil
	}
}
