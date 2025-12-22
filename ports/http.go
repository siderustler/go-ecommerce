package ports

import (
	"context"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/siderustler/go-ecommerce/customer"
	"github.com/siderustler/go-ecommerce/product"
	"github.com/siderustler/go-ecommerce/store"
)

type httpServer struct {
	srv      *fiber.App
	handlers *handlers
}

func NewHttpServer(
	customerServices *customer.Services,
	productServices *product.Services,
	basketServices *store.Services,
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
	anonymoUserGrp := h.srv.Group("/", anonymoUser)
	anonymoUserGrp.Get("/products", h.handlers.getProductsRedirect)
	anonymoUserGrp.Get("/products/:page", h.handlers.getProducts)
	anonymoUserGrp.Get("/products/details/:productID", h.handlers.getProductDetails)
	anonymoUserGrp.Get("/filter/products", h.handlers.getFilterProducts)
	anonymoUserGrp.Get("/", h.handlers.getDashboard)
	anonymoUserGrp.Get("/basket", h.handlers.getBasket)
	anonymoUserGrp.Get("/basket/customer/billing", h.handlers.getBillingInfo)
	anonymoUserGrp.Get("/basket/customer/shipping", h.handlers.getShippingInfo)
	anonymoUserGrp.Get("/basket/checkout", h.handlers.getCheckout)

	anonymoUserGrp.Post("/filter/products/validate/price", h.handlers.filterProductsPriceValidate)
	anonymoUserGrp.Post("/basket/update", h.handlers.updateBasket)
	anonymoUserGrp.Post("/basket/add", h.handlers.addItemToBasket)
	anonymoUserGrp.Post("/basket/customer/billing", h.handlers.postBillingInfo)
	anonymoUserGrp.Post("/basket/customer/shipping", h.handlers.postShippingInfo)
	h.srv.Post("/api/checkout", h.handlers.createCheckout)

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
