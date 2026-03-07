package ports

import (
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/session"
)

func ignoreCacheStaticFilesInDev(c fiber.Ctx) error {
	if os.Getenv("ENVIRONMENT") == "DEV" {
		c.Response().Header.Set("Cache-Control", "no-store")
	}
	return c.Next()
}

func isAuthorized(c fiber.Ctx) error {
	sess := session.FromContext(c)
	authorized, ok := sess.Get("authorized").(bool)
	if !ok || !authorized {
		if isHTMXRequest(c) {
			c.Set("HX-Redirect", "/oauth/login")
		} else {
			return c.Redirect().To("/oauth/login")
		}
	}
	return c.Next()
}
