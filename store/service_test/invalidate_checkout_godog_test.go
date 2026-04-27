package store_service_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/cucumber/godog"
	"github.com/siderustler/go-ecommerce/store"
	store_domain "github.com/siderustler/go-ecommerce/store/domain"
)

type invalidateCheckoutFeatureState struct {
	services              *store.Services
	checkoutID            string
	checkout              store_domain.Checkout
	stock                 store_domain.Stock
	preStock              store_domain.Stock
	missingStockProductID string
	err                   error
}

func newInvalidateCheckoutFeatureState() *invalidateCheckoutFeatureState {
	s := &invalidateCheckoutFeatureState{}
	s.reset()
	return s
}

func (s *invalidateCheckoutFeatureState) reset() {
	s.checkoutID = ""
	s.checkout = zeroCheckout()
	s.stock = emptyStock()
	s.preStock = emptyStock()
	s.missingStockProductID = ""
	s.err = nil
	repo := repositoryMock{
		updateCheckoutFn: func(ctx context.Context, checkoutID string, updateFn func(checkout *store_domain.Checkout, stock *store_domain.Stock) error) error {
			return updateFn(&s.checkout, &s.stock)
		},
	}
	s.services = newServices(repo)
}

func (s *invalidateCheckoutFeatureState) givenAlreadyInvalidated(checkoutID, userID string) error {
	s.checkoutID = checkoutID
	s.checkout = invalidatedCheckout(checkoutID, userID, map[string]store_domain.CartProduct{})
	return nil
}

func (s *invalidateCheckoutFeatureState) givenPendingReserving(checkoutID, userID string, count int, productID string) error {
	s.checkoutID = checkoutID
	s.checkout = pendingCheckout(checkoutID, userID, map[string]store_domain.CartProduct{productID: store_domain.NewCartProduct(productID, count)})
	return nil
}

func (s *invalidateCheckoutFeatureState) givenReservesNoProducts() error {
	s.checkout.Items = map[string]store_domain.CartProduct{}
	return nil
}

func (s *invalidateCheckoutFeatureState) givenStock(productID string, available int, reserved int) error {
	s.stock.Items[productID] = store_domain.NewStockItem(productID, available, reserved)
	return nil
}

func (s *invalidateCheckoutFeatureState) givenNoStockRecord(productID string) error {
	s.missingStockProductID = productID
	delete(s.stock.Items, productID)
	return nil
}

func (s *invalidateCheckoutFeatureState) whenInvalidates(checkoutID string) error {
	s.checkoutID = checkoutID
	s.preStock = cloneStock(s.stock)
	s.err = s.services.Command.InvalidateCheckout.Handle(context.Background(), store.InvalidateCheckoutCmd{CheckoutID: checkoutID})
	return nil
}

func (s *invalidateCheckoutFeatureState) thenRejects() error {
	if s.err == nil {
		return fmt.Errorf("expected rejection")
	}
	return nil
}

func (s *invalidateCheckoutFeatureState) thenCheckoutInvalidated() error {
	if s.checkout.Status != store_domain.CheckoutInvalidated {
		return fmt.Errorf("expected invalidated checkout")
	}
	return nil
}

func (s *invalidateCheckoutFeatureState) thenStock(productID string, available int, reserved int) error {
	item, ok := s.stock.Items[productID]
	if !ok || item.AvailableAmount != available || item.ReservedAmount != reserved {
		return fmt.Errorf("unexpected stock for %q", productID)
	}
	return nil
}

func (s *invalidateCheckoutFeatureState) thenMissingStockDoesNotBlock() error {
	if s.err != nil {
		return fmt.Errorf("expected success, got %v", s.err)
	}
	if _, ok := s.stock.Items[s.missingStockProductID]; ok {
		return fmt.Errorf("expected stock record for %q to remain absent", s.missingStockProductID)
	}
	return nil
}

func InitializeInvalidateCheckoutScenario(ctx *godog.ScenarioContext) {
	s := newInvalidateCheckoutFeatureState()
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		s.reset()
		return ctx, nil
	})
	ctx.Step(`^checkout "([^"]*)" for customer "([^"]*)" is already invalidated$`, s.givenAlreadyInvalidated)
	ctx.Step(`^checkout "([^"]*)" for customer "([^"]*)" is pending and reserves (\d+) units? of product "([^"]*)"$`, s.givenPendingReserving)
	ctx.Step(`^the checkout reserves no products$`, s.givenReservesNoProducts)
	ctx.Step(`^stock for product "([^"]*)" has (\d+) available units? and (\d+) reserved units?$`, s.givenStock)
	ctx.Step(`^there is no stock record for product "([^"]*)"$`, s.givenNoStockRecord)
	ctx.Step(`^the store invalidates checkout "([^"]*)"$`, s.whenInvalidates)
	ctx.Step(`^the store rejects the request$`, s.thenRejects)
	ctx.Step(`^the checkout becomes invalidated$`, s.thenCheckoutInvalidated)
	ctx.Step(`^stock for product "([^"]*)" has (\d+) available units? and (\d+) reserved units?$`, s.thenStock)
	ctx.Step(`^the missing stock record does not block invalidation$`, s.thenMissingStockDoesNotBlock)
}

func TestInvalidateCheckoutFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeInvalidateCheckoutScenario,
		Options:             &godog.Options{Format: "pretty", Paths: []string{"features/invalidate_checkout.feature"}, TestingT: t},
	}
	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run invalidate checkout feature tests")
	}
}
