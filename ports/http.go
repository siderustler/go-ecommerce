package ports

import (
	"context"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
)

type httpServer struct {
	srv *fiber.App
}

func NewHttpServer() *httpServer {
	h := &httpServer{srv: fiber.New()}
	h.srv.Use("/public", ignoreCacheStaticFilesInDev)

	h.srv.Get("/products", getProductsRedirect)
	h.srv.Get("/products/:prod", getProducts)
	h.srv.Get("/products/details/:product", getProductDetails)

	h.srv.Post("/products/details/:product/decrement", postProductDetailsDecrement)
	h.srv.Post("/products/details/:product/increment", postProductDetailsIncrement)
	h.srv.Post("/products/details/:product/basket-add", postProductDetailsBasketAdd)
	h.srv.Post("/products/:prod/increment", postProductsIncrement)
	h.srv.Post("/products/:prod/decrement", postProductsDecrement)
	h.srv.Post("/products/:prod/basket-add", postProductsBasketAdd)

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
