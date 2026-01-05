package ports

import (
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
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
	r *customer.Services
}

func (m middleware) anonymoUser(c *fiber.Ctx) error {
	sessionToken := c.Cookies("session")
	token, err := auth.ParseJwtToken(sessionToken)
	isUserAuthorized := err == nil
	if !isUserAuthorized {
		fmt.Printf("checking jwt token from session cookie %+v\n", err)
		userID := uuid.NewString()
		token = auth.NewJwtToken(userID)
		rawToken, err := token.Sign()
		if err != nil {
			fmt.Printf("\nissuing new jwt token: %v", err)
			return nil
		}
		err = m.r.CreateCustomer(c.Context(), customer.NewCustomer(
			userID,
			customer.Credentials{},
			customer.Billing{ID: uuid.NewString()},
			customer.ShippingAddress{ID: uuid.NewString()},
		))
		if err != nil {
			fmt.Printf("\ncreating customer: %v", err)
			return nil
		}
		c.Cookie(&fiber.Cookie{Name: "session", Value: rawToken, HTTPOnly: true, SameSite: "Strict"})
	}
	ctx, err := token.ClaimsToContext(c.Context())
	if err != nil {
		fmt.Printf("setting claims to context: %v", err)
		return nil
	}
	c.SetUserContext(ctx)

	return c.Next()
}

/*
	Aktualnie user dostaje userID po wejsciu na strone, billing i shipping adres uzupelnia przed zaplata.
	Chcemy aby user mial mozliwosc pozostania anonimowym userem oraz mial mozliwosc tez zalogowania sie.
	W przypadku gdy user bedzie zalogowany od razu i nie bedzie mial nic w koszyku nic nie musimy robic.
	W przypadku gdy user nie bedize zalogowany i nie bbedzie mial nic wkoszyku i sie zaloguje nic nie bedzie trzeba robic.
	W przypadku gdy user nie ebbdzie zalogowany i doda cos do koszyka trzeba koszyk albo zmergowac albo nadpisac calkowicie.
	W przypadku gdy user nie bedzie zalogowany a ma juz stworzony checkout trzeba nadpisac jego koszyk (checkut zostanie nadpisany).
	W przypadku gdy user sie wyloguje, tworzona jest nowa sesja.

	Ilekroc koszyk bedzie mergowany albo nadpisywany, aktualny checkout uzytkownika
	 trzeba inwalidowac a rezerwacje w stock trzeba zwalniac
*/
