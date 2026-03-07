package store_service_test

import (
	"context"
	"testing"

	store_domain "github.com/siderustler/go-ecommerce/store/domain"
)

func TestInvalidateCheckout_AlreadyInvalidated_ReturnsError(t *testing.T) {
	checkout := store_domain.NewCheckout("ch1", "u1", map[string]store_domain.CartProduct{}, "", store_domain.CheckoutInvalidated)
	stock := store_domain.Stock{Items: map[string]store_domain.StockItem{}}

	_, _, err := runInvalidateCheckout(checkout, stock)
	if err == nil {
		t.Fatal("expected error for already invalidated checkout")
	}
}

func TestInvalidateCheckout_ReleaseReservationFails_ReturnsError(t *testing.T) {
	checkout := store_domain.NewCheckout(
		"ch2",
		"u2",
		map[string]store_domain.CartProduct{"p2": store_domain.NewCartProduct("p2", 2)},
		"",
		store_domain.CheckoutPending,
	)
	stock := store_domain.Stock{
		Items: map[string]store_domain.StockItem{
			"p2": store_domain.NewStockItem("p2", 10, 1),
		},
	}

	_, _, err := runInvalidateCheckout(checkout, stock)
	if err == nil {
		t.Fatal("expected error when releasing more than reserved")
	}
}

func TestInvalidateCheckout_MissingStockItem_InvalidatesCheckout(t *testing.T) {
	checkout := store_domain.NewCheckout(
		"ch3",
		"u3",
		map[string]store_domain.CartProduct{"missing": store_domain.NewCartProduct("missing", 1)},
		"",
		store_domain.CheckoutPending,
	)
	stock := store_domain.Stock{Items: map[string]store_domain.StockItem{}}

	gotCheckout, _, err := runInvalidateCheckout(checkout, stock)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if gotCheckout.Status != store_domain.CheckoutInvalidated {
		t.Fatalf("expected checkout status %s, got %s", store_domain.CheckoutInvalidated, gotCheckout.Status)
	}
}

func TestInvalidateCheckout_ReleasesReservationAndInvalidatesCheckout(t *testing.T) {
	checkout := store_domain.NewCheckout(
		"ch4",
		"u4",
		map[string]store_domain.CartProduct{"p4": store_domain.NewCartProduct("p4", 2)},
		"",
		store_domain.CheckoutPending,
	)
	stock := store_domain.Stock{
		Items: map[string]store_domain.StockItem{
			"p4": store_domain.NewStockItem("p4", 8, 2),
		},
	}

	gotCheckout, gotStock, err := runInvalidateCheckout(checkout, stock)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if gotCheckout.Status != store_domain.CheckoutInvalidated {
		t.Fatalf("expected checkout status %s, got %s", store_domain.CheckoutInvalidated, gotCheckout.Status)
	}
	item := gotStock.Items["p4"]
	if item.ReservedAmount != 0 {
		t.Fatalf("expected reserved amount 0, got %d", item.ReservedAmount)
	}
}

func runInvalidateCheckout(
	checkout store_domain.Checkout,
	stock store_domain.Stock,
) (store_domain.Checkout, store_domain.Stock, error) {
	capturedCheckout := checkout
	capturedStock := stock

	repo := repositoryMock{
		updateCheckoutFn: func(
			ctx context.Context,
			checkoutID string,
			updateFn func(checkout *store_domain.Checkout, stock *store_domain.Stock) error,
		) error {
			return updateFn(&capturedCheckout, &capturedStock)
		},
	}

	err := newServices(repo).InvalidateCheckout(context.Background(), checkout.ID)
	return capturedCheckout, capturedStock, err
}
