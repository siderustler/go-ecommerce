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
	page, _ := strconv.Atoi(c.Params("page"))
	priceFrom, _ := strconv.ParseFloat(c.Query("price-from"),32)
	priceTo, _ := strconv.ParseFloat(c.Query("price-to"),32)
	machinesFilter := c.Query("machines") == "true"
	gardeningFilter := c.Query("gardening") == "true"
	partsFilter := c.Query("parts") == "true"
	electroFilter := c.Query("electro") == "true"
	electroMachinesFilter := c.Query("electromachines") == "true"
	products, err := h.services.GetProducts(c.Context(), page)

	//FIXME -- display empty product list
	if err != nil {
		return c.Redirect("/products/1")
	}

	filterViewModel := views.NewFilterViewModel(float32(priceFrom),float32(priceTo),machinesFilter,gardeningFilter,partsFilter,electroFilter,electroMachinesFilter)
	productsViewModel := views.NewProductsListViewModel(
		products,
		filterViewModel, 
		page, 
		10,
	)

	if isHTMXRequest(c) {
		if filterViewModel.HasError() {
			return render(c,views.Products(productsViewModel), views.PriceFromInputErrFragment)
		}
		return render(c,views.Products(productsViewModel),views.ProductListFragment)
	}


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
	priceFrom, _ := strconv.ParseFloat(c.Query("price-from"),32)
	priceTo, _ := strconv.ParseFloat(c.Query("price-to"),32)

	products, err := h.services.GetProducts(c.Context(), page)
	if err != nil {
		return c.Redirect("/products/1")
	}
	var machinesFilter,electroFilter, gardeningFilter,partsFilter, electroMachinesFilter bool
	productsListViewModel := views.NewProductsListViewModel(
		products,		views.NewFilterViewModel(float32(priceFrom),float32(priceTo),machinesFilter,gardeningFilter,partsFilter,electroFilter,electroMachinesFilter), page, 10)
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
	priceFrom, _ := strconv.ParseFloat(c.Query("price-from"),32)
	priceTo, _ := strconv.ParseFloat(c.Query("price-to"),32)

	products, err := h.services.GetProducts(c.Context(), page)
	if err != nil {
		return c.Redirect("/products/1")
	}
	var machinesFilter,electroFilter, gardeningFilter,partsFilter, electroMachinesFilter bool
	productsListViewModel := views.NewProductsListViewModel(products, 		views.NewFilterViewModel(float32(priceFrom),float32(priceTo),machinesFilter,gardeningFilter,partsFilter,electroFilter,electroMachinesFilter),page, 10)
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
	priceFrom, _ := strconv.ParseFloat(c.Query("price-from"),32)
	priceTo, _ := strconv.ParseFloat(c.Query("price-to"),32)
	var machinesFilter,electroFilter, gardeningFilter,partsFilter, electroMachinesFilter bool
	products, err := h.services.GetProducts(c.Context(), page)
	if err != nil {
		return c.Redirect("/products/1")
	}

	productsListViewModel := views.NewProductsListViewModel(products,views.NewFilterViewModel(float32(priceFrom),float32(priceTo),machinesFilter,gardeningFilter,partsFilter,electroFilter,electroMachinesFilter), page, 10)
	productsListViewModel.ChangeProductBasketCount(productID, basketCount)

	fmt.Printf("Adding to basket: %s and count: %v", productID, productsListViewModel)

	if isHTMXRequest(c) {
		fragments := append([]any{}, fmt.Sprintf("%+v-%s", views.BasketAddCounter, productID))

		return render(c, views.Products(productsListViewModel), fragments...)
	}

	return render(c, views.Products(productsListViewModel))
}


func (h handlers) getFilterProducts(c *fiber.Ctx) error {
	priceFrom,_ := strconv.ParseFloat(c.FormValue("price-from"), 32)
	priceTo,_ := strconv.ParseFloat(c.FormValue("price-to"),32)
	machines := c.FormValue("machines") == "true"
	gardening := c.FormValue("gardening") == "true"
	parts := c.FormValue("parts") == "true"
	electro := c.FormValue("electro") == "true"
	electroMachines := c.FormValue("electromachines") == "true"
	filterViewModel := views.NewFilterViewModel(float32(priceFrom),float32(priceTo),machines,gardening,parts,electro,electroMachines)
	if isHTMXRequest(c) && filterViewModel.HasError() {
		return render(c,views.ProductsFilter(filterViewModel), views.PriceFromInputErrFragment)
	}
	return render(c, views.ProductsFilter(filterViewModel))
}

func (h handlers) postFilterProducts(c *fiber.Ctx) error {
	page,_ := c.ParamsInt("prod")
	priceFrom,_ := strconv.ParseFloat(c.FormValue("price-from"), 32)
	priceTo,_ := strconv.ParseFloat(c.FormValue("price-to"),32)
	machines := c.FormValue("machines") == "true"
	gardening := c.FormValue("gardening") == "true"
	parts := c.FormValue("parts") == "true"
	electro := c.FormValue("electro") == "true"
	electroMachines := c.FormValue("electromachines") == "true"

	filterViewModel := views.NewFilterViewModel(float32(priceFrom),float32(priceTo),machines,gardening,parts,electro,electroMachines)
	products,err := h.services.GetProducts(c.Context(), 1)
	//FIXME -- render empty list
	if err != nil {
		return c.Redirect("/products/1")
	}
	productsListViewModel := views.NewProductsListViewModel(products,filterViewModel,page,10)
	if isHTMXRequest(c) {
		if filterViewModel.HasError() {
			return render(c,views.Products(productsListViewModel), views.PriceFromInputErrFragment)
		}
		return render(c,views.Products(productsListViewModel), views.ProductListFragment, views.PriceFromInputErrFragment)
	}
	return render(c,views.Products(productsListViewModel))
}