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
	h.srv.Get("/products/:prod", h.handlers.getProducts)
	h.srv.Get("/products/details/:product", h.handlers.getProductDetails)
	h.srv.Get("/filter/products", h.handlers.getFilterProducts)

	h.srv.Post("/products/details/:product/decrement", h.handlers.postProductDetailsDecrement)
	h.srv.Post("/products/details/:product/increment", h.handlers.postProductDetailsIncrement)
	h.srv.Post("/products/details/:product/basket-add", h.handlers.postProductDetailsBasketAdd)
	h.srv.Post("/products/:prod/increment", h.handlers.postProductsIncrement)
	h.srv.Post("/products/:prod/decrement", h.handlers.postProductsDecrement)
	h.srv.Post("/products/:prod/basket-add", h.handlers.postProductsBasketAdd)
	h.srv.Post("/filter/products", h.handlers.postFilterProducts)
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
