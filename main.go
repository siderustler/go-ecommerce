package main

import (
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/siderustler/go-ecommerce/views"
)

func isHTMXRequest(c *fiber.Ctx) bool {
	_, ok := c.GetReqHeaders()["Hx-Request"]
	return ok
}

func main() {
	httpServer := fiber.New()
	httpServer.Get("/products/:product", func(c *fiber.Ctx) error {
		//go:inline
		var selectedImage = func(img string) int {
			if imgNum, err := strconv.Atoi(img); err == nil {
				return imgNum
			}
			return 0
		}

		imgQueryParam := strings.Trim(c.Query("img"), " ")

		productViewModel := views.NewProductDetailViewModel(
			"essa",
			"essa",
			[]string{"/public/products/essa/1.webp", "/public/products/essa/2.webp", "/public/products/essa/3.webp"},
			[]string{},
			[]string{},
			1.99,
			selectedImage(imgQueryParam),
		)

		var fragments []any
		if isHTMXRequest(c) && imgQueryParam != "" {
			fragments = append(fragments, views.ImageSelectorFragment)
		}

		return Render(c, views.ProductDetails(productViewModel), fragments...)
	})
	httpServer.Use("/static", func(c *fiber.Ctx) error {
		c.Response().Header.Set("Cache-Control", "no-store")
		return c.Next()
	})
	httpServer.Static("/public", "./views/public")
	httpServer.Listen(":8080")
}

func Render(c *fiber.Ctx, component templ.Component, fragments ...any) error {
	c.Set("Content-Type", "text/html")
	if len(fragments) > 0 {
		return templ.RenderFragments(c.Context(), c.Response().BodyWriter(), component, fragments...)
	}
	return component.Render(c.Context(), c.Response().BodyWriter())
}
