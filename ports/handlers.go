package ports

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/go-querystring/query"
	"github.com/google/uuid"
	"github.com/siderustler/go-ecommerce/customer"
	"github.com/siderustler/go-ecommerce/ports/views"
	"github.com/siderustler/go-ecommerce/ports/views/components"
	"github.com/siderustler/go-ecommerce/product"
	"github.com/siderustler/go-ecommerce/store"
	store_domain "github.com/siderustler/go-ecommerce/store/domain"
	"github.com/stripe/stripe-go/v83"
	"github.com/stripe/stripe-go/v83/checkout/session"
)

type handlers struct {
	productServices  *product.Services
	storeServices    *store.Services
	customerServices *customer.Services
}

func (h handlers) getProductsRedirect(c *fiber.Ctx) error {
	return c.Redirect("/products/1")
}

func (h handlers) getProducts(c *fiber.Ctx) error {
	page, _ := c.ParamsInt("page")
	userID := c.Cookies("session")

	var filterViewModel views.FilterViewModel
	_ = c.QueryParser(&filterViewModel)
	filterViewModel.Validate()

	var productsListViewModel views.ProductsListViewModel
	_ = c.QueryParser(&productsListViewModel)

	var navBarViewModel components.NavBarViewModel
	_ = c.QueryParser(&navBarViewModel)

	products, err := h.productServices.Products(c.Context(), page, filterViewModel.MapToDomainFilter())
	//FIXME -- display empty product list
	if err != nil {
		fmt.Printf("ERRRO :%+v", err)
		return nil
	}

	cartCount, err := h.storeServices.CartCount(c.Context(), userID)
	//FIXME
	if err != nil {
		return nil
	}

	navBarViewModel.Align(cartCount)

	//FIXME -- get max products count to display (paginated)
	maxPagesBoundary := 10

	values, _ := query.Values(filterViewModel)
	encodedFilter := values.Encode()

	productsListViewModel.Align(products, filterViewModel, navBarViewModel, page, maxPagesBoundary, encodedFilter)

	productListUrl := fmt.Sprintf("/products/%d?%s", page, encodedFilter)
	isAlreadyOnProductList := strings.HasSuffix(c.Get("HX-Current-Url"), productListUrl)

	if !isAlreadyOnProductList {
		c.Append("HX-Push-Url", productListUrl)
	}

	if filterViewModel.HasError() {
		c.Append("HX-Trigger", "validatePrice")
	}

	if productsListViewModel.DecrementBasketCount || productsListViewModel.IncrementBasketCount {
		return renderFragmentOrView(c, views.Products(productsListViewModel), "basket-adder-"+productsListViewModel.ChangeCountID)
	}

	return renderFragmentOrView(c, views.Products(productsListViewModel), views.ProductListFragment)
}

func (h handlers) getProductDetails(c *fiber.Ctx) error {
	var productDetailViewModel views.ProductDetailViewModel
	var navBarViewModel components.NavBarViewModel
	userID := c.Cookies("session")

	_ = c.QueryParser(&productDetailViewModel)
	_ = c.QueryParser(&navBarViewModel)

	productID := c.Params("productID", "")
	productDetails, err := h.productServices.GetProductDetails(c.Context(), productID)
	//FIXME?
	if err != nil {
		return c.Redirect("/products/1")
	}

	cartCount, err := h.storeServices.CartCount(c.Context(), userID)
	//FIXME
	if err != nil {
		return c.Redirect("/products/1")
	}

	navBarViewModel.Align(cartCount)
	productDetailViewModel.Align(productDetails, navBarViewModel)

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
		fragments = append(fragments, views.BasketAdderFragment)
	}

	isOnProductDetails := strings.Contains(c.Get("Hx-Current-Url"), "/products/details/")

	if !isOnProductDetails {
		return renderFragmentOrView(c, views.ProductDetails(productDetailViewModel), views.ProductDetailFragment)
	}
	return renderFragmentOrView(c, views.ProductDetails(productDetailViewModel), fragments...)
}

