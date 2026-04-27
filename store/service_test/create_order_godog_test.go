package store_service_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/cucumber/godog"
	"github.com/siderustler/go-ecommerce/store"
	store_domain "github.com/siderustler/go-ecommerce/store/domain"
)

type createOrderFeatureState struct {
	services   *store.Services
	userID     string
	checkoutID string
	cart       store_domain.Cart
	checkout   store_domain.Checkout
	stock      store_domain.Stock
	products   []store_domain.Product
	order      store_domain.Order
	err        error
}

func newCreateOrderFeatureState() *createOrderFeatureState {
	s := &createOrderFeatureState{}
	s.reset()
	return s
}

func (s *createOrderFeatureState) reset() {
	s.userID = ""
	s.checkoutID = ""
	s.cart = zeroCart()
	s.checkout = zeroCheckout()
	s.stock = emptyStock()
	s.products = nil
	s.order = store_domain.Order{}
	s.err = nil
	repo := repositoryMock{
		createOrderFn: func(ctx context.Context, checkoutID string, createFn func(cart *store_domain.Cart, checkout *store_domain.Checkout, stock *store_domain.Stock, products []store_domain.Product) (store_domain.Order, error)) error {
			order, err := createFn(&s.cart, &s.checkout, &s.stock, s.products)
			if err == nil {
				s.order = order
			}
			return err
		},
	}
	s.services = newServices(repo)
}

func (s *createOrderFeatureState) givenInactiveCart(userID string) error {
	s.userID = userID
	s.cart = inactiveCart(userID, map[string]store_domain.CartProduct{})
	return nil
}

func (s *createOrderFeatureState) givenActiveCartContaining(userID string, count int, productID string) error {
	s.userID = userID
	s.cart = activeCart(userID, map[string]store_domain.CartProduct{productID: store_domain.NewCartProduct(productID, count)})
	return nil
}

func (s *createOrderFeatureState) givenPendingCheckout(checkoutID, userID string) error {
	s.userID = userID
	s.checkoutID = checkoutID
	s.checkout = pendingCheckout(checkoutID, userID, map[string]store_domain.CartProduct{})
	return nil
}

func (s *createOrderFeatureState) givenCheckoutContainsNoProducts() error {
	s.checkout.Items = map[string]store_domain.CartProduct{}
	return nil
}

func (s *createOrderFeatureState) givenPendingCheckoutReserving(userID string, checkoutID string, count int, productID string) error {
	s.userID = userID
	s.checkoutID = checkoutID
	s.checkout = pendingCheckout(checkoutID, userID, map[string]store_domain.CartProduct{productID: store_domain.NewCartProduct(productID, count)})
	return nil
}

func (s *createOrderFeatureState) givenStock(productID string, available int, reserved int) error {
	s.stock.Items[productID] = store_domain.NewStockItem(productID, available, reserved)
	return nil
}

func (s *createOrderFeatureState) givenProductPrices(productID string, actual float32, discount float32) error {
	s.products = []store_domain.Product{store_domain.NewProduct(productID, productID, actual, discount)}
	return nil
}

func (s *createOrderFeatureState) whenCreatesOrder(checkoutID string) error {
	s.checkoutID = checkoutID
	s.err = s.services.Command.CreateOrder.Handle(context.Background(), store.CreateOrderCmd{CheckoutID: checkoutID, OrderTime: "now"})
	return nil
}

func (s *createOrderFeatureState) thenRejects() error {
	if s.err == nil {
		return fmt.Errorf("expected rejection")
	}
	return nil
}

func (s *createOrderFeatureState) thenCartInactive() error {
	if s.cart.Status != store_domain.CartInactive {
		return fmt.Errorf("expected inactive cart")
	}
	return nil
}

func (s *createOrderFeatureState) thenCheckoutFinalized() error {
	if s.checkout.Status != store_domain.CheckoutFinalized {
		return fmt.Errorf("expected finalized checkout")
	}
	return nil
}

func (s *createOrderFeatureState) thenStock(productID string, available int, reserved int) error {
	item, ok := s.stock.Items[productID]
	if !ok || item.AvailableAmount != available || item.ReservedAmount != reserved {
		return fmt.Errorf("unexpected stock for %q", productID)
	}
	return nil
}

func (s *createOrderFeatureState) thenOrderContainsLines(count int, productID string) error {
	if s.err != nil {
		return fmt.Errorf("expected created order, got %v", s.err)
	}
	if s.order.ID == "" || s.order.CheckoutID != s.checkoutID {
		return fmt.Errorf("expected order to be created for checkout %q", s.checkoutID)
	}
	if len(s.order.Products) != count || s.order.Products[0].Name != productID {
		return fmt.Errorf("expected %d order line(s) for %q", count, productID)
	}
	return nil
}

func (s *createOrderFeatureState) thenDiscountedPrice(productID string, price float32) error {
	for _, product := range s.order.Products {
		if product.Name == productID {
			if product.ItemPrice != price {
				return fmt.Errorf("expected discounted price %.2f for %q", price, productID)
			}
			return nil
		}
	}
	return fmt.Errorf("expected order line for %q", productID)
}

func InitializeCreateOrderScenario(ctx *godog.ScenarioContext) {
	s := newCreateOrderFeatureState()
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		s.reset()
		return ctx, nil
	})
	ctx.Step(`^customer "([^"]*)" has an inactive cart$`, s.givenInactiveCart)
	ctx.Step(`^customer "([^"]*)" has an active cart containing (\d+) units? of product "([^"]*)"$`, s.givenActiveCartContaining)
	ctx.Step(`^checkout "([^"]*)" for customer "([^"]*)" is pending$`, s.givenPendingCheckout)
	ctx.Step(`^the checkout contains no products$`, s.givenCheckoutContainsNoProducts)
	ctx.Step(`^customer "([^"]*)" has a pending checkout "([^"]*)" which reserves (\d+) units? of product "([^"]*)"$`, s.givenPendingCheckoutReserving)
	ctx.Step(`^customer "([^"]*)" has pending checkout "([^"]*)" which reserves (\d+) units? of product "([^"]*)"$`, s.givenPendingCheckoutReserving)
	ctx.Step(`^stock for product "([^"]*)" has (\d+) available units? and (\d+) reserved units?$`, s.givenStock)
	ctx.Step(`^product "([^"]*)" has regular price ([\d.]+) and discount price ([\d.]+)$`, s.givenProductPrices)
	ctx.Step(`^the store creates an order from checkout "([^"]*)"$`, s.whenCreatesOrder)
	ctx.Step(`^the store rejects the request$`, s.thenRejects)
	ctx.Step(`^the cart becomes inactive$`, s.thenCartInactive)
	ctx.Step(`^the checkout becomes finalized$`, s.thenCheckoutFinalized)
	ctx.Step(`^stock for product "([^"]*)" has (\d+) available units? and (\d+) reserved units?$`, s.thenStock)
	ctx.Step(`^the order contains (\d+) line for product "([^"]*)"$`, s.thenOrderContainsLines)
	ctx.Step(`^the order line for product "([^"]*)" uses discounted price ([\d.]+)$`, s.thenDiscountedPrice)
}

func TestCreateOrderFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeCreateOrderScenario,
		Options:             &godog.Options{Format: "pretty", Paths: []string{"features/create_order.feature"}, TestingT: t},
	}
	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run create order feature tests")
	}
}
