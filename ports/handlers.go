package ports

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/google/go-querystring/query"
	"github.com/google/uuid"
	"github.com/siderustler/go-ecommerce/customer"
	"github.com/siderustler/go-ecommerce/ports/auth"
	"github.com/siderustler/go-ecommerce/ports/views"
	"github.com/siderustler/go-ecommerce/ports/views/components"
	"github.com/siderustler/go-ecommerce/product"
	"github.com/siderustler/go-ecommerce/store"
	store_domain "github.com/siderustler/go-ecommerce/store/domain"
	"github.com/stripe/stripe-go/v84"
	stripeSession "github.com/stripe/stripe-go/v84/checkout/session"
	"github.com/stripe/stripe-go/v84/webhook"
	"golang.org/x/oauth2"
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
	userID := auth.UserIDFromContext(c.UserContext())

	var filterViewModel views.FilterViewModel
	_ = c.QueryParser(&filterViewModel)
	filterViewModel.Validate()
	limit := 5

	var productsListViewModel views.ProductsListViewModel
	_ = c.QueryParser(&productsListViewModel)

	var navBarViewModel components.NavBarViewModel
	_ = c.QueryParser(&navBarViewModel)

	products, maxProductCount, err := h.productServices.Products(c.Context(), page, limit, filterViewModel.MapToDomainFilter())
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

	values, _ := query.Values(filterViewModel)
	encodedFilter := values.Encode()

	productsListViewModel.Align(products, filterViewModel, navBarViewModel, page, maxProductCount/limit, encodedFilter)

	productListUrl := fmt.Sprintf("/products/%d?%s", page, encodedFilter)
	isAlreadyOnProductList := strings.HasSuffix(c.Get("Hx-Current-Url"), productListUrl)

	if !isAlreadyOnProductList {
		c.Append("Hx-Push-Url", productListUrl)
	}

	if productsListViewModel.DecrementBasketCount || productsListViewModel.IncrementBasketCount {
		return renderFragmentOrView(c, views.Products(productsListViewModel), "basket-adder-"+productsListViewModel.ChangeCountID)
	}

	return renderFragmentOrView(c, views.Products(productsListViewModel), views.ProductListFragment)
}

