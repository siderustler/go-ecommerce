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

type mergeUserCartsFeatureState struct {
	services     *store.Services
	fromUserID   string
	toUserID     string
	fromCart     store_domain.Cart
	toCart       store_domain.Cart
	fromCheckout store_domain.Checkout
	toCheckout   store_domain.Checkout
	stock        store_domain.Stock
	preFromCart  store_domain.Cart
	preToCart    store_domain.Cart
	preStock     store_domain.Stock
	err          error
}

func newMergeUserCartsFeatureState() *mergeUserCartsFeatureState {
	s := &mergeUserCartsFeatureState{}
	s.reset()
	return s
}

func (s *mergeUserCartsFeatureState) reset() {
	s.fromUserID = ""
	s.toUserID = ""
	s.fromCart = zeroCart()
	s.toCart = zeroCart()
	s.fromCheckout = zeroCheckout()
	s.toCheckout = zeroCheckout()
	s.stock = emptyStock()
	s.preFromCart = zeroCart()
	s.preToCart = zeroCart()
	s.preStock = emptyStock()
	s.err = nil
	repo := repositoryMock{
		mergeUserCartsFn: func(ctx context.Context, fromUserID string, toUserID string, mergeFn func(*store_domain.Cart, *store_domain.Cart, *store_domain.Checkout, *store_domain.Checkout, *store_domain.Stock) error) error {
			return mergeFn(&s.fromCart, &s.toCart, &s.fromCheckout, &s.toCheckout, &s.stock)
		},
	}
	s.services = newServices(repo)
}

func (s *mergeUserCartsFeatureState) givenGuestNoActiveCart(userID string) error {
	s.fromUserID = userID
	s.fromCart = zeroCart()
	return nil
}

func (s *mergeUserCartsFeatureState) givenSignedInEmptyCart(userID string) error {
	s.toUserID = userID
	s.toCart = activeCart(userID, map[string]store_domain.CartProduct{})
	return nil
}

func (s *mergeUserCartsFeatureState) givenGuestActiveCartContaining(userID string, count int, productID string) error {
	s.fromUserID = userID
	s.fromCart = activeCart(userID, map[string]store_domain.CartProduct{productID: store_domain.NewCartProduct(productID, count)})
	return nil
}

func (s *mergeUserCartsFeatureState) givenSignedInActiveCart(userID string) error {
	s.toUserID = userID
	s.toCart = activeCart(userID, map[string]store_domain.CartProduct{})
	return nil
}

func (s *mergeUserCartsFeatureState) givenSignedInActiveCartContaining(userID string, count int, productID string) error {
	s.toUserID = userID
	s.toCart = activeCart(userID, map[string]store_domain.CartProduct{productID: store_domain.NewCartProduct(productID, count)})
	return nil
}

func (s *mergeUserCartsFeatureState) givenSignedInNoCartYet(userID string) error {
	s.toUserID = userID
	s.toCart = zeroCart()
	return nil
}

func (s *mergeUserCartsFeatureState) givenNeitherCheckout() error {
	s.fromCheckout = zeroCheckout()
	s.toCheckout = zeroCheckout()
	return nil
}

func (s *mergeUserCartsFeatureState) givenGuestPendingCheckoutReserving(userID string, count int, productID string) error {
	s.fromUserID = userID
	s.fromCheckout = pendingCheckout("checkout-"+userID, userID, map[string]store_domain.CartProduct{productID: store_domain.NewCartProduct(productID, count)})
	return nil
}

func (s *mergeUserCartsFeatureState) givenSignedInPendingCheckoutReserving(userID string, count int, productID string) error {
	s.toUserID = userID
	s.toCheckout = pendingCheckout("checkout-"+userID, userID, map[string]store_domain.CartProduct{productID: store_domain.NewCartProduct(productID, count)})
	return nil
}

func (s *mergeUserCartsFeatureState) givenSignedInPendingCheckoutNoReserved(userID string) error {
	s.toUserID = userID
	s.toCheckout = pendingCheckout("checkout-"+userID, userID, map[string]store_domain.CartProduct{})
	return nil
}

