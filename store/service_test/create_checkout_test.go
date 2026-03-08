package store_service_test

import (
	"context"
	"testing"

	"github.com/siderustler/go-ecommerce/store"
	store_domain "github.com/siderustler/go-ecommerce/store/domain"
)

func TestCheckoutOrCreate_DoesNothingWhenPendingCheckoutAlreadyExists(t *testing.T) {
	existingCheckout := store_domain.NewCheckout(
		"ch-existing",
		"u-existing",
		map[string]store_domain.CartProduct{"p1": store_domain.NewCartProduct("p1", 1)},
		"2026-01-01T00:00:00Z",
		store_domain.CheckoutPending,
	)
	repo := repositoryMock{
		checkoutOrCreateFn: func(
			ctx context.Context,
			incomingUserID string,
			insertFn func(cart *store_domain.Cart, stock *store_domain.Stock) (store_domain.Checkout, error),
		) (store_domain.Checkout, error) {
			return existingCheckout, nil
		},
	}

	checkout, err := newServices(repo).Command.CheckoutOrCreate.Handle(context.Background(), store.CheckoutOrCreateCmd{UserID: "u-existing"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if checkout.ID != existingCheckout.ID {
		t.Fatalf("expected existing checkout %s, got %s", existingCheckout.ID, checkout.ID)
	}
}

func TestCheckoutOrCreate_ErrorsWhenCartIsZero(t *testing.T) {
	cart := store_domain.Cart{Products: map[string]store_domain.CartProduct{}}
	stock := store_domain.Stock{Items: map[string]store_domain.StockItem{}}
	_, err := runCheckoutOrCreate("u1", &cart, &stock)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestCheckoutOrCreate_ErrorsWhenStockItemMissing(t *testing.T) {
	cart := store_domain.Cart{ID: "c2", Products: map[string]store_domain.CartProduct{"p2": store_domain.NewCartProduct("p2", 1)}}
	stock := store_domain.Stock{Items: map[string]store_domain.StockItem{}}
	_, err := runCheckoutOrCreate("u2", &cart, &stock)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestCheckoutOrCreate_ErrorsWhenReservationFails(t *testing.T) {
	cart := store_domain.Cart{ID: "c3", Products: map[string]store_domain.CartProduct{"p3": store_domain.NewCartProduct("p3", 2)}}
	stock := store_domain.Stock{Items: map[string]store_domain.StockItem{"p3": store_domain.NewStockItem("p3", 1, 0)}}
	_, err := runCheckoutOrCreate("u3", &cart, &stock)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestCheckoutOrCreate_ReservesItemsAndReturnsPendingCheckout(t *testing.T) {
	cart := store_domain.Cart{ID: "c4", Products: map[string]store_domain.CartProduct{"p4": store_domain.NewCartProduct("p4", 2)}}
	stock := store_domain.Stock{Items: map[string]store_domain.StockItem{"p4": store_domain.NewStockItem("p4", 5, 1)}}
	checkout, err := runCheckoutOrCreate("u4", &cart, &stock)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if checkout.Status != store_domain.CheckoutPending {
		t.Fatalf("expected pending checkout, got %s", checkout.Status)
	}
	if len(checkout.Items) != 1 {
		t.Fatalf("expected one checkout item, got %d", len(checkout.Items))
	}
	if stock.Items["p4"].AvailableAmount != 3 {
		t.Fatalf("expected available amount 3, got %d", stock.Items["p4"].AvailableAmount)
	}
	if stock.Items["p4"].ReservedAmount != 3 {
		t.Fatalf("expected reserved amount 3, got %d", stock.Items["p4"].ReservedAmount)
	}
}

func runCheckoutOrCreate(userID string, cart *store_domain.Cart, stock *store_domain.Stock) (store_domain.Checkout, error) {
	capturedCheckout := store_domain.Checkout{}
	repo := repositoryMock{
		checkoutOrCreateFn: func(
			ctx context.Context,
			incomingUserID string,
			insertFn func(cart *store_domain.Cart, stock *store_domain.Stock) (store_domain.Checkout, error),
		) (store_domain.Checkout, error) {
			checkout, err := insertFn(cart, stock)
			capturedCheckout = checkout
			return checkout, err
		},
	}

	checkout, err := newServices(repo).Command.CheckoutOrCreate.Handle(context.Background(), store.CheckoutOrCreateCmd{UserID: userID})
	if !checkout.IsZero() {
		capturedCheckout = checkout
	}
	return capturedCheckout, err
}
