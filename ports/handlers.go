package ports

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/go-querystring/query"
	"github.com/google/uuid"
	"github.com/siderustler/go-ecommerce/ports/views"
	"github.com/siderustler/go-ecommerce/ports/views/components"
	"github.com/siderustler/go-ecommerce/services"
	"github.com/stripe/stripe-go/v83"
	"github.com/stripe/stripe-go/v83/checkout/session"
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
		fragments = append(fragments, views.BasketAdderFragment)
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

func (h handlers) getBasket(c *fiber.Ctx) error {
	var navBarViewModel components.NavBarViewModel
	//FIXME retrieving search value in navbar while js is not enabled

	session := c.Cookies("session")
	basketProducts, err := h.services.GetBasket(c.Context(), session)
	//FIXME?
	if err != nil {
		return c.Redirect("/")
	}
	navBarViewModel.Align(basketProducts)
	basketViewModel := views.NewBasketViewModel(basketProducts, navBarViewModel)

	if !isHTMXRequest(c) {
		return render(c, views.Basket(basketViewModel))
	}

	return render(c, views.Basket(basketViewModel), views.BasketFragment)
}

func (h handlers) updateBasket(c *fiber.Ctx) error {
	var basketViewModel views.BasketViewModel
	_ = c.BodyParser(&basketViewModel)

	session := c.Cookies("session")
	if basketViewModel.DecBasket {
		_ = h.services.DecrementBasketItem(c.Context(), session, basketViewModel.ChangeCountID)
	}
	if basketViewModel.IncBasket {
		_ = h.services.IncrementBasketItem(c.Context(), session, basketViewModel.ChangeCountID)
	}
	basketProducts, _ := h.services.GetBasket(c.Context(), session)

	basketViewModel.Align(basketProducts, components.NavBarViewModel{})
	if isHTMXRequest(c) {
		return render(c, views.Basket(basketViewModel), views.BasketItemFragment(basketViewModel.ChangeCountID))
	}
	return c.Redirect("/basket")
}

func (h handlers) addItemToBasket(c *fiber.Ctx) error {
	session := c.Cookies("session")

	var basketAdd struct {
		Count     int    `form:"count"`
		ProductID string `form:"productID"`
		Redirect  string `form:"redirect"`
	}
	_ = c.BodyParser(&basketAdd)
	isBasketSessionEstablished := session != ""
	if !isBasketSessionEstablished {
		session = uuid.NewString()
	}
	c.Cookie(&fiber.Cookie{Name: "session", Value: session})
	basketCount, _ := h.services.AddToBasket(c.Context(), session, basketAdd.ProductID, basketAdd.Count)
	if isHTMXRequest(c) {
		return render(c, components.NavBar(components.NewNavBarViewModel("", basketCount)), components.BasketCountFragment)
	}
	return c.Redirect(basketAdd.Redirect)
}

func (h handlers) getBillingInfo(c *fiber.Ctx) error {
	var billingInfoViewModel views.BillingInfoViewModel
	var navBarViewModel components.NavBarViewModel
	//FIXME retrieving search value in navbar while js is not enabled (use form or a tag and messy query?)
	session := c.Cookies("session")

	basketProducts, err := h.services.GetBasket(c.Context(), session)
	//FIXME?
	if err != nil {
		return c.Redirect("/")
	}
	navBarViewModel.Align(basketProducts)
	billingInfoViewModel.Align(navBarViewModel)
	if isHTMXRequest(c) {
		return render(c, views.BillingInfo(billingInfoViewModel), views.BillingInfoFragment)
	}
	return render(c, views.BillingInfo(billingInfoViewModel))
}

