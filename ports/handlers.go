package ports

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/go-querystring/query"
	"github.com/siderustler/go-ecommerce/ports/views"
	"github.com/siderustler/go-ecommerce/ports/views/components"
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
	_ = c.QueryParser(&filterViewModel)
	filterViewModel.Validate()

	var productsListViewModel views.ProductsListViewModel
	_ = c.QueryParser(&productsListViewModel)

	var navBarViewModel components.NavBarViewModel
	_ = c.QueryParser(&navBarViewModel)

	products, err := h.services.GetProducts(c.Context(), page, filterViewModel.MapToDomainFilter())
	//FIXME -- display empty product list
	if err != nil {
		return c.Redirect("/products/1")
	}

	basketProducts, err := h.services.GetBasket(c.Context(), "")
	//FIXME
	if err != nil {
		return c.Redirect("/products/1")
	}

	navBarViewModel.Align(basketProducts)

	//FIXME -- get max products count to display (paginated)
	maxPagesBoundary := 10

	values, _ := query.Values(filterViewModel)
	encodedFilter := values.Encode()

	productsListViewModel.Align(products, filterViewModel, navBarViewModel, page, maxPagesBoundary, encodedFilter)

	if !isHTMXRequest(c) {
		return render(c, views.Products(productsListViewModel))
	}

	productListUrl := fmt.Sprintf("/products/%d?%s", page, encodedFilter)
	isAlreadyOnProductList := strings.HasSuffix(c.Get("HX-Current-Url"), productListUrl)

	if !isAlreadyOnProductList {
		c.Append("HX-Push-Url", productListUrl)
	}

	if filterViewModel.HasError() {
		c.Append("HX-Trigger", "validatePrice")
	}

	if productsListViewModel.DecrementBasketCount || productsListViewModel.IncrementBasketCount {
		return render(c, views.Products(productsListViewModel), "basket-adder-"+productsListViewModel.ChangeCountID)
	}

	return render(c, views.Products(productsListViewModel), views.ProductListFragment)
}

func (h handlers) getProductDetails(c *fiber.Ctx) error {
	var productDetailViewModel views.ProductDetailViewModel
	var navBarViewModel components.NavBarViewModel

	_ = c.QueryParser(&productDetailViewModel)
	_ = c.QueryParser(&navBarViewModel)

	productID := c.Params("productID", "")
	productDetails, err := h.services.GetProductDetails(c.Context(), productID)
	//FIXME?
	if err != nil {
		return c.Redirect("/products/1")
	}

	basketProducts, err := h.services.GetBasket(c.Context(), "")
	//FIXME
	if err != nil {
		return c.Redirect("/products/1")
	}

	navBarViewModel.Align(basketProducts)
	productDetailViewModel.Align(productDetails, navBarViewModel)

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
	if productDetailViewModel.DecrementBasketCount || productDetailViewModel.IncrementBasketCount {
		fragments = append(fragments, "basket-adder")
	}

	isOnProductDetails := strings.Contains(c.Get("Hx-Current-Url"), "/products/details/")

	if !isOnProductDetails {
		return render(c, views.ProductDetails(productDetailViewModel), views.ProductDetailFragment)
	}
	return render(c, views.ProductDetails(productDetailViewModel), fragments...)
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

	basketProducts, err := h.services.GetBasket(c.Context(), "")
	//FIXME
	if err != nil {
		return c.Redirect("/")
	}
	navBarViewModel := components.NewNavBarViewModel("", len(basketProducts))
	if !isHTMXRequest(c) {
		return render(c, views.Dashboard(views.NewDashboardViewModel(promos, slide, promotionPage, navBarViewModel)))
	}

	isPromotionRequest := c.Query("promotions") != ""
	if isPromotionRequest {
		return render(c, views.Dashboard(views.NewDashboardViewModel(promos, slide, promotionPage, navBarViewModel)), views.PromotedProductSelectorFragment)
	}

	return render(c, views.Dashboard(views.NewDashboardViewModel(promos, slide, promotionPage, navBarViewModel)), views.DashboardSelectorFragment)
}
