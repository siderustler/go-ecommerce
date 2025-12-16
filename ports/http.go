package ports

import (
	"context"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/siderustler/go-ecommerce/basket"
	"github.com/siderustler/go-ecommerce/customer"
	"github.com/siderustler/go-ecommerce/product"
)

type httpServer struct {
	srv      *fiber.App
	handlers *handlers
}

func NewHttpServer(
	customerServices *customer.Services,
	productServices *product.Services,
	basketServices *basket.Services,
) *httpServer {
	h := &httpServer{
		srv: fiber.New(),
		handlers: &handlers{
			customerServices: customerServices,
			productServices:  productServices,
			basketServices:   basketServices,
		},
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
	h.srv.Get("/basket/checkout", h.handlers.getCheckout)

	h.srv.Post("/filter/products/validate/price", h.handlers.filterProductsPriceValidate)
	h.srv.Post("/basket/update", h.handlers.updateBasket)
	h.srv.Post("/basket/add", h.handlers.addItemToBasket)
	h.srv.Post("/basket/customer/billing", h.handlers.postBillingInfo)
	h.srv.Post("/basket/customer/shipping", h.handlers.postShippingInfo)
	h.srv.Post("/api/checkout", h.handlers.createCheckout)

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
