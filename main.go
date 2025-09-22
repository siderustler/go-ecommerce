package main

import (
	"os"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/siderustler/go-ecommerce/views"
)

func main() {
	f := fiber.New()
	f.Get("/details", func(c *fiber.Ctx) error {
		if c.Query("fragment") == "xd" {
			RenderFragment(c, views.ProductDetails(), os.Stdout)
			os.Stdout.Write([]byte("rendering"))
			return RenderFragment(c, views.ProductDetails(), views.DetailFragment)
		}
		return Render(c, views.ProductDetails())
	})
	f.Listen(":8080")
}

func Render(c *fiber.Ctx, component templ.Component) error {
	c.Set("Content-Type", "text/html")
	return component.Render(c.Context(), c.Response().BodyWriter())
}

func RenderFragment(c *fiber.Ctx, component templ.Component, ids any) error {
	c.Set("Content-Type", "text/html")
	return templ.RenderFragments(c.Context(), c.Response().BodyWriter(), component, ids)
}
