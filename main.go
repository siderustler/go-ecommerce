package main

import (
	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/siderustler/go-ecommerce/views"
)

func main() {
	httpServer := fiber.New()
	httpServer.Get("/details", func(c *fiber.Ctx) error {
		return Render(c, views.ProductDetails())
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
