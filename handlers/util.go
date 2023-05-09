package handlers

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
)

func getAuthenticatedUsername(c *fiber.Ctx) string {
	userInfo := c.Locals("userInfo").(*jwt.Token)
	claims := userInfo.Claims.(jwt.MapClaims)
	name := claims["name"].(string)
	return name
}
