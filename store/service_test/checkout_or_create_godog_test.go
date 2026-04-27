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

type checkoutOrCreateFeatureState struct {
	services               *store.Services
	userID                 string
	cart                   store_domain.Cart
	checkout               store_domain.Checkout
	stock                  store_domain.Stock
	preCheckout            store_domain.Checkout
	preStock               store_domain.Stock
	returnExistingCheckout bool
	err                    error
}

func newCheckoutOrCreateFeatureState() *checkoutOrCreateFeatureState {
	s := &checkoutOrCreateFeatureState{}
	s.reset()
	return s
}

func (s *checkoutOrCreateFeatureState) reset() {
	s.userID = ""
	s.cart = zeroCart()
	s.checkout = zeroCheckout()
	s.stock = emptyStock()
	s.preCheckout = zeroCheckout()
	s.preStock = emptyStock()
	s.returnExistingCheckout = false
	s.err = nil
	repo := repositoryMock{
		checkoutOrCreateFn: func(ctx context.Context, userID string, insertFn func(cart *store_domain.Cart, stock *store_domain.Stock) (store_domain.Checkout, error)) (store_domain.Checkout, error) {
			if s.returnExistingCheckout {
				return s.checkout, nil
			}
			return insertFn(&s.cart, &s.stock)
		},
	}
	s.services = newServices(repo)
}

func (s *checkoutOrCreateFeatureState) givenExistingPendingCheckout(userID, productID string) error {
	s.userID = userID
	s.returnExistingCheckout = true
	s.checkout = pendingCheckout("checkout-"+userID, userID, map[string]store_domain.CartProduct{productID: store_domain.NewCartProduct(productID, 1)})
	return nil
}

func (s *checkoutOrCreateFeatureState) givenNoCart(userID string) error {
	s.userID = userID
	s.cart = zeroCart()
	return nil
}

func (s *checkoutOrCreateFeatureState) givenCartContaining(userID string, count int, productID string) error {
	s.userID = userID
	s.cart = activeCart(userID, map[string]store_domain.CartProduct{productID: store_domain.NewCartProduct(productID, count)})
	return nil
}

func (s *checkoutOrCreateFeatureState) givenNoReservedStockForCart() error { return nil }

func (s *checkoutOrCreateFeatureState) givenNoStockRecord(productID string) error {
	delete(s.stock.Items, productID)
	return nil
}

func (s *checkoutOrCreateFeatureState) givenStock(productID string, available int, reserved int) error {
	s.stock.Items[productID] = store_domain.NewStockItem(productID, available, reserved)
	return nil
}

func (s *checkoutOrCreateFeatureState) whenStartsCheckout(userID string) error {
	s.userID = userID
	s.preCheckout = cloneCheckout(s.checkout)
	s.preStock = cloneStock(s.stock)
	checkout, err := s.services.Command.CheckoutOrCreate.Handle(context.Background(), store.CheckoutOrCreateCmd{UserID: userID})
	s.err = err
	if err == nil {
		s.checkout = checkout
	}
	return nil
}

func (s *checkoutOrCreateFeatureState) thenRejects() error {
	if s.err == nil {
		return fmt.Errorf("expected rejection")
	}
	return nil
}

func (s *checkoutOrCreateFeatureState) thenReturnsExistingCheckout() error {
	if s.err != nil || s.checkout.Status != store_domain.CheckoutPending {
		return fmt.Errorf("expected existing pending checkout")
	}
	if !reflect.DeepEqual(cloneCheckout(s.checkout), s.preCheckout) {
		return fmt.Errorf("expected existing checkout unchanged")
	}
	if !reflect.DeepEqual(cloneStock(s.stock), s.preStock) {
		return fmt.Errorf("expected stock unchanged when returning existing checkout")
	}
	return nil
}

func (s *checkoutOrCreateFeatureState) thenCreatesPendingCheckout(userID string) error {
	if s.err != nil || s.checkout.IsZero() || s.checkout.UserID != userID || s.checkout.Status != store_domain.CheckoutPending {
		return fmt.Errorf("expected pending checkout for %q", userID)
	}
	return nil
}

func (s *checkoutOrCreateFeatureState) thenCheckoutContains(count int, productID string) error {
	item, ok := s.checkout.Items[productID]
	if !ok || item.Count != count {
		return fmt.Errorf("expected checkout to contain %d of %q", count, productID)
	}
	return nil
}

func (s *checkoutOrCreateFeatureState) thenStock(productID string, available int, reserved int) error {
	item, ok := s.stock.Items[productID]
	if !ok || item.AvailableAmount != available || item.ReservedAmount != reserved {
		return fmt.Errorf("unexpected stock for %q", productID)
	}
	return nil
}

func InitializeCheckoutOrCreateScenario(ctx *godog.ScenarioContext) {
	s := newCheckoutOrCreateFeatureState()
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		s.reset()
		return ctx, nil
	})
	ctx.Step(`^customer "([^"]*)" already has a pending checkout for product "([^"]*)"$`, s.givenExistingPendingCheckout)
	ctx.Step(`^customer "([^"]*)" does not have a cart yet$`, s.givenNoCart)
	ctx.Step(`^customer "([^"]*)" has a cart containing (\d+) units? of product "([^"]*)"$`, s.givenCartContaining)
	ctx.Step(`^there is no reserved stock for the cart$`, s.givenNoReservedStockForCart)
	ctx.Step(`^there is no stock record for product "([^"]*)"$`, s.givenNoStockRecord)
	ctx.Step(`^stock for product "([^"]*)" has (\d+) available units? and (\d+) reserved units?$`, s.givenStock)
	ctx.Step(`^customer "([^"]*)" starts checkout$`, s.whenStartsCheckout)
	ctx.Step(`^the store rejects the request$`, s.thenRejects)
	ctx.Step(`^the store returns the existing pending checkout unchanged$`, s.thenReturnsExistingCheckout)
	ctx.Step(`^the store creates a pending checkout for customer "([^"]*)"$`, s.thenCreatesPendingCheckout)
	ctx.Step(`^the checkout contains (\d+) units? of product "([^"]*)"$`, s.thenCheckoutContains)
	ctx.Step(`^stock for product "([^"]*)" has (\d+) available units? and (\d+) reserved units?$`, s.thenStock)
}

func TestCheckoutOrCreateFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeCheckoutOrCreateScenario,
		Options:             &godog.Options{Format: "pretty", Paths: []string{"features/checkout_or_create.feature"}, TestingT: t},
	}
	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run checkout or create feature tests")
	}
}