func (h handlers) getFilterProducts(c *fiber.Ctx) error {
	var filterViewModel views.FilterViewModel
	//FIXME render error
	_ = c.QueryParser(&filterViewModel)
	filterViewModel.Validate()

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
	return renderFragmentOrView(c, views.ProductsFilter(filterViewModel), views.ProductsFilterFragment)
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
	return renderFragmentOrView(c, views.ProductsFilter(filterViewModel), views.PriceFilterFragment)
}

func (h handlers) getDashboard(c *fiber.Ctx) error {
	slide := c.QueryInt("slide", 0)
	promotionPage := c.QueryInt("promotions", 0)
	userID := c.Cookies("session")
	promos, err := h.productServices.GetPromotions(c.Context())
	//FIXME
	if err != nil {
		fmt.Printf("retrieving promotions: %v", err)
		return c.Redirect("/")
	}

	cart, err := h.storeServices.Cart(c.Context(), userID)
	//FIXME
	if err != nil {
		return c.Redirect("/")
	}
	navBarViewModel := components.NewNavBarViewModel("", len(cart.Products))

	isPromotionRequest := c.Query("promotions") != ""
	if isPromotionRequest {
		return renderFragmentOrView(c, views.Dashboard(views.NewDashboardViewModel(promos, slide, promotionPage, navBarViewModel)), views.PromotedProductSelectorFragment)
	}

	return renderFragmentOrView(c, views.Dashboard(views.NewDashboardViewModel(promos, slide, promotionPage, navBarViewModel)), views.DashboardSelectorFragment)
}

func (h handlers) getBasket(c *fiber.Ctx) error {
	//FIXME retrieving search value in navbar while js is not enabled
	var navBarViewModel components.NavBarViewModel
	var basketViewModel views.BasketViewModel

	userID := c.Cookies("session")
	cart, err := h.storeServices.Cart(c.Context(), userID)
	//FIXME?
	if err != nil {
		return c.Redirect("/")
	}
	cartCount := len(cart.Products)
	navBarViewModel.Align(cartCount)

	//FIXME -- create mapper
	productIds := make([]string, 0, cartCount)
	for productID := range cart.Products {
		productIds = append(productIds, productID)
	}
	products, err := h.productServices.ProductsByIDs(c.Context(), productIds)
	//FIXME?
	if err != nil {
		return c.Redirect("/")
	}
	basketViewModel.Align(products, cart.Products, navBarViewModel)

	return renderFragmentOrView(c, views.Basket(basketViewModel), views.BasketFragment)
}

func (h handlers) updateBasket(c *fiber.Ctx) error {
	var basketViewModel views.BasketViewModel
	var navBarViewModel components.NavBarViewModel
	_ = c.BodyParser(&basketViewModel)
	userID := c.Cookies("session")
	var err error
	if basketViewModel.IncBasket {
		err = h.storeServices.AddProductToCart(c.Context(), userID, store_domain.NewCartProduct(basketViewModel.ChangeCountID, 1))
		if err != nil {
			return renderFragmentOrView(c, views.Basket(basketViewModel), views.BasketItemFragment(basketViewModel.ChangeCountID))
		}
	}
	if basketViewModel.DecBasket {
		err = h.storeServices.RemoveProductFromCart(c.Context(), userID, store_domain.NewCartProduct(basketViewModel.ChangeCountID, 1))
		if err != nil {
			return renderFragmentOrView(c, views.Basket(basketViewModel), views.BasketItemFragment(basketViewModel.ChangeCountID))
		}
	}
	cart, err := h.storeServices.Cart(c.Context(), userID)
	if err != nil {
		return renderFragmentOrView(c, views.Basket(basketViewModel), views.BasketItemFragment(basketViewModel.ChangeCountID))
	}
	productIds := make([]string, 0, len(cart.Products))
	for productID := range cart.Products {
		productIds = append(productIds, productID)
	}
	products, err := h.productServices.ProductsByIDs(c.Context(), productIds)
	if err != nil {
		return renderFragmentOrView(c, views.Basket(basketViewModel), views.BasketItemFragment(basketViewModel.ChangeCountID))
	}
	navBarViewModel.Align(len(cart.Products))
	basketViewModel.Align(products, cart.Products, navBarViewModel)

	return renderFragmentOrRedirect(c, views.Basket(basketViewModel), "/basket", views.BasketItemFragment(basketViewModel.ChangeCountID), views.BasketSummaryFragment)
}