func (h handlers) getProductDetails(c *fiber.Ctx) error {
	var productDetailViewModel views.ProductDetailViewModel
	var navBarViewModel components.NavBarViewModel
	userID := auth.UserIDFromContext(c.UserContext())

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

func (h handlers) getDashboard(c *fiber.Ctx) error {
	slide := c.QueryInt("slide", 1)
	promotionPage := c.QueryInt("promotions", 1)
	userID := auth.UserIDFromContext(c.UserContext())
	maxPromosPerPage := 3
	promos, promoCount, err := h.productServices.Promotions(c.Context(), promotionPage, maxPromosPerPage)
	//FIXME
	if err != nil {
		fmt.Printf("retrieving promotions: %v", err)
	}

	cartCount, err := h.storeServices.CartCount(c.Context(), userID)
	//FIXME
	if err != nil {
		return c.Redirect("/")
	}
	navBarViewModel := components.NewNavBarViewModel("", cartCount)

	isPromotionRequest := c.Query("promotions") != ""
	isSliderRequest := c.Query("slide") != ""
	if isPromotionRequest || isSliderRequest {
		return renderFragmentOrView(c, views.Dashboard(views.NewDashboardViewModel(promos, slide, promotionPage, promoCount, maxPromosPerPage, navBarViewModel)), views.PromotedProductSelectorFragment)
	}

	return renderFragmentOrView(c, views.Dashboard(views.NewDashboardViewModel(promos, slide, promotionPage, promoCount, maxPromosPerPage, navBarViewModel)), views.DashboardSelectorFragment)
}

func (h handlers) getBasket(c *fiber.Ctx) error {
	//FIXME retrieving search value in navbar while js is not enabled
	var navBarViewModel components.NavBarViewModel
	var basketViewModel views.BasketViewModel

	userID := auth.UserIDFromContext(c.UserContext())
	cart, err := h.storeServices.Cart(c.Context(), userID)
	//FIXME?
	if err != nil {
		return c.Redirect("/")
	}
	cartCount := len(cart.Products)
	navBarViewModel.Align(cartCount)

	//FIXME -- create mapper
	productIds := slices.Collect(maps.Keys(cart.Products))
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
	userID := auth.UserIDFromContext(c.UserContext())

	var err error
	if basketViewModel.IncBasketItem {
		err = h.storeServices.AddProductToCart(c.Context(), userID, store_domain.NewCartProduct(basketViewModel.ChangeCountID, 1))
		if err != nil {
			return renderFragmentOrView(c, views.Basket(basketViewModel), views.BasketItemFragment(basketViewModel.ChangeCountID))
		}
	}
	if basketViewModel.DecBasketItem || basketViewModel.RemoveBasketItem {
		err = h.storeServices.RemoveProductFromCart(c.Context(), userID, store_domain.NewCartProduct(basketViewModel.ChangeCountID, 1))
		if err != nil {
			return renderFragmentOrView(c, views.Basket(basketViewModel), views.BasketItemFragment(basketViewModel.ChangeCountID))
		}
	}
	cart, err := h.storeServices.Cart(c.Context(), userID)
	if err != nil {
		return renderFragmentOrView(c, views.Basket(basketViewModel), views.BasketItemFragment(basketViewModel.ChangeCountID))
	}

	products, err := h.productServices.ProductsByIDs(c.Context(), slices.Collect(maps.Keys(cart.Products)))
	if err != nil {
		return renderFragmentOrView(c, views.Basket(basketViewModel), views.BasketItemFragment(basketViewModel.ChangeCountID))
	}
	navBarViewModel.Align(len(cart.Products))
	basketViewModel.Align(products, cart.Products, navBarViewModel)
	isCartEmpty := len(cart.Products) < 1
	if isCartEmpty {
		return renderFragmentOrView(c, views.Basket(basketViewModel), views.EmptyBasketFragment, views.BasketSummaryFragment, components.BasketCountFragment)
	}
	if basketViewModel.RemoveBasketItem {
		return renderFragmentOrView(c, views.Basket(basketViewModel), views.BasketSummaryFragment, components.BasketCountFragment)
	}
	return renderFragmentOrRedirect(c, views.Basket(basketViewModel), "/basket", views.BasketItemFragment(basketViewModel.ChangeCountID), views.BasketSummaryFragment)
}

func (h handlers) addItemToBasket(c *fiber.Ctx) error {
	userID := auth.UserIDFromContext(c.UserContext())
	var basketAdd struct {
		Count     int    `form:"count"`
		ProductID string `form:"productID"`
		Redirect  string `form:"redirect"`
	}
	_ = c.BodyParser(&basketAdd)

	err := h.storeServices.AddProductToCart(c.Context(), userID, store_domain.NewCartProduct(basketAdd.ProductID, basketAdd.Count))
	if err != nil {
		//FIXME\
		fmt.Printf("ERR: %+v", err)
		return c.Redirect(basketAdd.Redirect)
	}
	cartCount, err := h.storeServices.CartCount(c.Context(), userID)
	if err != nil {
		//FIXME
		fmt.Printf("ERR: %+v", err)

		return c.Redirect(basketAdd.Redirect)
	}

	return renderFragmentOrRedirect(c, components.NavBar(components.NewNavBarViewModel("", cartCount)), basketAdd.Redirect, components.BasketCountFragment)
}

func (h handlers) getBasketBillingInfo(c *fiber.Ctx) error {
	var billingInfoViewModel views.BillingInfoViewModel
	var navBarViewModel components.NavBarViewModel
	//FIXME retrieving search value in navbar while js is not enabled (use form or a tag and messy query?)
	userID := auth.UserIDFromContext(c.UserContext())

	cartCount, err := h.storeServices.CartCount(c.Context(), userID)
	//FIXME?
	if err != nil {
		return c.Redirect("/")
	}
	customer, _ := h.customerServices.Customer(c.Context(), userID)

	navBarViewModel.Align(cartCount)
	billingInfoViewModel.WithNavBarViewModel(navBarViewModel)

	hasShipping := !customer.Shipping.IsZero()
	hasBilling := !customer.Billing.IsZero()
	if hasShipping {
		checkoutStartUrl := "/basket/checkout"
		c.Append("Hx-Push-Url", checkoutStartUrl)
		return renderFragmentOrRedirect(c, views.CheckoutStart(navBarViewModel), checkoutStartUrl, views.CheckoutFragment)
	}
	if hasBilling {
		shippingUrl := "/basket/customer/shipping"
		var shippingInfoViewModel views.ShippingInfoViewModel
		shippingInfoViewModel.WithNavBarViewModel(navBarViewModel)
		c.Append("Hx-Push-Url", shippingUrl)
		return renderFragmentOrRedirect(c, views.BasketShippingInfo(shippingInfoViewModel), shippingUrl, views.ShippingInfoFragment)
	}
	return renderFragmentOrView(c, views.BasketBillingInfo(billingInfoViewModel), views.BillingInfoFragment)
}

func (h handlers) postBillingInfo(c *fiber.Ctx) error {
	var billingInfoViewModel views.BillingInfoViewModel
	var navBarViewModel components.NavBarViewModel
	//FIXME retrieving search value in navbar while js is not enabled (use form or a tag and messy query?)
	userID := auth.UserIDFromContext(c.UserContext())
	_ = c.BodyParser(&billingInfoViewModel)

	cartCount, err := h.storeServices.CartCount(c.Context(), userID)
	//FIXME?
	if err != nil {
		return c.Redirect("/")
	}
	navBarViewModel.Align(cartCount)
	billingInfoViewModel.WithNavBarViewModel(navBarViewModel)
	customer, err := billingInfoViewModel.ParseToDomainCustomer(userID)
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
	accountUrl := "/account"
	c.Append("Hx-Push-Url", accountUrl)

	return renderFragmentOrRedirect(c, views.AccountProfile(views.NewProfileViewModel(customer, navBarViewModel)), accountUrl, views.ProfileFragment)
}

func (h handlers) getBillingInfo(c *fiber.Ctx) error {
	var billingInfoViewModel views.BillingInfoViewModel
	var navBarViewModel components.NavBarViewModel
	//FIXME retrieving search value in navbar while js is not enabled (use form or a tag and messy query?)
	userID := auth.UserIDFromContext(c.UserContext())

	cartCount, err := h.storeServices.CartCount(c.Context(), userID)
	//FIXME?
	if err != nil {
		return c.Redirect("/")
	}
	navBarViewModel.Align(cartCount)
	billingInfoViewModel.WithNavBarViewModel(navBarViewModel)
	customer, err := h.customerServices.Customer(c.Context(), userID)
	if err != nil {
		return c.Redirect("/")
	}
	billingInfoViewModel.WithCustomer(customer)
	c.Append("Hx-Push-Url", "/account/customer/billing")
	return renderFragmentOrView(c, views.BillingInfo(billingInfoViewModel), views.BillingInfoFragment)
}

func (h handlers) postBasketBillingInfo(c *fiber.Ctx) error {
	var billingInfoViewModel views.BillingInfoViewModel
	var navBarViewModel components.NavBarViewModel
	//FIXME retrieving search value in navbar while js is not enabled (use form or a tag and messy query?)
	userID := auth.UserIDFromContext(c.UserContext())
	_ = c.BodyParser(&billingInfoViewModel)

	cartCount, err := h.storeServices.CartCount(c.Context(), userID)
	//FIXME?
	if err != nil {
		return c.Redirect("/")
	}
	navBarViewModel.Align(cartCount)
	billingInfoViewModel.WithNavBarViewModel(navBarViewModel)
	customer, err := billingInfoViewModel.ParseToDomainCustomer(userID)
	if err != nil {
		fmt.Printf("error occured creating customer: %+v", err)
		billingInfoViewModel.MapDomainErrorToViewModelError(err)
		return renderFragmentOrView(c, views.BasketBillingInfo(billingInfoViewModel), views.BillingInfoFragment)
	}
	err = h.customerServices.CreateCustomer(c.Context(), customer)
	if err != nil {
		fmt.Printf("ERROR")
		//FIXME
		// DISPLAY TOAST OTN ERROR
		fmt.Printf("error occured creating customer: %+v", err)
		return renderFragmentOrView(c, views.BasketBillingInfo(billingInfoViewModel), views.BillingInfoFragment)
	}
	if !billingInfoViewModel.UseBillingAddressAsShipping {
		var shippingInfoViewModel views.ShippingInfoViewModel
		shippingInfoViewModel.WithNavBarViewModel(navBarViewModel)
		addShippingAddressUrl := "/basket/customer/shipping"

		c.Append("Hx-Push-Url", addShippingAddressUrl)
		return renderFragmentOrRedirect(c, views.BasketShippingInfo(shippingInfoViewModel), addShippingAddressUrl, views.ShippingInfoFragment)
	}
	paymentUrl := "/basket/checkout"
	c.Append("Hx-Push-Url", paymentUrl)

	return renderFragmentOrRedirect(c, views.CheckoutStart(navBarViewModel), paymentUrl, views.CheckoutFragment)
}

func (h handlers) getShippingInfo(c *fiber.Ctx) error {
	var shippingInfoViewModel views.ShippingInfoViewModel
	var navBarViewModel components.NavBarViewModel
	//FIXME retrieving search value in navbar while js is not enabled (use form or a tag and messy query?)
	userID := auth.UserIDFromContext(c.UserContext())

	cartCount, err := h.storeServices.CartCount(c.Context(), userID)
	//FIXME?
	if err != nil {
		return c.Redirect("/")
	}
	customer, _ := h.customerServices.Customer(c.Context(), userID)
	navBarViewModel.Align(cartCount)
	shippingInfoViewModel.WithNavBarViewModel(navBarViewModel).WithShipping(customer.Shipping)
	shippingUrl := "/account/customer/shipping"
	c.Append("Hx-Push-Url", shippingUrl)

	return renderFragmentOrView(c, views.ShippingInfo(shippingInfoViewModel), views.ShippingInfoFragment)
}

func (h handlers) postShippingInfo(c *fiber.Ctx) error {
	var navBarViewModel components.NavBarViewModel
	var shippingInfoViewModel views.ShippingInfoViewModel
	id := auth.UserIDFromContext(c.UserContext())
	//FIXME retrieving search value in navbar while js is not enabled (use form or a tag and messy query?)
	_ = c.BodyParser(&shippingInfoViewModel)

	cartCount, err := h.storeServices.CartCount(c.Context(), id)
	//FIXME?
	if err != nil {
		return c.Redirect("/")
	}

	navBarViewModel.Align(cartCount)
	shippingInfoViewModel.WithNavBarViewModel(navBarViewModel)

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
	accountUrl := "/account"
	c.Append("Hx-Push-Url", accountUrl)

	customer, err := h.customerServices.Customer(c.Context(), id)
	if err != nil {
		fmt.Printf("error: %v", err)
		//FIXME -- toast on error
		return renderFragmentOrView(c, views.ShippingInfo(shippingInfoViewModel), views.ShippingInfoFragment)
	}
	return renderFragmentOrRedirect(c, views.AccountProfile(views.NewProfileViewModel(customer, navBarViewModel)), accountUrl, views.ProfileFragment)
}

func (h handlers) getBasketShippingInfo(c *fiber.Ctx) error {
	var shippingInfoViewModel views.ShippingInfoViewModel
	var navBarViewModel components.NavBarViewModel
	//FIXME retrieving search value in navbar while js is not enabled (use form or a tag and messy query?)
	userID := auth.UserIDFromContext(c.UserContext())

	cartCount, err := h.storeServices.CartCount(c.Context(), userID)
	//FIXME?
	if err != nil {
		return c.Redirect("/")
	}
	navBarViewModel.Align(cartCount)
	shippingInfoViewModel.WithNavBarViewModel(navBarViewModel)
	shippingUrl := "/basket/customer/shipping"
	c.Append("Hx-Push-Url", shippingUrl)

	return renderFragmentOrView(c, views.BasketShippingInfo(shippingInfoViewModel), views.ShippingInfoFragment)
}

func (h handlers) postBasketShippingInfo(c *fiber.Ctx) error {
	var navBarViewModel components.NavBarViewModel
	var shippingInfoViewModel views.ShippingInfoViewModel
	id := auth.UserIDFromContext(c.UserContext())
	//FIXME retrieving search value in navbar while js is not enabled (use form or a tag and messy query?)
	_ = c.BodyParser(&shippingInfoViewModel)

	cartCount, err := h.storeServices.CartCount(c.Context(), id)
	//FIXME?
	if err != nil {
		return c.Redirect("/")
	}
	navBarViewModel.Align(cartCount)
	shippingInfoViewModel.WithNavBarViewModel(navBarViewModel)

	shipping, err := shippingInfoViewModel.ParseToDomainShippingAddress(uuid.NewString())
	if err != nil {
		shippingInfoViewModel.MapDomainErrorToViewModelError(err)
		return renderFragmentOrView(c, views.BasketShippingInfo(shippingInfoViewModel), views.ShippingInfoFragment)
	}
	err = h.customerServices.AddShippingAddress(c.Context(), id, shipping)
	if err != nil {
		fmt.Printf("error: %v", err)
		//FIXME -- toast on error
		return renderFragmentOrView(c, views.BasketShippingInfo(shippingInfoViewModel), views.ShippingInfoFragment)
	}
	paymentUrl := "/basket/checkout"
	c.Append("Hx-Push-Url", paymentUrl)

	return renderFragmentOrRedirect(c, views.CheckoutStart(navBarViewModel), paymentUrl, views.CheckoutFragment)
}

func (h handlers) getCheckoutStart(c *fiber.Ctx) error {
	var navBarViewModel components.NavBarViewModel
	userID := auth.UserIDFromContext(c.UserContext())
	cartCount, err := h.storeServices.CartCount(c.Context(), userID)
	if isCartEmpty := cartCount == 0; err != nil || isCartEmpty {
		return c.Redirect("/basket")
	}
	customer, err := h.customerServices.Customer(c.Context(), userID)
	if err != nil || customer.IsZero() {
		return c.Redirect("/basket")
	}
	navBarViewModel.Align(cartCount)

	return renderFragmentOrView(c, views.CheckoutStart(navBarViewModel), views.CheckoutFragment)
}

func (h handlers) getCheckoutFinalized(c *fiber.Ctx) error {
	var navBarViewModel components.NavBarViewModel

	checkoutSession := c.Query("session_id")
	s, err := stripeSession.Get(checkoutSession, nil)
	if err != nil {
		//FIXME
		return c.Redirect("/basket/checkout")
	}
	checkoutPaidSuccessfully := s.Status == "complete"
	checkoutViewModel := views.NewCheckoutViewModel(checkoutPaidSuccessfully, navBarViewModel)
	c.Append("HX-Push-Url", "/basket/checkout/finalize")
	return renderFragmentOrView(c, views.CheckoutFinalized(checkoutViewModel), views.CheckoutFragment, components.BasketCountFragment)
}

func (h handlers) createCheckout(c *fiber.Ctx) error {
	//move it to services
	userID := auth.UserIDFromContext(c.UserContext())
	type errStruct struct {
		Err string `json:"error"`
	}
	err := h.storeServices.CreateCheckout(c.Context(), userID)
	if err != nil {
		fmt.Printf("error is :%+v", err)
		return c.Status(http.StatusBadRequest).JSON(errStruct{Err: fmt.Sprintf("error creating cehckout: %v", err.Error())})
	}
	checkout, err := h.storeServices.CheckoutByUserID(c.Context(), userID)
	if err != nil {
		fmt.Printf("error is CHECKOUTBYSEURID:%+v", err)
		return c.Status(http.StatusBadRequest).JSON(errStruct{Err: fmt.Sprintf("retrieving checkout: %v", err.Error())})
	}
	products, err := h.productServices.ProductsByIDs(c.Context(), slices.Collect(maps.Keys(checkout.Items)))
	if err != nil {
		fmt.Printf("error is PROCUCST:%+v", err)

		return c.Status(http.StatusInternalServerError).JSON(errStruct{Err: fmt.Sprintf("error retrieving products for checkout creation: %v", err.Error())})
	}
	sess, err := h.storeServices.CreateStripeCheckout(c.Context(), checkout.ID, checkout.Items, products)
	if err != nil {
		fmt.Printf("error is CSTIRPE CHEKCOUT:%+v", err)

		return c.Status(http.StatusInternalServerError).JSON(errStruct{Err: fmt.Sprintf("error creating stripe checkout: %v", err.Error())})
	}
	data := struct {
		ClientSecret string `json:"clientSecret"`
	}{
		ClientSecret: sess.ClientSecret,
	}

	return c.JSON(data)
}

func oauthLoginHandler(auth *auth.Authenticator, sessionStore *session.Store) func(c *fiber.Ctx) error {
	generateRandomState := func() (string, error) {
		b := make([]byte, 32)
		_, err := rand.Read(b)
		if err != nil {
			return "", err
		}

		state := base64.StdEncoding.EncodeToString(b)

		return state, nil
	}
	return func(c *fiber.Ctx) error {
		state, err := generateRandomState()
		if err != nil {
			return c.SendString(err.Error())
		}
		sess, err := sessionStore.Get(c)
		if err != nil {
			return c.SendString("retrieving session: " + err.Error())
		}
		sess.Set("state", state)
		err = sess.Save()
		if err != nil {
			return c.SendString("saving session: " + err.Error())
		}
		return c.Redirect(auth.AuthCodeURL(state, oauth2.AccessTypeOffline), http.StatusTemporaryRedirect)
	}
}

func (h handlers) oauthCallbackHandler(oauth *auth.Authenticator, sessionStore *session.Store) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		sess, err := sessionStore.Get(c)
		if err != nil {
			return c.SendString("retrieving session: " + err.Error())
		}
		if c.Query("state") != sess.Get("state") {
			return c.SendString("Invalid state parameter")
		}

		token, err := oauth.Exchange(c.Context(), c.Query("code"))
		if err != nil {
			return c.SendString("Failed to exchange an authorization code for a token.")
		}

		idToken, err := oauth.VerifyIDToken(c.Context(), token)
		if err != nil {
			return c.SendString("Failed to verify ID Token.")
		}

		previousUser := sess.Get("user_id").(string)
		sess.Set("ip", c.IP())
		expiryTime := time.Now().Unix() + int64(auth.TokenExpiryTime)
		sess.Set("expiry", expiryTime)
		sess.Set("user_id", idToken.Subject)
		if err = sess.Save(); err != nil {
			return c.SendString("saving session: " + err.Error())
		}
		if err = h.customerServices.CreateCustomer(c.Context(), customer.NewCustomer(
			idToken.Subject,
			customer.Credentials{},
			customer.Billing{},
			customer.ShippingAddress{},
		)); err != nil {
			return c.SendString("creating user: " + err.Error())
		}
		if err = h.storeServices.MergeUserCarts(c.Context(), previousUser, idToken.Subject); err != nil {
			return c.SendString("merging cart: " + err.Error())
		}
		return c.Redirect("/account", http.StatusTemporaryRedirect)
	}
}

