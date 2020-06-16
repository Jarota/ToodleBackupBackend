package handlers

import (
	"fmt"

	"github.com/gofiber/fiber"

	"github.com/jarota/toodle-backup/auth"
	// "github.com/jarota/toodle-backup/user"
	// "github.com/jarota/toodle-backup/database"
)

type credentials struct {
	username	string	`json:"username"`
	password	string	`json:"password"`
}

func HelloWorld(c *fiber.Ctx) {
	c.Send("Hello, World!\n")
}

func Register(c *fiber.Ctx) {
	c.Send("Register A New User\n")

	hash, err := auth.HashPassword("password")
	if err != nil {
		panic(err)
	}

	fmt.Println(hash)
	// TODO store new user in DB
	// user := user.NewUser(name, hash, ...)
}

func Login(c *fiber.Ctx) {
	c.Send("Login an existing user\n")
}

func Logout(c *fiber.Ctx) {
	c.Send("Logout a user\n")
}

func ConnCloudStorage(c *fiber.Ctx) {
	var service string
	service = c.Params("name")

	c.Send("Connect to: " + service)
}