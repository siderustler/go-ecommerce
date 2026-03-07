package store_service_test

import (
	"context"
	"testing"

	store_domain "github.com/siderustler/go-ecommerce/store/domain"
)

func TestCreateCheckout_ErrorsWhenCartIsZero(t *testing.T) {
	cart := store_domain.Cart{Products: map[string]store_domain.CartProduct{}}
	stock := store_domain.Stock{Items: map[string]store_domain.StockItem{}}
	_, err := runCreateCheckout("u1", &cart, &stock)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestCreateCheckout_ErrorsWhenStockItemMissing(t *testing.T) {
	cart := store_domain.Cart{ID: "c2", Products: map[string]store_domain.CartProduct{"p2": store_domain.NewCartProduct("p2", 1)}}
	stock := store_domain.Stock{Items: map[string]store_domain.StockItem{}}
	_, err := runCreateCheckout("u2", &cart, &stock)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestCreateCheckout_ErrorsWhenReservationFails(t *testing.T) {
	cart := store_domain.Cart{ID: "c3", Products: map[string]store_domain.CartProduct{"p3": store_domain.NewCartProduct("p3", 2)}}
	stock := store_domain.Stock{Items: map[string]store_domain.StockItem{"p3": store_domain.NewStockItem("p3", 1, 0)}}
	_, err := runCreateCheckout("u3", &cart, &stock)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestCreateCheckout_ReservesItemsAndReturnsPendingCheckout(t *testing.T) {
	cart := store_domain.Cart{ID: "c4", Products: map[string]store_domain.CartProduct{"p4": store_domain.NewCartProduct("p4", 2)}}
	stock := store_domain.Stock{Items: map[string]store_domain.StockItem{"p4": store_domain.NewStockItem("p4", 5, 1)}}
	checkout, err := runCreateCheckout("u4", &cart, &stock)
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

func runCreateCheckout(userID string, cart *store_domain.Cart, stock *store_domain.Stock) (store_domain.Checkout, error) {
	capturedCheckout := store_domain.Checkout{}
	repo := repositoryMock{
		createCheckoutFn: func(
			ctx context.Context,
			incomingUserID string,
			insertFn func(cart *store_domain.Cart, stock *store_domain.Stock) (store_domain.Checkout, error),
		) error {
			checkout, err := insertFn(cart, stock)
			capturedCheckout = checkout
			return err
		},
	}

	err := newServices(repo).CreateCheckout(context.Background(), userID)
	return capturedCheckout, err
}
