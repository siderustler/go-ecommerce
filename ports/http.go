package ports

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/postgres/v3"
	"github.com/siderustler/go-ecommerce/customer"
	"github.com/siderustler/go-ecommerce/ports/auth"
	"github.com/siderustler/go-ecommerce/product"
	store "github.com/siderustler/go-ecommerce/store"
)

type httpServer struct {
	srv      *fiber.App
	handlers *handlers
}

func NewHttpServer(
	customerServices *customer.Services,
	productServices *product.Services,
	storeServices *store.Services,
) *httpServer {
	authenticator, err := auth.New()
	if err != nil {
		panic(fmt.Errorf("creating new oauth authenticator: %v", err))
	}
	h := &httpServer{
		srv: fiber.New(),
		handlers: &handlers{
			customerServices: customerServices,
			productServices:  productServices,
			storeServices:    storeServices,
		},
	}
	sessionStore := session.New(session.Config{
		Expiration: time.Minute * 15,
		Storage: postgres.New(postgres.Config{
			ConnectionURI: os.Getenv("DATABASE_URI"),
		}),
		CookieHTTPOnly: true,
		//FIXME
		//CookieSecure: true,
	})

	m := &middleware{r: customerServices, sessionStore: sessionStore}
	h.srv.Use("/public", ignoreCacheStaticFilesInDev)
	h.srv.Get("/oauth/login", oauthLoginHandler(authenticator, sessionStore))
	h.srv.Get("/oauth/callback", h.handlers.oauthCallbackHandler(authenticator, sessionStore))
	h.srv.Get("/oauth/logout", oauthLogoutHandler(sessionStore))

	h.srv.Use("/", m.auth)
	authorizedSession := h.srv.Group("/user", m.authSessionVerifier)
	authorizedSession.Get("/", usersHandler(sessionStore))
	h.srv.Get("/products", h.handlers.getProductsRedirect)
	h.srv.Get("/products/:page", h.handlers.getProducts)
	h.srv.Get("/products/details/:productID", h.handlers.getProductDetails)
	h.srv.Get("/filter/products", h.handlers.getFilterProducts)
	h.srv.Get("/", h.handlers.getDashboard)
	h.srv.Get("/basket", h.handlers.getBasket)
	h.srv.Get("/basket/customer/billing", h.handlers.getBillingInfo)
	h.srv.Get("/basket/customer/shipping", h.handlers.getShippingInfo)
	h.srv.Get("/basket/checkout", h.handlers.getCheckoutStart)
	h.srv.Get("/basket/checkout/finalize", h.handlers.getCheckoutFinalized)

	h.srv.Post("/filter/products/validate/price", h.handlers.filterProductsPriceValidate)
	h.srv.Post("/basket/update", h.handlers.updateBasket)
	h.srv.Post("/basket/add", h.handlers.addItemToBasket)
	h.srv.Post("/basket/customer/billing", h.handlers.postBillingInfo)
	h.srv.Post("/basket/customer/shipping", h.handlers.postShippingInfo)
	h.srv.Post("/api/checkout", h.handlers.createCheckout)
	h.srv.Post("/api/stripe/wh", h.handlers.checkoutStripeWebhook)

	h.srv.Static("/public", "./ports/views/public")
	return h
}

func (h *httpServer) Run(ctx context.Context, addr string) error {
	return h.srv.Listen(addr)
}

func renderFragmentOrView(c *fiber.Ctx, component templ.Component, fragments ...any) error {
	c.Set("Content-Type", "text/html")
	if len(fragments) > 0 && isHTMXRequest(c) {
		return templ.RenderFragments(c.Context(), c.Response().BodyWriter(), component, fragments...)
	}
	return component.Render(c.Context(), c.Response().BodyWriter())
}

func renderFragmentOrRedirect(c *fiber.Ctx, component templ.Component, redirect string, fragments ...any) error {
	c.Set("Content-Type", "text/html")
	if len(fragments) > 0 && isHTMXRequest(c) {
		return templ.RenderFragments(c.Context(), c.Response().BodyWriter(), component, fragments...)
	}
	return c.Redirect(redirect)
}
func isHTMXRequest(c *fiber.Ctx) bool {
	_, ok := c.GetReqHeaders()["Hx-Request"]
	return ok
}
