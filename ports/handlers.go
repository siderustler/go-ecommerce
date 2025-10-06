package ports

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/siderustler/go-ecommerce/ports/views"
	"github.com/siderustler/go-ecommerce/services"
)

type handlers struct {
	services *services.Services
}

func (h handlers) getProductsRedirect(c *fiber.Ctx) error {
	return c.Redirect("/products/1")
}

func (h handlers) getProducts(c *fiber.Ctx) error {
	pageParam := c.Params("page")

	page, err := strconv.Atoi(pageParam)
	if err != nil || page < 1 {
		page = 1
	}

	products, err := h.services.GetProducts(c.Context(), page)
	if err != nil {
		return c.Redirect("/products/1")
	}

	productsViewModel := views.NewProductsListViewModel(products, page, 10)

	return render(c, views.Products(productsViewModel))
}

func (h handlers) getProductDetails(c *fiber.Ctx) error {
	var fragments []any
	//go:inline
	var selectedImage = func() int {
		imgQueryParam := strings.Trim(c.Query("img"), " ")
		if imgNum, err := strconv.Atoi(imgQueryParam); err == nil {
			fragments = append(fragments, views.ImageSelectorFragment)
			return imgNum
		}
		return 0
	}

	var expandAdditionalInfo = func(param string, cb func()) bool {
		additionalInfoQueryParam := strings.Trim(c.Query(param), " ")
		if additionalInfoQueryParam == "true" || additionalInfoQueryParam == "false" {
			cb()
		}
		return additionalInfoQueryParam == "true"
	}
	productDetails, err := h.services.GetProductDetails(c.Context(), "essa")
	if err != nil {
		return c.Redirect("/products/1")
	}
	productViewModel := views.NewProductDetailViewModel(
		productDetails,
		selectedImage(),
		expandAdditionalInfo("info", func() {
			fragments = append(fragments, views.ExpandProductInfoFragment)
		}),
		expandAdditionalInfo("tech-params", func() {
			fragments = append(fragments, views.ExpandTechnicalParametersFragment)
		}),
		expandAdditionalInfo("shipping", func() {
			fragments = append(fragments, views.ExpandShippingInfoFragment)
		}),
		expandAdditionalInfo("local", func() {
			fragments = append(fragments, views.ExpandLocalInfoFragment)
		}),
		1,
	)

	if !isHTMXRequest(c) {
		return render(c, views.ProductDetails(productViewModel))
	}

	return render(c, views.ProductDetails(productViewModel), fragments...)
}

func (h handlers) postProductDetailsDecrement(c *fiber.Ctx) error {
	basketCount, _ := strconv.Atoi(c.FormValue("count"))

	productDetails, err := h.services.GetProductDetails(c.Context(), "essa")
	if err != nil {
		return c.Redirect("/products/1")
	}
	productViewModel := views.NewProductDetailViewModel(
		productDetails,
		1,
		false,
		false,
		false,
		false,
		basketCount,
	)

	productViewModel.DecrementBasketCount()
	if !isHTMXRequest(c) {
		return render(c, views.ProductDetails(productViewModel))
	}
	var fragments []any
	fragments = append(fragments, views.BasketAddCounter)
	return render(c, views.ProductDetails(productViewModel), fragments...)
}

func (h handlers) postProductDetailsIncrement(c *fiber.Ctx) error {
	basketCount, _ := strconv.Atoi(c.FormValue("count"))

	productDetails, err := h.services.GetProductDetails(c.Context(), "essa")
	if err != nil {
		return c.Redirect("/products/1")
	}

	productViewModel := views.NewProductDetailViewModel(
		productDetails,
		1,
		false,
		false,
		false,
		false,
		basketCount,
	)

	productViewModel.IncrementBasketCount()
	if !isHTMXRequest(c) {
		return render(c, views.ProductDetails(productViewModel))
	}
	var fragments []any
	fragments = append(fragments, views.BasketAddCounter)
	return render(c, views.ProductDetails(productViewModel), fragments...)
}

func (h handlers) postProductDetailsBasketAdd(c *fiber.Ctx) error {
	basketCount, _ := strconv.Atoi(c.FormValue("count"))

	productDetails, err := h.services.GetProductDetails(c.Context(), "essa")
	if err != nil {
		return c.Redirect("/products/1")
	}
	productViewModel := views.NewProductDetailViewModel(
		productDetails,
		1,
		false,
		false,
		false,
		false,
		basketCount,
	)

	if !isHTMXRequest(c) {
		return render(c, views.ProductDetails(productViewModel))
	}
	var fragments []any
	fragments = append(fragments, views.BasketAddCounter)
	return render(c, views.ProductDetails(productViewModel), fragments...)
}

func (h handlers) postProductsIncrement(c *fiber.Ctx) error {
	productID := c.Query("id")
	basketCount, _ := strconv.Atoi(c.FormValue("count"))
	page, _ := strconv.Atoi(c.Params("prod"))

	products, err := h.services.GetProducts(c.Context(), page)
	if err != nil {
		return c.Redirect("/products/1")
	}

	productsListViewModel := views.NewProductsListViewModel(products, page, 10)
	productsListViewModel.ChangeProductBasketCount(productID, basketCount+1)
	if isHTMXRequest(c) {
		fragments := append([]any{}, fmt.Sprintf("%+v-%s", views.BasketAddCounter, productID))

		return render(c, views.Products(productsListViewModel), fragments...)
	}

	return render(c, views.Products(productsListViewModel))
}

func (h handlers) postProductsDecrement(c *fiber.Ctx) error {
	basketCount, _ := strconv.Atoi(c.FormValue("count"))
	page, _ := strconv.Atoi(c.Params("prod"))
	productID := c.Query("id")

	products, err := h.services.GetProducts(c.Context(), page)
	if err != nil {
		return c.Redirect("/products/1")
	}

	productsListViewModel := views.NewProductsListViewModel(products, page, 10)
	productsListViewModel.ChangeProductBasketCount(productID, basketCount-1)
	if isHTMXRequest(c) {
		fragments := append([]any{}, fmt.Sprintf("%+v-%s", views.BasketAddCounter, productID))

		return render(c, views.Products(productsListViewModel), fragments...)
	}

	return render(c, views.Products(productsListViewModel))
}

func (h handlers) postProductsBasketAdd(c *fiber.Ctx) error {
	basketCount, _ := strconv.Atoi(c.FormValue("count"))
	page, _ := strconv.Atoi(c.Params("prod"))
	productID := c.Query("id")

	products, err := h.services.GetProducts(c.Context(), page)
	if err != nil {
		return c.Redirect("/products/1")
	}

	productsListViewModel := views.NewProductsListViewModel(products, page, 10)
	productsListViewModel.ChangeProductBasketCount(productID, basketCount)

	fmt.Printf("Adding to basket: %s and count: %v", productID, productsListViewModel)

	if isHTMXRequest(c) {
		fragments := append([]any{}, fmt.Sprintf("%+v-%s", views.BasketAddCounter, productID))

		return render(c, views.Products(productsListViewModel), fragments...)
	}

	return render(c, views.Products(productsListViewModel))
}
