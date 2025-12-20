package ports

import (
	"os"

	"github.com/gofiber/fiber/v2"
)

func ignoreCacheStaticFilesInDev(c *fiber.Ctx) error {
	if os.Getenv("ENVIRONMENT") == "DEV" {
		c.Response().Header.Set("Cache-Control", "no-store")
	}
	return c.Next()
}

func anonymoUser(c *fiber.Ctx) error {
	sessionEstablished := c.Cookies("session") != ""
	if !sessionEstablished {
		c.Cookie(&fiber.Cookie{Name: "session", Value: "XD"})
	}
	return c.Next()
}