func (s *mergeUserCartsFeatureState) givenGuestPendingCheckoutNoReserved(userID string) error {
	s.fromUserID = userID
	s.fromCheckout = pendingCheckout("checkout-"+userID, userID, map[string]store_domain.CartProduct{})
	return nil
}

func (s *mergeUserCartsFeatureState) givenStock(productID string, available int, reserved int) error {
	s.stock.Items[productID] = store_domain.NewStockItem(productID, available, reserved)
	return nil
}

func (s *mergeUserCartsFeatureState) whenMerges(fromUserID string, toUserID string) error {
	s.fromUserID = fromUserID
	s.toUserID = toUserID
	s.preFromCart = cloneCart(s.fromCart)
	s.preToCart = cloneCart(s.toCart)
	s.preStock = cloneStock(s.stock)
	s.err = s.services.Command.MergeUserCarts.Handle(context.Background(), store.MergeUserCartsCmd{FromID: fromUserID, ToID: toUserID})
	return nil
}

func (s *mergeUserCartsFeatureState) thenRejects() error {
	if s.err == nil {
		return fmt.Errorf("expected rejection")
	}
	return nil
}

func (s *mergeUserCartsFeatureState) thenCompletesWithoutChanging() error {
	if s.err != nil || !reflect.DeepEqual(cloneCart(s.fromCart), s.preFromCart) || !reflect.DeepEqual(cloneCart(s.toCart), s.preToCart) {
		return fmt.Errorf("expected merge without cart changes")
	}
	return nil
}

func (s *mergeUserCartsFeatureState) thenSignedInHasMergedCount(userID string, count int, productID string) error {
	if s.toCart.CustomerID != userID {
		return fmt.Errorf("expected target cart for %q", userID)
	}
	item, ok := s.toCart.Products[productID]
	if !ok || item.Count != count {
		return fmt.Errorf("expected merged count %d of %q", count, productID)
	}
	return nil
}

func (s *mergeUserCartsFeatureState) thenGuestCartInactive(userID string) error {
	if s.fromCart.CustomerID != userID || s.fromCart.Status != store_domain.CartInactive {
		return fmt.Errorf("expected guest cart inactive")
	}
	return nil
}

func (s *mergeUserCartsFeatureState) thenBothCheckoutsInvalidated() error {
	if s.fromCheckout.Status != store_domain.CheckoutInvalidated || s.toCheckout.Status != store_domain.CheckoutInvalidated {
		return fmt.Errorf("expected both checkouts invalidated")
	}
	return nil
}

func (s *mergeUserCartsFeatureState) thenStock(productID string, available int, reserved int) error {
	item, ok := s.stock.Items[productID]
	if !ok || item.AvailableAmount != available || item.ReservedAmount != reserved {
		return fmt.Errorf("unexpected stock for %q", productID)
	}
	return nil
}

func (s *mergeUserCartsFeatureState) thenCreatesNewActiveCartForSignedIn(userID string) error {
	if s.err != nil || s.toCart.IsZero() || s.toCart.CustomerID != userID || s.toCart.Status != store_domain.CartActive {
		return fmt.Errorf("expected new active cart for signed-in customer %q", userID)
	}
	return nil
}

func (s *mergeUserCartsFeatureState) thenNewCartContains(count int, productID string) error {
	item, ok := s.toCart.Products[productID]
	if !ok || item.Count != count {
		return fmt.Errorf("expected new cart to contain %d of %q", count, productID)
	}
	return nil
}

