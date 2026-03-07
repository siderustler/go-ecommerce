package store_service_test

import (
	"context"
	"testing"

	store_domain "github.com/siderustler/go-ecommerce/store/domain"
)

func TestMergeUserCarts_DoesNothingWhenFromCartIsZero(t *testing.T) {
	from := store_domain.Cart{Products: map[string]store_domain.CartProduct{}}
	to := store_domain.Cart{Products: map[string]store_domain.CartProduct{}}
	fromCheckout := store_domain.Checkout{Items: map[string]store_domain.CartProduct{}}
	toCheckout := store_domain.Checkout{Items: map[string]store_domain.CartProduct{}}
	stock := store_domain.Stock{Items: map[string]store_domain.StockItem{}}

	err := runMergeUserCarts(&from, &to, &fromCheckout, &toCheckout, &stock)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestMergeUserCarts_ErrorsWhenToCartMergeFails(t *testing.T) {
	from := store_domain.NewCart("from", "u1", map[string]store_domain.CartProduct{"p1": store_domain.NewCartProduct("p1", 1)}, "", store_domain.CartActive)
	to := store_domain.NewCart("to", "u2", map[string]store_domain.CartProduct{}, "", store_domain.CartInactive)
	fromCheckout := store_domain.NewCheckout("fch", "u1", map[string]store_domain.CartProduct{}, "", store_domain.CheckoutPending)
	toCheckout := store_domain.NewCheckout("tch", "u2", map[string]store_domain.CartProduct{}, "", store_domain.CheckoutPending)
	stock := store_domain.Stock{Items: map[string]store_domain.StockItem{}}

	err := runMergeUserCarts(&from, &to, &fromCheckout, &toCheckout, &stock)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestMergeUserCarts_ErrorsWhenInvalidatingFromCheckoutFails(t *testing.T) {
	from := store_domain.NewCart("from", "u1", map[string]store_domain.CartProduct{"p1": store_domain.NewCartProduct("p1", 1)}, "", store_domain.CartActive)
	to := store_domain.NewCart("to", "u2", map[string]store_domain.CartProduct{}, "", store_domain.CartActive)
	fromCheckout := store_domain.NewCheckout("fch", "u1", map[string]store_domain.CartProduct{}, "", store_domain.CheckoutInvalidated)
	toCheckout := store_domain.NewCheckout("tch", "u2", map[string]store_domain.CartProduct{}, "", store_domain.CheckoutPending)
	stock := store_domain.Stock{Items: map[string]store_domain.StockItem{}}

	err := runMergeUserCarts(&from, &to, &fromCheckout, &toCheckout, &stock)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestMergeUserCarts_ErrorsWhenInvalidatingToCheckoutFails(t *testing.T) {
	from := store_domain.NewCart("from", "u1", map[string]store_domain.CartProduct{"p1": store_domain.NewCartProduct("p1", 1)}, "", store_domain.CartActive)
	to := store_domain.NewCart("to", "u2", map[string]store_domain.CartProduct{}, "", store_domain.CartActive)
	fromCheckout := store_domain.NewCheckout("fch", "u1", map[string]store_domain.CartProduct{}, "", store_domain.CheckoutPending)
	toCheckout := store_domain.NewCheckout("tch", "u2", map[string]store_domain.CartProduct{}, "", store_domain.CheckoutInvalidated)
	stock := store_domain.Stock{Items: map[string]store_domain.StockItem{}}

	err := runMergeUserCarts(&from, &to, &fromCheckout, &toCheckout, &stock)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestMergeUserCarts_ErrorsWhenReleasingFromReservationFails(t *testing.T) {
	from := store_domain.NewCart("from", "u1", map[string]store_domain.CartProduct{"p1": store_domain.NewCartProduct("p1", 1)}, "", store_domain.CartActive)
	to := store_domain.NewCart("to", "u2", map[string]store_domain.CartProduct{}, "", store_domain.CartActive)
	fromCheckout := store_domain.NewCheckout("fch", "u1", map[string]store_domain.CartProduct{"p1": store_domain.NewCartProduct("p1", 2)}, "", store_domain.CheckoutPending)
	toCheckout := store_domain.NewCheckout("tch", "u2", map[string]store_domain.CartProduct{}, "", store_domain.CheckoutPending)
	stock := store_domain.Stock{Items: map[string]store_domain.StockItem{"p1": store_domain.NewStockItem("p1", 10, 1)}}

	err := runMergeUserCarts(&from, &to, &fromCheckout, &toCheckout, &stock)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestMergeUserCarts_ErrorsWhenReleasingToReservationFails(t *testing.T) {
	from := store_domain.NewCart("from", "u1", map[string]store_domain.CartProduct{"p1": store_domain.NewCartProduct("p1", 1)}, "", store_domain.CartActive)
	to := store_domain.NewCart("to", "u2", map[string]store_domain.CartProduct{}, "", store_domain.CartActive)
	fromCheckout := store_domain.NewCheckout("fch", "u1", map[string]store_domain.CartProduct{}, "", store_domain.CheckoutPending)
	toCheckout := store_domain.NewCheckout("tch", "u2", map[string]store_domain.CartProduct{"p1": store_domain.NewCartProduct("p1", 2)}, "", store_domain.CheckoutPending)
	stock := store_domain.Stock{Items: map[string]store_domain.StockItem{"p1": store_domain.NewStockItem("p1", 10, 1)}}

	err := runMergeUserCarts(&from, &to, &fromCheckout, &toCheckout, &stock)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestMergeUserCarts_MergesAndReleasesReservations(t *testing.T) {
	from := store_domain.NewCart("from", "u1", map[string]store_domain.CartProduct{"p1": store_domain.NewCartProduct("p1", 1)}, "", store_domain.CartActive)
	to := store_domain.NewCart("to", "u2", map[string]store_domain.CartProduct{"p1": store_domain.NewCartProduct("p1", 2)}, "", store_domain.CartActive)
	fromCheckout := store_domain.NewCheckout("fch", "u1", map[string]store_domain.CartProduct{"p1": store_domain.NewCartProduct("p1", 1)}, "", store_domain.CheckoutPending)
	toCheckout := store_domain.NewCheckout("tch", "u2", map[string]store_domain.CartProduct{"p1": store_domain.NewCartProduct("p1", 2)}, "", store_domain.CheckoutPending)
	stock := store_domain.Stock{Items: map[string]store_domain.StockItem{"p1": store_domain.NewStockItem("p1", 7, 3)}}

	err := runMergeUserCarts(&from, &to, &fromCheckout, &toCheckout, &stock)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if to.Products["p1"].Count != 3 {
		t.Fatalf("expected merged count 3, got %d", to.Products["p1"].Count)
	}
	if from.Status != store_domain.CartInactive {
		t.Fatalf("expected from cart inactive, got %s", from.Status)
	}
	if fromCheckout.Status != store_domain.CheckoutInvalidated {
		t.Fatalf("expected from checkout invalidated, got %s", fromCheckout.Status)
	}
	if toCheckout.Status != store_domain.CheckoutInvalidated {
		t.Fatalf("expected to checkout invalidated, got %s", toCheckout.Status)
	}
	if stock.Items["p1"].ReservedAmount != 0 {
		t.Fatalf("expected reserved amount 0, got %d", stock.Items["p1"].ReservedAmount)
	}
}

func TestMergeUserCarts_ReturnsErrorWhenFromCartInactivateFails(t *testing.T) {
	from := store_domain.NewCart("from", "u1", map[string]store_domain.CartProduct{"p1": store_domain.NewCartProduct("p1", 1)}, "", store_domain.CartInactive)
	to := store_domain.NewCart("to", "u2", map[string]store_domain.CartProduct{}, "", store_domain.CartActive)
	fromCheckout := store_domain.Checkout{Items: map[string]store_domain.CartProduct{}}
	toCheckout := store_domain.Checkout{Items: map[string]store_domain.CartProduct{}}
	stock := store_domain.Stock{Items: map[string]store_domain.StockItem{}}

	err := runMergeUserCarts(&from, &to, &fromCheckout, &toCheckout, &stock)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestMergeUserCarts_CreatesZeroTargetCartAndMergesProducts(t *testing.T) {
	from := store_domain.NewCart("from", "guest", map[string]store_domain.CartProduct{"p1": store_domain.NewCartProduct("p1", 2)}, "", store_domain.CartActive)
	to := store_domain.Cart{}
	fromCheckout := store_domain.Checkout{Items: map[string]store_domain.CartProduct{}}
	toCheckout := store_domain.Checkout{Items: map[string]store_domain.CartProduct{}}
	stock := store_domain.Stock{Items: map[string]store_domain.StockItem{}}

	err := runMergeUserCarts(&from, &to, &fromCheckout, &toCheckout, &stock)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if to.ID == "" {
		t.Fatalf("expected target cart id to be initialized")
	}
	if to.CustomerID != "to" {
		t.Fatalf("expected target customer id %q, got %q", "to", to.CustomerID)
	}
	if to.Status != store_domain.CartActive {
		t.Fatalf("expected target cart status %q, got %q", store_domain.CartActive, to.Status)
	}
	if to.Products["p1"].Count != 2 {
		t.Fatalf("expected merged product count %d, got %d", 2, to.Products["p1"].Count)
	}
}

func TestMergeUserCarts_DoesNotFailWhenCheckoutsAreZero(t *testing.T) {
	from := store_domain.NewCart("from", "u1", map[string]store_domain.CartProduct{"p1": store_domain.NewCartProduct("p1", 1)}, "", store_domain.CartActive)
	to := store_domain.NewCart("to", "u2", map[string]store_domain.CartProduct{}, "", store_domain.CartActive)
	fromCheckout := store_domain.Checkout{Items: map[string]store_domain.CartProduct{}}
	toCheckout := store_domain.Checkout{Items: map[string]store_domain.CartProduct{}}
	stock := store_domain.Stock{Items: map[string]store_domain.StockItem{}}

	err := runMergeUserCarts(&from, &to, &fromCheckout, &toCheckout, &stock)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func runMergeUserCarts(
	fromCart *store_domain.Cart,
	toCart *store_domain.Cart,
	fromCheckout *store_domain.Checkout,
	toCheckout *store_domain.Checkout,
	stock *store_domain.Stock,
) error {
	repo := repositoryMock{
		mergeUserCartsFn: func(
			ctx context.Context,
			fromUserID string,
			toUserID string,
			mergeFn func(
				fromCart *store_domain.Cart,
				toCart *store_domain.Cart,
				fromCheckout *store_domain.Checkout,
				toCheckout *store_domain.Checkout,
				stock *store_domain.Stock,
			) error,
		) error {
			return mergeFn(fromCart, toCart, fromCheckout, toCheckout, stock)
		},
	}

	return newServices(repo).MergeUserCarts(context.Background(), "from", "to")
}
