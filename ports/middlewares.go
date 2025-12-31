package ports

import (
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/siderustler/go-ecommerce/customer"
)

func ignoreCacheStaticFilesInDev(c *fiber.Ctx) error {
	if os.Getenv("ENVIRONMENT") == "DEV" {
		c.Response().Header.Set("Cache-Control", "no-store")
	}
	return c.Next()
}

type middleware struct {
	r *customer.Services
}

func (m middleware) anonymoUser(c *fiber.Ctx) error {
	sessionEstablished := c.Cookies("session") != ""
	if !sessionEstablished {
		userID := uuid.NewString()
		err := m.r.CreateCustomer(c.Context(), customer.NewCustomer(
			userID,
			customer.Credentials{},
			customer.Billing{ID: uuid.NewString()},
			customer.ShippingAddress{ID: uuid.NewString()},
		))
		if err != nil {
			fmt.Printf("creating customer: %+v", err)
			return nil
		}
		c.Cookie(&fiber.Cookie{Name: "session", Value: userID})
	}
	return c.Next()
}
