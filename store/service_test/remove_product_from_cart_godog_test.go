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

type removeProductFromCartFeatureState struct {
	services *store.Services
	userID   string
	cart     store_domain.Cart
	checkout store_domain.Checkout
	stock    store_domain.Stock
	item     store_domain.CartProduct
	preCart  store_domain.Cart
	preStock store_domain.Stock
	err      error
}

func newRemoveProductFromCartFeatureState() *removeProductFromCartFeatureState {
	s := &removeProductFromCartFeatureState{}
	s.reset()
	return s
}

func (s *removeProductFromCartFeatureState) reset() {
	s.userID = ""
	s.cart = zeroCart()
	s.checkout = zeroCheckout()
	s.stock = emptyStock()
	s.item = store_domain.CartProduct{}
	s.preCart = zeroCart()
	s.preStock = emptyStock()
	s.err = nil
	repo := repositoryMock{
		upsertCartFn: func(ctx context.Context, userID string, item store_domain.CartProduct, upsertFn func(*store_domain.Cart, *store_domain.Checkout, *store_domain.Stock, store_domain.StockItem) error) error {
			return upsertFn(&s.cart, &s.checkout, &s.stock, store_domain.StockItem{})
		},
	}
	s.services = newServices(repo)
}

func (s *removeProductFromCartFeatureState) givenActiveCartWithUnits(userID string, count int, productID string) error {
	s.userID = userID
	s.cart = activeCart(userID, map[string]store_domain.CartProduct{productID: store_domain.NewCartProduct(productID, count)})
	return nil
}

func (s *removeProductFromCartFeatureState) givenActiveCartWithNoUnits(userID, productID string) error {
	s.userID = userID
	s.cart = activeCart(userID, map[string]store_domain.CartProduct{})
	return nil
}

func (s *removeProductFromCartFeatureState) givenNoCheckout(userID string) error {
	s.userID = userID
	s.checkout = zeroCheckout()
	return nil
}

func (s *removeProductFromCartFeatureState) givenPendingCheckoutReserving(userID string, count int, productID string) error {
	s.userID = userID
	s.checkout = pendingCheckout("checkout-"+userID, userID, map[string]store_domain.CartProduct{productID: store_domain.NewCartProduct(productID, count)})
	return nil
}

func (s *removeProductFromCartFeatureState) givenStock(productID string, available int, reserved int) error {
	s.stock.Items[productID] = store_domain.NewStockItem(productID, available, reserved)
	return nil
}

func (s *removeProductFromCartFeatureState) givenNoStockRecord(productID string) error {
	delete(s.stock.Items, productID)
	return nil
}

func (s *removeProductFromCartFeatureState) whenRemoves(userID string, count int, productID string) error {
	s.userID = userID
	s.preCart = cloneCart(s.cart)
	s.preStock = cloneStock(s.stock)
	s.item = store_domain.NewCartProduct(productID, count)
	s.err = s.services.Command.RemoveProductFromCart.Handle(context.Background(), store.RemoveProductFromCartCmd{UserID: userID, CartProduct: s.item})
	return nil
}

func (s *removeProductFromCartFeatureState) whenTriesToRemove(userID string, count int, productID string) error {
	return s.whenRemoves(userID, count, productID)
}

func (s *removeProductFromCartFeatureState) thenRejects() error {
	if s.err == nil {
		return fmt.Errorf("expected rejection")
	}
	return nil
}

func (s *removeProductFromCartFeatureState) thenCartContains(count int, productID string) error {
	product, ok := s.cart.Products[productID]
	if !ok || product.Count != count {
		return fmt.Errorf("expected cart to contain %d of %q", count, productID)
	}
	return nil
}

func (s *removeProductFromCartFeatureState) thenProductRemoved(productID string) error {
	if _, ok := s.cart.Products[productID]; ok {
		return fmt.Errorf("expected %q removed", productID)
	}
	return nil
}