func (h handlers) accountHandler(c *fiber.Ctx) error {
	userID := auth.UserIDFromContext(c.UserContext())
	var navBarViewModel components.NavBarViewModel
	cartCount, err := h.storeServices.CartCount(c.Context(), userID)
	//FIXME
	if err != nil {
		return c.Redirect("/")
	}
	navBarViewModel.Align(cartCount)

	customer, err := h.customerServices.Customer(c.Context(), userID)
	if err != nil {
		//FIXME TOAST
		return c.Redirect("/")
	}

	return renderFragmentOrView(c, views.AccountProfile(views.NewProfileViewModel(customer, navBarViewModel)), views.ProfileFragment)
}

func oauthLogoutHandler(sessionStore *session.Store) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		sess, err := sessionStore.Get(c)
		if err != nil {
			return c.SendString("retrieving session: " + err.Error())
		}
		if err = sess.Destroy(); err != nil {
			return c.SendString("destroying session: " + err.Error())
		}
		logoutUrl, err := url.Parse("https://" + os.Getenv("AUTH0_DOMAIN") + "/v2/logout")
		if err != nil {
			return c.SendString(err.Error())
		}
		returnTo, err := url.Parse(string(c.Request().URI().Scheme()) + string(c.Request().Host()))
		if err != nil {
			return c.SendString(err.Error())
		}
		parameters := url.Values{}
		parameters.Add("returnTo", returnTo.String())
		parameters.Add("client_id", os.Getenv("AUTH0_CLIENT_ID"))
		logoutUrl.RawQuery = parameters.Encode()
		return c.Redirect(logoutUrl.String(), http.StatusTemporaryRedirect)
	}
}

