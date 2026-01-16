package ports

import (
	"fmt"
	"os"
	"time"

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
		userID := uuid.NewString()
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
	isExpired, err := auth.VerifySession(c.IP(), sess)
	if err != nil {
		return c.SendString("verifying session: not authorized: " + err.Error())
	}

	if isExpired {
		refreshTokenResp, err := m.authenticator.RefreshToken(c.Context(), sess.Get("refresh_token").(string))
		if err != nil {

			fmt.Printf("refreshing session failed: %v", err)
			if err = sess.Destroy(); err != nil {
				fmt.Printf("destroying session: %v", err)
			}
			return c.Next()
		}
		sess.Set("access_token", refreshTokenResp.AccessToken)
		sess.Set("id_token", refreshTokenResp.IdToken)
		expiryTime := time.Now().Unix() + refreshTokenResp.ExpiresIn
		sess.Set("expiry", expiryTime)
		sess.SetExpiry(time.Duration(expiryTime))
		sess.Set("ip", c.IP())

		if err = sess.Save(); err != nil {
			fmt.Printf("saving session: %v", err.Error())
		}
	}
	return c.Next()
}