func (s *mergeUserCartsFeatureState) thenCompletesSuccessfully() error {
	if s.err != nil {
		return fmt.Errorf("expected successful merge, got %v", s.err)
	}
	if s.preFromCart.IsZero() {
		return fmt.Errorf("expected scenario to start with a source cart")
	}
	if s.fromCart.Status != store_domain.CartInactive {
		return fmt.Errorf("expected source cart to be inactive after merge")
	}
	expectedProducts := cloneCart(s.preToCart).Products
	if expectedProducts == nil {
		expectedProducts = map[string]store_domain.CartProduct{}
	}
	for productID, product := range s.preFromCart.Products {
		merged := expectedProducts[productID]
		merged.ProductID = productID
		merged.Count += product.Count
		expectedProducts[productID] = merged
	}
	if !reflect.DeepEqual(s.toCart.Products, expectedProducts) {
		return fmt.Errorf("expected merged cart contents to include products from both carts")
	}
	if !reflect.DeepEqual(cloneStock(s.stock), s.preStock) {
		return fmt.Errorf("expected stock unchanged when neither customer had a checkout")
	}
	return nil
}

func InitializeMergeUserCartsScenario(ctx *godog.ScenarioContext) {
	s := newMergeUserCartsFeatureState()
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		s.reset()
		return ctx, nil
	})
	ctx.Step(`^guest customer "([^"]*)" has no active cart contents$`, s.givenGuestNoActiveCart)
	ctx.Step(`^signed-in customer "([^"]*)" has an empty cart$`, s.givenSignedInEmptyCart)
	ctx.Step(`^guest customer "([^"]*)" has an active cart containing (\d+) units? of product "([^"]*)"$`, s.givenGuestActiveCartContaining)
	ctx.Step(`^signed-in customer "([^"]*)" has an active cart$`, s.givenSignedInActiveCart)
	ctx.Step(`^signed-in customer "([^"]*)" has an active cart containing (\d+) units? of product "([^"]*)"$`, s.givenSignedInActiveCartContaining)
	ctx.Step(`^signed-in customer "([^"]*)" does not have a cart yet$`, s.givenSignedInNoCartYet)
	ctx.Step(`^neither customer has a checkout$`, s.givenNeitherCheckout)
	ctx.Step(`^guest customer "([^"]*)" has a pending checkout reserving (\d+) units? of product "([^"]*)"$`, s.givenGuestPendingCheckoutReserving)
	ctx.Step(`^signed-in customer "([^"]*)" has a pending checkout reserving (\d+) units? of product "([^"]*)"$`, s.givenSignedInPendingCheckoutReserving)
	ctx.Step(`^signed-in customer "([^"]*)" has a pending checkout with no reserved items$`, s.givenSignedInPendingCheckoutNoReserved)
	ctx.Step(`^guest customer "([^"]*)" has a pending checkout with no reserved items$`, s.givenGuestPendingCheckoutNoReserved)
	ctx.Step(`^stock for product "([^"]*)" has (\d+) available units? and (\d+) reserved units?$`, s.givenStock)
	ctx.Step(`^the store merges customer "([^"]*)" into customer "([^"]*)"$`, s.whenMerges)
	ctx.Step(`^the store rejects the request$`, s.thenRejects)
	ctx.Step(`^the store completes the merge without changing either cart$`, s.thenCompletesWithoutChanging)
	ctx.Step(`^signed-in customer "([^"]*)" has (\d+) units? of product "([^"]*)" in the merged cart$`, s.thenSignedInHasMergedCount)
	ctx.Step(`^guest customer "([^"]*)"'s cart becomes inactive$`, s.thenGuestCartInactive)
	ctx.Step(`^both pending checkouts become invalidated$`, s.thenBothCheckoutsInvalidated)
	ctx.Step(`^stock for product "([^"]*)" has (\d+) available units? and (\d+) reserved units?$`, s.thenStock)
	ctx.Step(`^the store creates a new active cart for signed-in customer "([^"]*)"$`, s.thenCreatesNewActiveCartForSignedIn)
	ctx.Step(`^the new cart contains (\d+) units? of product "([^"]*)"$`, s.thenNewCartContains)
	ctx.Step(`^the store completes the merge successfully$`, s.thenCompletesSuccessfully)
}

func TestMergeUserCartsFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeMergeUserCartsScenario,
		Options:             &godog.Options{Format: "pretty", Paths: []string{"features/merge_user_carts.feature"}, TestingT: t},
	}
	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run merge user carts feature tests")
	}
}