func (h handlers) checkoutStripeWebhook(c *fiber.Ctx) error {
	endpointSecret := os.Getenv("STRIPE_WEBHOOK_SIGNING_SECRET")

	event, err := webhook.ConstructEvent(c.Body(), c.Get("Stripe-Signature"), endpointSecret)
	type errStruct struct {
		Err string `json:"error"`
	}
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(errStruct{Err: fmt.Sprintf("Parsing body: %v", err)})
	}
	if event.Type != stripe.EventTypeCheckoutSessionCompleted && event.Type != stripe.EventTypeCheckoutSessionExpired {
		return c.SendStatus(http.StatusOK)
	}
	var checkoutSession stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &checkoutSession); err != nil {
		fmt.Printf("unmarshal checkout: %v", err)

		return c.Status(http.StatusInternalServerError).JSON(errStruct{Err: fmt.Sprintf("unmarshalling checkout session: %v", err)})
	}

	switch event.Type {
	case stripe.EventTypeCheckoutSessionCompleted:
		err = h.storeServices.CreateOrder(
			c.Context(),
			checkoutSession.ClientReferenceID,
			time.Unix(event.Created, 0).Format(time.RFC3339),
		)
		if err != nil {
			fmt.Printf(" CREAITNGORDER: %v", err)
			return c.Status(http.StatusInternalServerError).JSON(errStruct{Err: fmt.Sprintf("creating order: %v", err)})
		}
	case stripe.EventTypeCheckoutSessionExpired:
		var session stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &session); err != nil {

			return c.Status(http.StatusInternalServerError).JSON(errStruct{Err: fmt.Sprintf("unmarshalling checkout session: %v", err)})
		}
		err := h.storeServices.InvalidateCheckout(c.Context(), session.ClientReferenceID)
		if err != nil {
			fmt.Printf("INVALIDATING EXPIRED CHECKOUT: %v", err)
			return c.Status(http.StatusInternalServerError).JSON(errStruct{Err: fmt.Sprintf("invalidating checkout: %v", err)})
		}
	}

	return c.SendStatus(http.StatusCreated)
}