func (h handlers) addItemToBasket(c *fiber.Ctx) error {
	userID := c.Cookies("session")
	var basketAdd struct {
		Count     int    `form:"count"`
		ProductID string `form:"productID"`
		Redirect  string `form:"redirect"`
	}
	_ = c.BodyParser(&basketAdd)

	err := h.storeServices.AddProductToCart(c.Context(), userID, store_domain.NewCartProduct(basketAdd.ProductID, basketAdd.Count))
	if err != nil {
		//FIXME
		return c.Redirect(basketAdd.Redirect)
	}
	cartCount, err := h.storeServices.CartCount(c.Context(), userID)
	if err != nil {
		//FIXME
		return c.Redirect(basketAdd.Redirect)
	}

	return renderFragmentOrRedirect(c, components.NavBar(components.NewNavBarViewModel("", cartCount)), basketAdd.Redirect, components.BasketCountFragment)
}

func (h handlers) getBillingInfo(c *fiber.Ctx) error {
	var billingInfoViewModel views.BillingInfoViewModel
	var navBarViewModel components.NavBarViewModel
	//FIXME retrieving search value in navbar while js is not enabled (use form or a tag and messy query?)
	userID := c.Cookies("userID")

	cartCount, err := h.storeServices.CartCount(c.Context(), userID)
	//FIXME?
	if err != nil {
		return c.Redirect("/")
	}
	navBarViewModel.Align(cartCount)
	billingInfoViewModel.Align(navBarViewModel)

	return renderFragmentOrView(c, views.BillingInfo(billingInfoViewModel), views.BillingInfoFragment)
}

func (h handlers) postBillingInfo(c *fiber.Ctx) error {
	var billingInfoViewModel views.BillingInfoViewModel
	var navBarViewModel components.NavBarViewModel
	//FIXME retrieving search value in navbar while js is not enabled (use form or a tag and messy query?)
	userID := c.Cookies("userID")
	_ = c.BodyParser(&billingInfoViewModel)

	cartCount, err := h.storeServices.CartCount(c.Context(), userID)
	//FIXME?
	if err != nil {
		return c.Redirect("/")
	}
	navBarViewModel.Align(cartCount)
	billingInfoViewModel.Align(navBarViewModel)
	customer, err := billingInfoViewModel.ParseToDomainCustomer()
	if err != nil {
		fmt.Printf("error occured creating customer: %+v", err)
		billingInfoViewModel.MapDomainErrorToViewModelError(err)
		return renderFragmentOrView(c, views.BillingInfo(billingInfoViewModel), views.BillingInfoFragment)
	}

	err = h.customerServices.CreateCustomer(c.Context(), customer)
	if err != nil {
		fmt.Printf("ERROR")
		//FIXME
		// DISPLAY TOAST OTN ERROR
		fmt.Printf("error occured creating customer: %+v", err)
		return renderFragmentOrView(c, views.BillingInfo(billingInfoViewModel), views.BillingInfoFragment)
	}
	c.Cookie(&fiber.Cookie{Name: "userID", Value: customer.ID})
	if !billingInfoViewModel.UseBillingAddressAsShipping {
		var shippingInfoViewModel views.ShippingInfoViewModel
		shippingInfoViewModel.Align(navBarViewModel)
		addShippingAddressUrl := "/basket/customer/shipping"

		c.Append("Hx-Push-Url", addShippingAddressUrl)
		return renderFragmentOrRedirect(c, views.ShippingInfo(shippingInfoViewModel), addShippingAddressUrl, views.ShippingInfoFragment)
	}
	paymentUrl := "/basket/checkout"
	c.Append("Hx-Push-Url", paymentUrl)

	return renderFragmentOrRedirect(c, views.Checkout(views.NewCheckoutViewModel(false)), paymentUrl, views.CheckoutFragment)
}

