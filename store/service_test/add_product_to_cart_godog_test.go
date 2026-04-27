package store_service_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/cucumber/godog"
	"github.com/siderustler/go-ecommerce/store"
	store_domain "github.com/siderustler/go-ecommerce/store/domain"
)

type addProductToCartFeatureState struct {
	services              *store.Services
	userID                string
	cart                  store_domain.Cart
	checkout              store_domain.Checkout
	stock                 store_domain.Stock
	item                  store_domain.CartProduct
	preCart               store_domain.Cart
	preStock              store_domain.Stock
	missingStockProductID string
	err                   error
}

func newAddProductToCartFeatureState() *addProductToCartFeatureState {
	s := &addProductToCartFeatureState{}
	s.reset()
	return s
}

func (s *addProductToCartFeatureState) reset() {
	s.userID = ""
	s.cart = zeroCart()
	s.checkout = zeroCheckout()
	s.stock = emptyStock()
	s.item = store_domain.CartProduct{}
	s.preCart = zeroCart()
	s.preStock = emptyStock()
	s.missingStockProductID = ""
	s.err = nil
	repo := repositoryMock{
		upsertCartFn: func(ctx context.Context, userID string, item store_domain.CartProduct, upsertFn func(*store_domain.Cart, *store_domain.Checkout, *store_domain.Stock, store_domain.StockItem) error) error {
			stockItem := store_domain.StockItem{}
			if existing, ok := s.stock.Items[item.ProductID]; ok {
				stockItem = existing
			}
			return upsertFn(&s.cart, &s.checkout, &s.stock, stockItem)
		},
	}
	s.services = newServices(repo)
}

func (s *addProductToCartFeatureState) givenActiveCartContainingNoProducts(userID string) error {
	s.userID = userID
	s.cart = activeCart(userID, map[string]store_domain.CartProduct{})
	return nil
}

func (s *addProductToCartFeatureState) givenActiveCart(userID string) error {
	s.userID = userID
	s.cart = activeCart(userID, map[string]store_domain.CartProduct{})
	return nil
}

func (s *addProductToCartFeatureState) givenDidntAddAnyProductYet(userID string) error {
	s.userID = userID
	s.cart = zeroCart()
	return nil
}

func (s *addProductToCartFeatureState) givenPendingCheckoutReserving(userID string, count int, productID string) error {
	s.userID = userID
	s.checkout = pendingCheckout("checkout-"+userID, userID, map[string]store_domain.CartProduct{
		productID: store_domain.NewCartProduct(productID, count),
	})
	return nil
}

func (s *addProductToCartFeatureState) givenPendingCheckoutContainingProduct(userID string, productID string) error {
	s.userID = userID
	s.checkout = pendingCheckout("checkout-"+userID, userID, map[string]store_domain.CartProduct{
		productID: store_domain.NewCartProduct(productID, 1),
	})
	return nil
}

func (s *addProductToCartFeatureState) givenProductNotAvailable(productID string) error {
	s.stock.Items[productID] = store_domain.NewStockItem(productID, 0, 0)
	return nil
}

func (s *addProductToCartFeatureState) givenProductHasAvailableStock(productID string, available int) error {
	s.stock.Items[productID] = store_domain.NewStockItem(productID, available, 0)
	return nil
}

func (s *addProductToCartFeatureState) givenStock(productID string, available int, reserved int) error {
	s.stock.Items[productID] = store_domain.NewStockItem(productID, available, reserved)
	return nil
}

func (s *addProductToCartFeatureState) givenNoStockRecord(productID string) error {
	s.missingStockProductID = productID
	delete(s.stock.Items, productID)
	return nil
}

func (s *addProductToCartFeatureState) whenAdds(userID string, count int, productID string) error {
	s.userID = userID
	s.preCart = cloneCart(s.cart)
	s.preStock = cloneStock(s.stock)
	s.item = store_domain.NewCartProduct(productID, count)
	s.err = s.services.Command.AddProductToCart.Handle(context.Background(), store.AddProductToCartCmd{
		UserID:       userID,
		ProductToAdd: s.item,
	})
	return nil
}

func (s *addProductToCartFeatureState) whenTriesToAdd(userID string, count int, productID string) error {
	return s.whenAdds(userID, count, productID)
}

func (s *addProductToCartFeatureState) thenRejects() error {
	if s.err == nil {
		return fmt.Errorf("expected rejection")
	}
	return nil
}

func (s *addProductToCartFeatureState) thenCartUnchanged(userID string) error {
	if !reflect.DeepEqual(cloneCart(s.cart), s.preCart) {
		return fmt.Errorf("expected cart for %q unchanged", userID)
	}
	return nil
}

func (s *addProductToCartFeatureState) thenCreatesNewActiveCart(userID string) error {
	if s.err != nil || s.cart.IsZero() || s.cart.CustomerID != userID || s.cart.Status != store_domain.CartActive {
		return fmt.Errorf("expected new active cart for %q", userID)
	}
	return nil
}

