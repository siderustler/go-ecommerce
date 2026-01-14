package ports

import (
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
		userID := uuid.NewString()
		sess.Set("user_id", userID)

		if err = sess.Save(); err != nil {
			return c.SendString("saving session: " + err.Error())
		}
	} else {
		userID = sess.Get("user_id").(string)
	}
	ctx := auth.UserIDToContext(c.Context(), userID)

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
