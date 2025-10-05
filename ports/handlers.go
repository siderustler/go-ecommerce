package ports

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/siderustler/go-ecommerce/ports/views"
)

func getProductsRedirect(c *fiber.Ctx) error {
	return c.Redirect("/products/1")
}

func getProducts(c *fiber.Ctx) error {
	pageParam := c.Params("page")

	page, err := strconv.Atoi(pageParam)
	if err != nil || page < 1 {
		page = 1
	}
	productsViewModel := views.NewProductsViewModel([]views.Product{
		views.NewProduct("1", "essa", "/public/products/essa/1.webp", 1.99, 1),
		views.NewProduct("2", "dwa", "/public/products/essa/1.webp", 2.99, 1),
		views.NewProduct("3", "trzy", "/public/products/essa/1.webp", 3.99, 1),
	}, page, 10)

	return render(c, views.Products(productsViewModel))
}

func getProductDetails(c *fiber.Ctx) error {
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

	productViewModel := views.NewProductDetailViewModel(
		"essa",
		"essa",
		[]string{"/public/products/essa/1.webp", "/public/products/essa/2.webp", "/public/products/essa/3.webp"},
		[]string{
			`Nowa podkaszarka Daewoo. 
				Dzięki niskiej wadze i niedużym rozmiarom 
				podkaszarka DATR 800E świetnie sprawdzi się na małej działce czy w ogródku przydomowym.
				`,
			`Nowa podkaszarka Daewoo. 
				Dzięki niskiej wadze i niedużym rozmiarom 
				podkaszarka DATR 800E świetnie sprawdzi się na małej działce czy w ogródku przydomowym.
				`,
			`Nowa podkaszarka Daewoo. 
				Dzięki niskiej wadze i niedużym rozmiarom 
				podkaszarka DATR 800E świetnie sprawdzi się na małej działce czy w ogródku przydomowym.
				`,
		},
		[]string{},
		1.99,
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

func postProductDetailsDecrement(c *fiber.Ctx) error {
	countQueryParam := c.FormValue("count")
	parsedCount, _ := strconv.Atoi(countQueryParam)
	basketCount := 1
	if parsedCount > 1 {
		basketCount = parsedCount - 1
	}
	productViewModel := views.NewProductDetailViewModel(
		"essa",
		"essa",
		[]string{"/public/products/essa/1.webp", "/public/products/essa/2.webp", "/public/products/essa/3.webp"},
		[]string{
			`Nowa podkaszarka Daewoo. 
				Dzięki niskiej wadze i niedużym rozmiarom 
				podkaszarka DATR 800E świetnie sprawdzi się na małej działce czy w ogródku przydomowym.
				`,
			`Nowa podkaszarka Daewoo. 
				Dzięki niskiej wadze i niedużym rozmiarom 
				podkaszarka DATR 800E świetnie sprawdzi się na małej działce czy w ogródku przydomowym.
				`,
			`Nowa podkaszarka Daewoo. 
				Dzięki niskiej wadze i niedużym rozmiarom 
				podkaszarka DATR 800E świetnie sprawdzi się na małej działce czy w ogródku przydomowym.
				`,
		},
		[]string{},
		1.99,
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

func postProductDetailsIncrement(c *fiber.Ctx) error {
	countQueryParam := c.FormValue("count")
	parsedCount, _ := strconv.Atoi(countQueryParam)
	basketCount := 1
	if parsedCount > 0 {
		basketCount = parsedCount + 1
	}
	productViewModel := views.NewProductDetailViewModel(
		"essa",
		"essa",
		[]string{"/public/products/essa/1.webp", "/public/products/essa/2.webp", "/public/products/essa/3.webp"},
		[]string{
			`Nowa podkaszarka Daewoo. 
				Dzięki niskiej wadze i niedużym rozmiarom 
				podkaszarka DATR 800E świetnie sprawdzi się na małej działce czy w ogródku przydomowym.
				`,
			`Nowa podkaszarka Daewoo. 
				Dzięki niskiej wadze i niedużym rozmiarom 
				podkaszarka DATR 800E świetnie sprawdzi się na małej działce czy w ogródku przydomowym.
				`,
			`Nowa podkaszarka Daewoo. 
				Dzięki niskiej wadze i niedużym rozmiarom 
				podkaszarka DATR 800E świetnie sprawdzi się na małej działce czy w ogródku przydomowym.
				`,
		},
		[]string{},
		1.99,
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

func postProductDetailsBasketAdd(c *fiber.Ctx) error {
	countQueryParam := c.FormValue("count")
	basketCount := 1
	parsedCount, _ := strconv.Atoi(countQueryParam)
	if parsedCount > 1 {
		basketCount = parsedCount
	}
	productViewModel := views.NewProductDetailViewModel(
		"essa",
		"essa",
		[]string{"/public/products/essa/1.webp", "/public/products/essa/2.webp", "/public/products/essa/3.webp"},
		[]string{
			`Nowa podkaszarka Daewoo. 
				Dzięki niskiej wadze i niedużym rozmiarom 
				podkaszarka DATR 800E świetnie sprawdzi się na małej działce czy w ogródku przydomowym.
				`,
			`Nowa podkaszarka Daewoo. 
				Dzięki niskiej wadze i niedużym rozmiarom 
				podkaszarka DATR 800E świetnie sprawdzi się na małej działce czy w ogródku przydomowym.
				`,
			`Nowa podkaszarka Daewoo. 
				Dzięki niskiej wadze i niedużym rozmiarom 
				podkaszarka DATR 800E świetnie sprawdzi się na małej działce czy w ogródku przydomowym.
				`,
		},
		[]string{},
		1.99,
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

func postProductsIncrement(c *fiber.Ctx) error {
	countQueryParam := c.FormValue("count")
	basketCount := 1
	parsedCount, _ := strconv.Atoi(countQueryParam)
	if parsedCount > 1 {
		basketCount = parsedCount
	}
	pageParam := c.Params("prod")

	page, err := strconv.Atoi(pageParam)
	if err != nil || page < 1 {
		page = 1
	}

	productIDQuery := c.Query("id")
	products := []views.Product{
		views.NewProduct("1", "essa", "/public/products/essa/1.webp", 1.99, 1),
		views.NewProduct("2", "dwa", "/public/products/essa/1.webp", 2.99, 1),
		views.NewProduct("3", "trzy", "/public/products/essa/1.webp", 3.99, 1),
	}
	for i := 0; i < len(products); i++ {
		if products[i].ID == productIDQuery {
			products[i].ChangeBasketCount(basketCount)
			products[i].Increment()
		}
	}
	productsViewModel := views.NewProductsViewModel(products, page, 10)
	if isHTMXRequest(c) {
		fragments := append([]any{}, fmt.Sprintf("%+v-%s", views.BasketAddCounter, productIDQuery))

		return render(c, views.Products(productsViewModel), fragments...)
	}

	return render(c, views.Products(productsViewModel))
}

func postProductsDecrement(c *fiber.Ctx) error {
	countQueryParam := c.FormValue("count")
	basketCount := 1
	parsedCount, _ := strconv.Atoi(countQueryParam)
	if parsedCount > 1 {
		basketCount = parsedCount
	}
	pageParam := c.Params("prod")

	page, err := strconv.Atoi(pageParam)
	if err != nil || page < 1 {
		page = 1
	}

	productIDQuery := c.Query("id")
	products := []views.Product{
		views.NewProduct("1", "essa", "/public/products/essa/1.webp", 1.99, 1),
		views.NewProduct("2", "dwa", "/public/products/essa/1.webp", 2.99, 1),
		views.NewProduct("3", "trzy", "/public/products/essa/1.webp", 3.99, 1),
	}
	for i := 0; i < len(products); i++ {
		if products[i].ID == productIDQuery {
			products[i].ChangeBasketCount(basketCount)
			products[i].Decrement()
		}
	}
	productsViewModel := views.NewProductsViewModel(products, page, 10)
	if isHTMXRequest(c) {
		fragments := append([]any{}, fmt.Sprintf("%+v-%s", views.BasketAddCounter, productIDQuery))

		return render(c, views.Products(productsViewModel), fragments...)
	}

	return render(c, views.Products(productsViewModel))
}
func postProductsBasketAdd(c *fiber.Ctx) error {
	countQueryParam := c.FormValue("count")
	basketCount := 1
	parsedCount, _ := strconv.Atoi(countQueryParam)
	if parsedCount > 1 {
		basketCount = parsedCount
	}
	pageParam := c.Params("prod")

	page, err := strconv.Atoi(pageParam)
	if err != nil || page < 1 {
		page = 1
	}

	productIDQuery := c.Query("id")
	products := []views.Product{
		views.NewProduct("1", "essa", "/public/products/essa/1.webp", 1.99, 1),
		views.NewProduct("2", "dwa", "/public/products/essa/1.webp", 2.99, 1),
		views.NewProduct("3", "trzy", "/public/products/essa/1.webp", 3.99, 1),
	}

	for i := 0; i < len(products); i++ {
		if products[i].ID == productIDQuery {
			products[i].ChangeBasketCount(basketCount)
		}
	}
	fmt.Printf("Adding to basket: %s", productIDQuery)

	productsViewModel := views.NewProductsViewModel(products, page, 10)
	if isHTMXRequest(c) {
		fragments := append([]any{}, fmt.Sprintf("%+v-%s", views.BasketAddCounter, productIDQuery))

		return render(c, views.Products(productsViewModel), fragments...)
	}

	return render(c, views.Products(productsViewModel))
}
