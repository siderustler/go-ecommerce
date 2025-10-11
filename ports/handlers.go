package ports

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/go-querystring/query"
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
	page, _ := strconv.Atoi(c.Params("prod"))
	products, err := h.services.GetProducts(c.Context(), page)

	//FIXME -- display empty product list
	if err != nil {
		return c.Redirect("/products/1")
	}

	var filterViewModel views.FilterViewModel
	//FIXME render error
	_ = c.QueryParser(&filterViewModel)
	filterViewModel.Validate()

	productsViewModel := views.NewProductsListViewModel(
		products,
		filterViewModel, 
		page, 
		10,
	)

	if isHTMXRequest(c) {
		if filterViewModel.HasError() {
			c.Append("HX-Trigger", "validatePrice")
			return nil
		}
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

	backUrl := c.Query("back")
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
		backUrl,
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
	backUrl := c.Query("back")
	productViewModel := views.NewProductDetailViewModel(
		productDetails,
		1,
		false,
		false,
		false,
		false,
		basketCount,
		backUrl,
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
	backUrl := c.Query("back")
	productViewModel := views.NewProductDetailViewModel(
		productDetails,
		1,
		false,
		false,
		false,
		false,
		basketCount,
		backUrl,
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
	backUrl := c.Query("back")
	productViewModel := views.NewProductDetailViewModel(
		productDetails,
		1,
		false,
		false,
		false,
		false,
		basketCount,
		backUrl,
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

	var filterViewModel views.FilterViewModel
	//FIXME render error
	_ = c.BodyParser(&filterViewModel)
	filterViewModel.Validate()

	productsListViewModel := views.NewProductsListViewModel(products,filterViewModel, page, 10)
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
	var filterViewModel views.FilterViewModel
	//FIXME render error
	_ = c.BodyParser(&filterViewModel)
	filterViewModel.Validate()
	productsListViewModel := views.NewProductsListViewModel(products, filterViewModel,page, 10)
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
	var filterViewModel views.FilterViewModel
	//FIXME render error
	_ = c.BodyParser(&filterViewModel)
	filterViewModel.Validate()
	productsListViewModel := views.NewProductsListViewModel(products,filterViewModel, page, 10)
	productsListViewModel.ChangeProductBasketCount(productID, basketCount)

	fmt.Printf("Adding to basket: %s and count: %v", productID, productsListViewModel)

	if isHTMXRequest(c) {
		fragments := append([]any{}, fmt.Sprintf("%+v-%s", views.BasketAddCounter, productID))

		return render(c, views.Products(productsListViewModel), fragments...)
	}

	return render(c, views.Products(productsListViewModel))
}


func (h handlers) getFilterProducts(c *fiber.Ctx) error {
	var filterViewModel views.FilterViewModel
	//FIXME render error
	_ = c.QueryParser(&filterViewModel)
	filterViewModel.Validate()

	queries,_ := query.Values(filterViewModel)
	isAnyQueryIncluded := len(queries) > 0
	var url string
	if isAnyQueryIncluded {
		url = "/filter/products?"+queries.Encode()
	} else {
		url = "/filter/products"
	}
	if isHTMXRequest(c) {
		currentUrl, ok :=  c.GetReqHeaders()["Hx-Current-Url"]
		if ok && len(currentUrl) >= 1 && !strings.HasSuffix(currentUrl[0], "/filter/products") {
			c.Append("HX-Push-Url", url)
		}
		return render(c, views.ProductsFilter(filterViewModel), views.ProductsFilterFragment)
	}
	return render(c, views.ProductsFilter(filterViewModel))
}

func (h handlers) filterProductsPriceValidate(c *fiber.Ctx) error {
	var filterViewModel views.FilterViewModel
	//FIXME render error
	_ = c.BodyParser(&filterViewModel)
	filterViewModel.Validate()

	var preserveFocus = func() {
		trigger, ok := c.GetReqHeaders()["Hx-Trigger"]
		if !ok || len(trigger) < 1 {
			return
		}
		c.Append("Hx-Trigger",fmt.Sprintf(`{"preserveFilterInputFocus":{"triggerElement" : "%s"}}`, trigger[0]))
	}
	preserveFocus()
	return render(c,views.ProductsFilter(filterViewModel), views.PriceFilterFragment)
}



func (h handlers) postFilterProducts(c *fiber.Ctx) error {
	var filterViewModel views.FilterViewModel
	_ = c.BodyParser(&filterViewModel)
	filterViewModel.Validate()
	page,_ := c.ParamsInt("prod")


	products,err := h.services.GetProducts(c.Context(), 1)
	//FIXME -- render empty list
	if err != nil {
		return c.Redirect("/products/1")
	}

	queries,_ := query.Values(filterViewModel)
	var url string
	isAnyQueryIncluded := len(queries) > 0
	if isAnyQueryIncluded {
		url = "/products/1?"+queries.Encode()
	} else {
		url = "/products/1"
	}

	productsListViewModel := views.NewProductsListViewModel(products,filterViewModel,page,10)
	if isHTMXRequest(c) {
		//FIXME -- rather than using events, 
		// use response target extension which allows to filter change target based on resp status
		if filterViewModel.HasError() {
			c.Append("HX-Trigger", "validatePrice")
		}

		c.Append("HX-Push-Url", url)
		return render(c,views.Products(productsListViewModel), views.ProductListFragment)
	}
	return c.Redirect(url)
}