func (s *removeProductFromCartFeatureState) thenKeepsAppliedQuantityChange() error {
	product := s.preCart.Products[s.item.ProductID]
	expected := product.Count - s.item.Count
	entirelyRemoved := expected <= 0
	if entirelyRemoved {
		if _, ok := s.cart.Products[s.item.ProductID]; ok {
			return fmt.Errorf("expected product removed")
		}
		return nil
	}
	current, ok := s.cart.Products[s.item.ProductID]
	if !ok || current.Count != expected {
		return fmt.Errorf("expected quantity change kept")
	}
	return nil
}

func (s *removeProductFromCartFeatureState) thenInvalidatesCheckout() error {
	if s.err != nil {
		return fmt.Errorf("expected successful cart update, got %v", s.err)
	}
	if s.checkout.Status != store_domain.CheckoutInvalidated {
		return fmt.Errorf("expected invalidated checkout")
	}
	return nil
}

func (s *removeProductFromCartFeatureState) thenReservationReleased(productID string) error {
	item, ok := s.stock.Items[productID]
	if !ok || item.ReservedAmount != 0 {
		return fmt.Errorf("expected reservation released for %q", productID)
	}
	return nil
}

func (s *removeProductFromCartFeatureState) thenMissingStockDoesNotBlock() error {
	if s.err != nil {
		return fmt.Errorf("expected success, got %v", s.err)
	}
	if !reflect.DeepEqual(s.preStock.Items[s.item.ProductID], s.stock.Items[s.item.ProductID]) {
		return fmt.Errorf("expected stock for %q unchanged while invalidating checkout", s.item.ProductID)
	}
	return nil
}

func InitializeRemoveProductFromCartScenario(ctx *godog.ScenarioContext) {
	s := newRemoveProductFromCartFeatureState()
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		s.reset()
		return ctx, nil
	})
	ctx.Step(`^customer "([^"]*)" has an active cart with (\d+) units? of product "([^"]*)"$`, s.givenActiveCartWithUnits)
	ctx.Step(`^customer "([^"]*)" has an active cart with no units of product "([^"]*)"$`, s.givenActiveCartWithNoUnits)
	ctx.Step(`^customer "([^"]*)" has no checkout yet$`, s.givenNoCheckout)
	ctx.Step(`^customer "([^"]*)" has a pending checkout reserving (\d+) units? of product "([^"]*)"$`, s.givenPendingCheckoutReserving)
	ctx.Step(`^stock for product "([^"]*)" has (\d+) available units? and (\d+) reserved units?$`, s.givenStock)
	ctx.Step(`^there is no stock record for product "([^"]*)"$`, s.givenNoStockRecord)
	ctx.Step(`^customer "([^"]*)" removes (\d+) units? of product "([^"]*)" from the cart$`, s.whenRemoves)
	ctx.Step(`^customer "([^"]*)" tries to remove (\d+) units? of product "([^"]*)" from the cart$`, s.whenTriesToRemove)
	ctx.Step(`^the store rejects the request$`, s.thenRejects)
	ctx.Step(`^the cart contains (\d+) units? of product "([^"]*)"$`, s.thenCartContains)
	ctx.Step(`^product "([^"]*)" is removed from the cart$`, s.thenProductRemoved)
	ctx.Step(`^the cart still contains (\d+) units? of product "([^"]*)"$`, s.thenCartContains)
	ctx.Step(`^the cart keeps the already applied quantity change$`, s.thenKeepsAppliedQuantityChange)
	ctx.Step(`^the store invalidates the pending checkout$`, s.thenInvalidatesCheckout)
	ctx.Step(`^the reservation for product "([^"]*)" is released$`, s.thenReservationReleased)
	ctx.Step(`^the missing stock record does not block the request$`, s.thenMissingStockDoesNotBlock)
}

func TestRemoveProductFromCartFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeRemoveProductFromCartScenario,
		Options:             &godog.Options{Format: "pretty", Paths: []string{"features/remove_product_from_cart.feature"}, TestingT: t},
	}
	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run remove product from cart feature tests")
	}
}
