package ports

import (
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/google/uuid"
	"github.com/siderustler/go-ecommerce/customer"
	"github.com/siderustler/go-ecommerce/ports/auth"
)

func ignoreCacheStaticFilesInDev(c *fiber.Ctx) error {
	if os.Getenv("ENVIRONMENT") == "DEV" {
		c.Response().Header.Set("Cache-Control", "no-store")
	}
	return c.Next()
}

type middleware struct {
	r             *customer.Services
	authenticator *auth.Authenticator
	sessionStore  *session.Store
}

func (m middleware) auth(c *fiber.Ctx) error {
	sess, err := m.sessionStore.Get(c)
	if err != nil {
		return c.SendString("retrieving store: " + err.Error())
	}
	var userID string
	if sess.Fresh() {
		userID = uuid.NewString()
		sess.Set("user_id", userID)

		if err = sess.Save(); err != nil {
			return c.SendString("saving session: " + err.Error())
		}
	} else {
		userID = sess.Get("user_id").(string)

	}
	ctx := auth.UserIDToContext(c.Context(), userID)
	ctx = auth.SessionToContext(ctx, sess)
	c.SetUserContext(ctx)

	return c.Next()
}

func (m middleware) authSessionVerifier(c *fiber.Ctx) error {
	sess := auth.SessionFromContext(c.UserContext())
	isExpired, err := auth.VerifySession(sess)
	if err != nil {
		fmt.Printf("verifying session: %v\n", err)
		c.Set("Hx-Redirect", "/oauth/login")
	}
	if isExpired {
		sess.Set("expiry", auth.TokenExpiryTime)
		sess.SetExpiry(auth.TokenExpiryTime)
		if err = sess.Save(); err != nil {
			fmt.Printf("saving session: %v", err.Error())
		}
	}
	return c.Next()
}