func (s *addProductToCartFeatureState) thenCartContains(count int, productID string) error {
	product, ok := s.cart.Products[productID]
	if !ok || product.Count != count {
		return fmt.Errorf("expected cart to contain %d of %q", count, productID)
	}
	return nil
}

func (s *addProductToCartFeatureState) thenInvalidatesPendingCheckout() error {
	if s.checkout.Status != store_domain.CheckoutInvalidated {
		return fmt.Errorf("expected checkout invalidated")
	}
	return nil
}

func (s *addProductToCartFeatureState) thenReleasesReservation(productID string) error {
	item, ok := s.stock.Items[productID]
	if !ok || item.ReservedAmount != 0 {
		return fmt.Errorf("expected reservation released for %q", productID)
	}
	return nil
}

func (s *addProductToCartFeatureState) thenStock(productID string, available int, reserved int) error {
	item, ok := s.stock.Items[productID]
	if !ok || item.AvailableAmount != available || item.ReservedAmount != reserved {
		return fmt.Errorf("unexpected stock for %q", productID)
	}
	return nil
}

func (s *addProductToCartFeatureState) thenSkipsMissingStockWithoutFailing() error {
	if s.err != nil {
		return fmt.Errorf("expected success, got %v", s.err)
	}
	if _, ok := s.stock.Items[s.missingStockProductID]; ok {
		return fmt.Errorf("expected missing stock record for %q to remain absent", s.missingStockProductID)
	}
	if !reflect.DeepEqual(s.preStock.Items[s.item.ProductID], s.stock.Items[s.item.ProductID]) {
		// The product being added should not mutate stock amounts in add-to-cart flow.
		return fmt.Errorf("expected stock amounts for %q to remain unchanged", s.item.ProductID)
	}
	if _, existedBefore := s.preStock.Items[s.missingStockProductID]; existedBefore {
		return fmt.Errorf("expected precondition to keep %q absent from stock", s.missingStockProductID)
	}
	return nil
}

func InitializeAddProductToCartScenario(ctx *godog.ScenarioContext) {
	s := newAddProductToCartFeatureState()
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		s.reset()
		return ctx, nil
	})
	ctx.Step(`^customer "([^"]*)" has an active cart containing no products$`, s.givenActiveCartContainingNoProducts)
	ctx.Step(`^customer "([^"]*)" has an active cart$`, s.givenActiveCart)
	ctx.Step(`^customer "([^"]*)" didnt add any product yet$`, s.givenDidntAddAnyProductYet)
	ctx.Step(`^customer "([^"]*)" has a pending checkout reserving (\d+) units? of product "([^"]*)"$`, s.givenPendingCheckoutReserving)
	ctx.Step(`^customer "([^"]*)" has a pending checkout containing product "([^"]*)"$`, s.givenPendingCheckoutContainingProduct)
	ctx.Step(`^product "([^"]*)" is not available in stock$`, s.givenProductNotAvailable)
	ctx.Step(`^product "([^"]*)" has (\d+) units available in stock$`, s.givenProductHasAvailableStock)
	ctx.Step(`^stock for product "([^"]*)" has (\d+) available units? and (\d+) reserved units?$`, s.givenStock)
	ctx.Step(`^there is no stock record for product "([^"]*)"$`, s.givenNoStockRecord)
	ctx.Step(`^customer "([^"]*)" adds (\d+) units? of product "([^"]*)" to the cart$`, s.whenAdds)
	ctx.Step(`^customer "([^"]*)" tries to add (\d+) units? of product "([^"]*)" to the cart$`, s.whenTriesToAdd)
	ctx.Step(`^the store rejects the request$`, s.thenRejects)
	ctx.Step(`^customer "([^"]*)"'s cart remains unchanged$`, s.thenCartUnchanged)
	ctx.Step(`^the store creates a new active cart for customer "([^"]*)"$`, s.thenCreatesNewActiveCart)
	ctx.Step(`^the cart contains (\d+) units? of product "([^"]*)"$`, s.thenCartContains)
	ctx.Step(`^the store invalidates the existing pending checkout$`, s.thenInvalidatesPendingCheckout)
	ctx.Step(`^the store invalidates the pending checkout$`, s.thenInvalidatesPendingCheckout)
	ctx.Step(`^the store releases the reservation for product "([^"]*)"$`, s.thenReleasesReservation)
	ctx.Step(`^stock for product "([^"]*)" has (\d+) available units? and (\d+) reserved units?$`, s.thenStock)
	ctx.Step(`^the store skips the missing stock record without failing the request$`, s.thenSkipsMissingStockWithoutFailing)
}

func TestAddProductToCartFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeAddProductToCartScenario,
		Options:             &godog.Options{Format: "pretty", Paths: []string{"features/add_product_to_cart.feature"}, TestingT: t},
	}
	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run add product to cart feature tests")
	}
}