func (h handlers) getShippingInfo(c *fiber.Ctx) error {
	var shippingInfoViewModel views.ShippingInfoViewModel
	var navBarViewModel components.NavBarViewModel
	//FIXME retrieving search value in navbar while js is not enabled (use form or a tag and messy query?)
	userID := c.Cookies("userID")

	cartCount, err := h.storeServices.CartCount(c.Context(), userID)
	//FIXME?
	if err != nil {
		return c.Redirect("/")
	}
	navBarViewModel.Align(cartCount)
	shippingInfoViewModel.Align(navBarViewModel)

	return renderFragmentOrView(c, views.ShippingInfo(shippingInfoViewModel), views.ShippingInfoFragment)
}

func (h handlers) postShippingInfo(c *fiber.Ctx) error {
	var navBarViewModel components.NavBarViewModel
	var shippingInfoViewModel views.ShippingInfoViewModel
	id := c.Cookies("userID")
	//FIXME retrieving search value in navbar while js is not enabled (use form or a tag and messy query?)
	_ = c.BodyParser(&shippingInfoViewModel)

	cartCount, err := h.storeServices.CartCount(c.Context(), id)
	//FIXME?
	if err != nil {
		return c.Redirect("/")
	}
	navBarViewModel.Align(cartCount)
	shippingInfoViewModel.Align(navBarViewModel)

	shipping, err := shippingInfoViewModel.ParseToDomainShippingAddress(uuid.NewString())
	if err != nil {
		shippingInfoViewModel.MapDomainErrorToViewModelError(err)
		return renderFragmentOrView(c, views.ShippingInfo(shippingInfoViewModel), views.ShippingInfoFragment)
	}
	err = h.customerServices.AddShippingAddress(c.Context(), id, shipping)
	if err != nil {
		fmt.Printf("error: %v", err)
		//FIXME -- toast on error
		return renderFragmentOrView(c, views.ShippingInfo(shippingInfoViewModel), views.ShippingInfoFragment)
	}
	paymentUrl := "/basket/checkout"
	c.Append("Hx-Push-Url", paymentUrl)

	return renderFragmentOrRedirect(c, views.Checkout(views.NewCheckoutViewModel(false)), paymentUrl, views.CheckoutFragment)
}

func (h handlers) getCheckout(c *fiber.Ctx) error {
	checkoutSession := c.Query("session_id")

	s, err := session.Get(checkoutSession, nil)
	checkoutFinalized := err == nil && s.Status == "complete"
	checkoutViewModel := views.NewCheckoutViewModel(checkoutFinalized)

	return renderFragmentOrView(c, views.Checkout(checkoutViewModel), views.CheckoutFragment)
}

func (h handlers) createCheckout(c *fiber.Ctx) error {
	//move it to services
	params := &stripe.CheckoutSessionParams{
		Mode:      stripe.String(string(stripe.CheckoutSessionModePayment)),
		UIMode:    stripe.String("embedded"),
		ReturnURL: stripe.String("http://localhost:8080/basket/checkout?session_id={CHECKOUT_SESSION_ID}"),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String("usd"),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String("T-shirt"),
					},
					UnitAmount: stripe.Int64(2000),
				},
				Quantity: stripe.Int64(1),
			},
		},
	}

	s, err := session.New(params)

	if err != nil {
		return err
	}

	data := struct {
		ClientSecret string `json:"clientSecret"`
	}{
		ClientSecret: s.ClientSecret,
	}

	return c.JSON(data)
}
