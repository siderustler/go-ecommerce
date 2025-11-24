package ports

import (
	"context"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/siderustler/go-ecommerce/services"
)

type httpServer struct {
	srv      *fiber.App
	handlers *handlers
}

func NewHttpServer(services *services.Services) *httpServer {
	h := &httpServer{
		srv:      fiber.New(),
		handlers: &handlers{services: services},
	}
	h.srv.Use("/public", ignoreCacheStaticFilesInDev)

	h.srv.Get("/products", h.handlers.getProductsRedirect)
	h.srv.Get("/products/:page", h.handlers.getProducts)
	h.srv.Get("/products/details/:productID", h.handlers.getProductDetails)
	h.srv.Get("/filter/products", h.handlers.getFilterProducts)
	h.srv.Get("/", h.handlers.getDashboard)
	h.srv.Get("/basket", h.handlers.getBasket)
	h.srv.Get("/basket/customer/billing", h.handlers.getBillingInfo)
	h.srv.Get("/basket/customer/shipping", h.handlers.getShippingInfo)

	h.srv.Post("/filter/products/validate/price", h.handlers.filterProductsPriceValidate)
	h.srv.Post("/basket", h.handlers.updateBasket)
	h.srv.Post("/basket/customer/billing", h.handlers.postBillingInfo)
	h.srv.Post("/basket/customer/shipping", h.handlers.postShippingInfo)

	h.srv.Static("/public", "./ports/views/public")
	return h
}

func (h *httpServer) Run(ctx context.Context, addr string) error {
	return h.srv.Listen(addr)
}

func render(c *fiber.Ctx, component templ.Component, fragments ...any) error {
	c.Set("Content-Type", "text/html")
	if len(fragments) > 0 {
		return templ.RenderFragments(c.Context(), c.Response().BodyWriter(), component, fragments...)
	}
	return component.Render(c.Context(), c.Response().BodyWriter())
}

func isHTMXRequest(c *fiber.Ctx) bool {
	_, ok := c.GetReqHeaders()["Hx-Request"]
	return ok
}
