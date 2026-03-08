package ports

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/gofiber/fiber/v3/middleware/static"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/session"
	"github.com/gofiber/storage/postgres/v3"
	"github.com/siderustler/go-ecommerce/customer"
	"github.com/siderustler/go-ecommerce/ports/auth"
	"github.com/siderustler/go-ecommerce/product"
	store "github.com/siderustler/go-ecommerce/store"
	"github.com/stripe/stripe-go/v84"
)

type httpServer struct {
	srv      *fiber.App
	handlers *handlers
}

func NewHttpServer(
	customerServices *customer.Services,
	productServices *product.Services,
	storeServices *store.Services,
	logger *slog.Logger,
) (*httpServer, error) {
	if logger == nil {
		logger = slog.Default()
	}
	stripe.Key = os.Getenv("STRIPE_SERVER_KEY")

	authenticator, err := auth.New()
	if err != nil {
		return nil, err
	}
	h := &httpServer{
		srv: fiber.New(),
		handlers: &handlers{
			customerServices: customerServices,
			productServices:  productServices,
			storeServices:    storeServices,
			logger:           logger.With("component", "ports.handlers"),
		},
	}
	sessionStore := session.New(session.Config{
		IdleTimeout: time.Hour * 24 * 3,
		Storage: postgres.New(postgres.Config{
			ConnectionURI: os.Getenv("DATABASE_URI"),
		}),
		CookieSameSite: "Lax",
		CookieHTTPOnly: true,
		CookieSecure:   os.Getenv("ENVIRONMENT") != "DEV",
	})
	h.srv.Use(requestIDMiddleware)
	h.srv.Use(clientIDMiddleware)
	h.srv.Use(requestLoggingMiddleware(logger.With("component", "ports.http")))
	h.srv.Use("/public", ignoreCacheStaticFilesInDev)

	auth := h.srv.Group("/", sessionStore)
	auth.Get("/oauth/logout", oauthLogoutHandler)
	auth.Get("/oauth/login", oauthLoginHandler(authenticator))
	auth.Get("/oauth/callback", h.handlers.oauthCallbackHandler(authenticator))
	account := h.srv.Group("/account", sessionStore, isAuthorized)
	account.Get("/", h.handlers.accountHandler)
	account.Get("/customer/billing", h.handlers.getBillingInfo)
	account.Get("/customer/shipping", h.handlers.getShippingInfo)
	auth.Get("/products", h.handlers.getProductsRedirect)
	auth.Get("/products/:page", h.handlers.getProducts)
	auth.Get("/products/details/:productID", h.handlers.getProductDetails)
	auth.Get("/filter/products", h.handlers.getFilterProducts)
	auth.Get("/", h.handlers.getDashboard)
	auth.Get("/basket", h.handlers.getBasket)
	auth.Get("/basket/customer/billing", h.handlers.getBasketBillingInfo)
	auth.Get("/basket/customer/shipping", h.handlers.getBasketShippingInfo)
	auth.Get("/basket/checkout", h.handlers.getCheckoutStart)
	auth.Get("/basket/checkout/finalize", h.handlers.getCheckoutFinalized)

	account.Post("/customer/shipping", h.handlers.postShippingInfo)
	account.Post("/customer/billing", h.handlers.postBillingInfo)
	auth.Post("/basket/update", h.handlers.updateBasket)
	auth.Post("/basket/add", h.handlers.addItemToBasket)
	auth.Post("/basket/customer/billing", h.handlers.postBasketBillingInfo)
	auth.Post("/basket/customer/shipping", h.handlers.postBasketShippingInfo)
	auth.Post("/api/checkout", h.handlers.createCheckout)
	h.srv.Post("/api/stripe/wh", h.handlers.checkoutStripeWebhook)

	h.srv.Get("/public*", static.New("./ports/views/public"))
	return h, nil
}

func (h *httpServer) Run(ctx context.Context, addr string) error {
	return h.srv.Listen(addr)
}

func renderFragmentOrView(c fiber.Ctx, component templ.Component, fragments ...any) error {
	c.Set("Content-Type", "text/html")
	if len(fragments) > 0 && isHTMXRequest(c) {
		return templ.RenderFragments(c.RequestCtx(), c.Response().BodyWriter(), component, fragments...)
	}
	return component.Render(c.RequestCtx(), c.Response().BodyWriter())
}

func renderFragmentOrRedirect(c fiber.Ctx, component templ.Component, redirect string, fragments ...any) error {
	c.Set("Content-Type", "text/html")
	if len(fragments) > 0 && isHTMXRequest(c) {
		return templ.RenderFragments(c.RequestCtx(), c.Response().BodyWriter(), component, fragments...)
	}
	return c.Redirect().To(redirect)
}
func isHTMXRequest(c fiber.Ctx) bool {
	_, ok := c.GetReqHeaders()["Hx-Request"]
	return ok
}

// fiber:context-methods migrated