func (h handlers) postBillingInfo(c *fiber.Ctx) error {
	var billingInfoViewModel views.BillingInfoViewModel
	var navBarViewModel components.NavBarViewModel
	//FIXME retrieving search value in navbar while js is not enabled (use form or a tag and messy query?)
	session := c.Cookies("session")
	_ = c.BodyParser(&billingInfoViewModel)

	basketProducts, err := h.services.GetBasket(c.Context(), session)
	//FIXME?
	if err != nil {
		return c.Redirect("/")
	}
	navBarViewModel.Align(basketProducts)
	billingInfoViewModel.Align(navBarViewModel)

	if billingInfoViewModel.HasError() {
		if isHTMXRequest(c) {
			return render(c, views.BillingInfo(billingInfoViewModel), views.BillingInfoFragment)
		}
		return render(c, views.BillingInfo(billingInfoViewModel))
	}

	customer := billingInfoViewModel.MapToDomainCustomer()
	err = h.services.CreateCustomer(c.Context(), session, customer)
	if err != nil {
		if isHTMXRequest(c) {
			//FIXME
			// DISPLAY TOAST OTN ERROR
			fmt.Printf("error occured creating customer: %+v", err)
			return render(c, views.BillingInfo(billingInfoViewModel), views.BillingInfoFragment)
		}
		return render(c, views.BillingInfo(billingInfoViewModel))
	}

	if !billingInfoViewModel.UseBillingAddressAsShipping {
		var shippingInfoViewModel views.ShippingInfoViewModel
		shippingInfoViewModel.Align(navBarViewModel)
		addShippingAddressUrl := "/basket/customer/shipping"
		if isHTMXRequest(c) {
			c.Append("Hx-Push-Url", addShippingAddressUrl)
			return render(c, views.ShippingInfo(shippingInfoViewModel), views.ShippingInfoFragment)
		}
		return c.Redirect(addShippingAddressUrl)
	}
	paymentUrl := "/basket/checkout"
	if isHTMXRequest(c) {
		c.Append("Hx-Push-Url", paymentUrl)
		return render(c, views.Checkout(views.NewCheckoutViewModel(false)), views.CheckoutFragment)
	}
	//FIXME -- redirect to payment
	return c.Redirect(paymentUrl)
}

func (h handlers) getShippingInfo(c *fiber.Ctx) error {
	var shippingInfoViewModel views.ShippingInfoViewModel
	var navBarViewModel components.NavBarViewModel
	//FIXME retrieving search value in navbar while js is not enabled (use form or a tag and messy query?)
	session := c.Cookies("session")

	basketProducts, err := h.services.GetBasket(c.Context(), session)
	//FIXME?
	if err != nil {
		return c.Redirect("/")
	}
	navBarViewModel.Align(basketProducts)
	shippingInfoViewModel.Align(navBarViewModel)

	if isHTMXRequest(c) {
		return render(c, views.ShippingInfo(shippingInfoViewModel), views.ShippingInfoFragment)
	}
	return render(c, views.ShippingInfo(shippingInfoViewModel))
}

func (h handlers) postShippingInfo(c *fiber.Ctx) error {
	var navBarViewModel components.NavBarViewModel
	var shippingInfoViewModel views.ShippingInfoViewModel
	//FIXME retrieving search value in navbar while js is not enabled (use form or a tag and messy query?)
	session := c.Cookies("session")
	_ = c.BodyParser(&shippingInfoViewModel)

	basketProducts, err := h.services.GetBasket(c.Context(), session)
	//FIXME?
	if err != nil {
		return c.Redirect("/")
	}
	navBarViewModel.Align(basketProducts)
	shippingInfoViewModel.Align(navBarViewModel)

	if shippingInfoViewModel.HasError() {
		if isHTMXRequest(c) {
			return render(c, views.ShippingInfo(shippingInfoViewModel), views.ShippingInfoFragment)
		}
		return render(c, views.ShippingInfo(shippingInfoViewModel))
	}
	err = h.services.CreateShippingInfo(c.Context(), session, shippingInfoViewModel.MapToDomainShippingAddress())
	if err != nil {
		if isHTMXRequest(c) {
			//FIXME -- toast on error
			return render(c, views.ShippingInfo(shippingInfoViewModel), views.ShippingInfoFragment)
		}
		return render(c, views.ShippingInfo(shippingInfoViewModel))
	}
	paymentUrl := "/basket/checkout"
	if isHTMXRequest(c) {
		c.Append("Hx-Push-Url", paymentUrl)
		return render(c, views.Checkout(views.NewCheckoutViewModel(false)), views.CheckoutFragment)
	}
	return c.Redirect(paymentUrl)
}

func (h handlers) getCheckout(c *fiber.Ctx) error {
	checkoutSession := c.Query("session_id")

	s, err := session.Get(checkoutSession, nil)
	checkoutFinalized := err == nil && s.Status == "complete"
	checkoutViewModel := views.NewCheckoutViewModel(checkoutFinalized)

	if isHTMXRequest(c) {
		return render(c, views.Checkout(checkoutViewModel), views.CheckoutFragment)
	}
	return render(c, views.Checkout(checkoutViewModel))
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
