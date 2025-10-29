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
	page, _ := c.ParamsInt("page")
	var filterViewModel views.FilterViewModel
	filterViewModel.Validate()

	_ = c.QueryParser(&filterViewModel)
	products, err := h.services.GetProducts(c.Context(), page, filterViewModel.MapToDomainFilter())
	//FIXME -- display empty product list
	if err != nil {
		return c.Redirect("/products/1")
	}

	productsViewModel := views.NewProductsListViewModel(
		products,
		filterViewModel,
		page,
		10,
	)

	if !isHTMXRequest(c) {
		return render(c, views.Products(productsViewModel))
	}

	if filterViewModel.HasError() {
		c.Append("HX-Trigger", "validatePrice")
		return render(c, views.Products(productsViewModel), views.ProductListFragment)
	}

	values, _ := query.Values(filterViewModel)
	encodedUrl := "/products/1?" + values.Encode()
	isAlreadyOnRequestedUrl := strings.Contains(c.Get("HX-Current-Url"), encodedUrl)

	if !isAlreadyOnRequestedUrl {
		c.Append("HX-Push-Url", encodedUrl)
	}

	return render(c, views.Products(productsViewModel), views.ProductListFragment)
}

func (h handlers) getProductDetails(c *fiber.Ctx) error {
	var productDetailViewModel views.ProductDetailViewModel
	_ = c.QueryParser(&productDetailViewModel)
	productID := c.Params("productID", "")
	productDetails, err := h.services.GetProductDetails(c.Context(), productID)
	if err != nil {
		return c.Redirect("/products/1")
	}

	productDetailViewModel.Align(productDetails)
	if !isHTMXRequest(c) {
		return render(c, views.ProductDetails(productDetailViewModel))
	}

	var fragments []any
	toggleLocalInfo := c.Query("local") != ""
	toggleProductInfo := c.Query("info") != ""
	toggleTechParams := c.Query("tech-params") != ""
	toggleShippingInfo := c.Query("shipping") != ""
	swapImg := c.Query("img") != ""

	if toggleLocalInfo {
		fragments = append(fragments, views.ExpandLocalInfoFragment)
	}
	if toggleProductInfo {
		fragments = append(fragments, views.ExpandProductInfoFragment)
	}
	if toggleTechParams {
		fragments = append(fragments, views.ExpandTechnicalParametersFragment)
	}
	if toggleShippingInfo {
		fragments = append(fragments, views.ExpandShippingInfoFragment)
	}
	if swapImg {
		fragments = append(fragments, views.ImageSelectorFragment)
	}

	isOnProductDetails := strings.Contains(c.Get("Hx-Current-Url"), "/products/details/")

	if !isOnProductDetails {
		return render(c, views.ProductDetails(productDetailViewModel), views.ProductDetailFragment)
	}
	return render(c, views.ProductDetails(productDetailViewModel), fragments...)
}

func (h handlers) postProductsIncrement(c *fiber.Ctx) error {
	productID := c.Query("id")
	basketCount, _ := strconv.Atoi(c.FormValue("count"))
	page, _ := strconv.Atoi(c.Params("prod"))

	var filterViewModel views.FilterViewModel
	//FIXME render error
	_ = c.BodyParser(&filterViewModel)
	filterViewModel.Validate()

	products, err := h.services.GetProducts(c.Context(), page, filterViewModel.MapToDomainFilter())
	if err != nil {
		return c.Redirect("/products/1")
	}

	productsListViewModel := views.NewProductsListViewModel(products, filterViewModel, page, 10)
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
	var filterViewModel views.FilterViewModel
	//FIXME render error
	_ = c.BodyParser(&filterViewModel)
	filterViewModel.Validate()

	products, err := h.services.GetProducts(c.Context(), page, filterViewModel.MapToDomainFilter())
	if err != nil {
		return c.Redirect("/products/1")
	}
	productsListViewModel := views.NewProductsListViewModel(products, filterViewModel, page, 10)
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
	var filterViewModel views.FilterViewModel
	_ = c.BodyParser(&filterViewModel)
	filterViewModel.Validate()

	productID := c.Query("id")
	products, err := h.services.GetProducts(c.Context(), page, filterViewModel.MapToDomainFilter())
	if err != nil {
		return c.Redirect("/products/1")
	}
	//FIXME render error
	productsListViewModel := views.NewProductsListViewModel(products, filterViewModel, page, 10)
	productsListViewModel.ChangeProductBasketCount(productID, basketCount)

	fmt.Printf("Adding to basket count %d of item %s\n", basketCount, productID)

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

	if !isHTMXRequest(c) {
		return render(c, views.ProductsFilter(filterViewModel))
	}
	currentUrl, ok := c.GetReqHeaders()["Hx-Current-Url"]
	isOnFilterPageAlready := ok && len(currentUrl) >= 1 && strings.HasSuffix(currentUrl[0], "/filter/products")
	if !isOnFilterPageAlready {
		queries, _ := query.Values(filterViewModel)
		isAnyQueryIncluded := len(queries) > 0
		url := "/filter/products"
		if isAnyQueryIncluded {
			url += "?" + queries.Encode()
		}
		c.Append("HX-Push-Url", url)
	}
	return render(c, views.ProductsFilter(filterViewModel), views.ProductsFilterFragment)
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
		c.Append("Hx-Trigger", fmt.Sprintf(`{"preserveFilterInputFocus":{"triggerElement" : "%s"}}`, trigger[0]))
	}
	preserveFocus()
	return render(c, views.ProductsFilter(filterViewModel), views.PriceFilterFragment)
}

func (h handlers) getDashboard(c *fiber.Ctx) error {
	slide := c.QueryInt("slide", 0)
	promotionPage := c.QueryInt("promotions", 0)
	promos, err := h.services.GetPromotions(c.Context())
	//FIXME
	if err != nil {
		fmt.Printf("retrieving promotions: %v", err)
		return c.Redirect("/")
	}

	if !isHTMXRequest(c) {
		return render(c, views.Dashboard(views.NewDashboardViewModel(promos, slide, promotionPage)))
	}

	isPromotionRequest := c.Query("promotions") != ""
	if isPromotionRequest {
		return render(c, views.Dashboard(views.NewDashboardViewModel(promos, slide, promotionPage)), views.PromotedProductSelectorFragment)
	}

	return render(c, views.Dashboard(views.NewDashboardViewModel(promos, slide, promotionPage)), views.DashboardSelectorFragment)
}
