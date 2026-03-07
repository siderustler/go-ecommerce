package store_service_test

import (
	"context"
	"testing"

	store_domain "github.com/siderustler/go-ecommerce/store/domain"
)

func TestRemoveProductFromCart_RejectsWhenStockUnavailable(t *testing.T) {
	cart := store_domain.Cart{Products: map[string]store_domain.CartProduct{"p1": store_domain.NewCartProduct("p1", 1)}, Status: store_domain.CartActive}
	checkout := store_domain.Checkout{Items: map[string]store_domain.CartProduct{}}
	stock := store_domain.Stock{Items: map[string]store_domain.StockItem{}}
	stockItem := store_domain.NewStockItem("p1", 0, 0)

	err := runRemoveProductFromCart("u1", store_domain.NewCartProduct("p1", 1), &cart, &checkout, &stock, stockItem)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if cart.Products["p1"].Count != 1 {
		t.Fatalf("expected count unchanged, got %d", cart.Products["p1"].Count)
	}
}

func TestRemoveProductFromCart_RemovesCountWhenCheckoutIsZero(t *testing.T) {
	cart := store_domain.Cart{ID: "c2", CustomerID: "u2", Products: map[string]store_domain.CartProduct{"p2": store_domain.NewCartProduct("p2", 2)}, Status: store_domain.CartActive}
	checkout := store_domain.Checkout{Items: map[string]store_domain.CartProduct{}}
	stock := store_domain.Stock{Items: map[string]store_domain.StockItem{}}
	stockItem := store_domain.NewStockItem("p2", 10, 0)

	err := runRemoveProductFromCart("u2", store_domain.NewCartProduct("p2", 1), &cart, &checkout, &stock, stockItem)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cart.Products["p2"].Count != 1 {
		t.Fatalf("expected count 1, got %d", cart.Products["p2"].Count)
	}
}

func TestRemoveProductFromCart_RemovesItemWhenCountReachesZero(t *testing.T) {
	cart := store_domain.Cart{ID: "c3", CustomerID: "u3", Products: map[string]store_domain.CartProduct{"p3": store_domain.NewCartProduct("p3", 1)}, Status: store_domain.CartActive}
	checkout := store_domain.Checkout{Items: map[string]store_domain.CartProduct{}}
	stock := store_domain.Stock{Items: map[string]store_domain.StockItem{}}
	stockItem := store_domain.NewStockItem("p3", 10, 0)

	err := runRemoveProductFromCart("u3", store_domain.NewCartProduct("p3", 1), &cart, &checkout, &stock, stockItem)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if _, exists := cart.Products["p3"]; exists {
		t.Fatalf("expected item to be removed")
	}
}

func TestRemoveProductFromCart_ReturnsErrorWhenCartRemoveFails(t *testing.T) {
	cart := store_domain.Cart{ID: "c4", CustomerID: "u4", Products: map[string]store_domain.CartProduct{}, Status: store_domain.CartActive}
	checkout := store_domain.Checkout{Items: map[string]store_domain.CartProduct{}}
	stock := store_domain.Stock{Items: map[string]store_domain.StockItem{}}
	stockItem := store_domain.NewStockItem("p4", 10, 0)

	err := runRemoveProductFromCart("u4", store_domain.NewCartProduct("p4", 1), &cart, &checkout, &stock, stockItem)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestRemoveProductFromCart_ReturnsErrorWhenCheckoutInvalidateFails(t *testing.T) {
	cart := store_domain.Cart{ID: "c5", CustomerID: "u5", Products: map[string]store_domain.CartProduct{"p5": store_domain.NewCartProduct("p5", 2)}, Status: store_domain.CartActive}
	checkout := store_domain.NewCheckout("ch5", "u5", map[string]store_domain.CartProduct{"p5": store_domain.NewCartProduct("p5", 1)}, "", store_domain.CheckoutInvalidated)
	stock := store_domain.Stock{Items: map[string]store_domain.StockItem{"p5": store_domain.NewStockItem("p5", 9, 1)}}
	stockItem := store_domain.NewStockItem("p5", 10, 0)

	err := runRemoveProductFromCart("u5", store_domain.NewCartProduct("p5", 1), &cart, &checkout, &stock, stockItem)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if cart.Products["p5"].Count != 1 {
		t.Fatalf("expected cart mutation before error, got %d", cart.Products["p5"].Count)
	}
}

func TestRemoveProductFromCart_ReturnsErrorWhenReleaseReservationFails(t *testing.T) {
	cart := store_domain.Cart{ID: "c6", CustomerID: "u6", Products: map[string]store_domain.CartProduct{"p6": store_domain.NewCartProduct("p6", 2)}, Status: store_domain.CartActive}
	checkout := store_domain.NewCheckout("ch6", "u6", map[string]store_domain.CartProduct{"p6": store_domain.NewCartProduct("p6", 2)}, "", store_domain.CheckoutPending)
	stock := store_domain.Stock{Items: map[string]store_domain.StockItem{"p6": store_domain.NewStockItem("p6", 9, 1)}}
	stockItem := store_domain.NewStockItem("p6", 10, 0)

	err := runRemoveProductFromCart("u6", store_domain.NewCartProduct("p6", 1), &cart, &checkout, &stock, stockItem)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if cart.Products["p6"].Count != 1 {
		t.Fatalf("expected cart mutation before error, got %d", cart.Products["p6"].Count)
	}
}

func TestRemoveProductFromCart_InvalidatesCheckoutAndReleasesReservation(t *testing.T) {
	cart := store_domain.Cart{ID: "c7", CustomerID: "u7", Products: map[string]store_domain.CartProduct{"p7": store_domain.NewCartProduct("p7", 2)}, Status: store_domain.CartActive}
	checkout := store_domain.NewCheckout("ch7", "u7", map[string]store_domain.CartProduct{"p7": store_domain.NewCartProduct("p7", 2)}, "", store_domain.CheckoutPending)
	stock := store_domain.Stock{Items: map[string]store_domain.StockItem{"p7": store_domain.NewStockItem("p7", 8, 2)}}
	stockItem := store_domain.NewStockItem("p7", 10, 0)

	err := runRemoveProductFromCart("u7", store_domain.NewCartProduct("p7", 1), &cart, &checkout, &stock, stockItem)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cart.Products["p7"].Count != 1 {
		t.Fatalf("expected count 1, got %d", cart.Products["p7"].Count)
	}
	if checkout.Status != store_domain.CheckoutInvalidated {
		t.Fatalf("expected checkout invalidated, got %s", checkout.Status)
	}
	if stock.Items["p7"].ReservedAmount != 0 {
		t.Fatalf("expected reserved amount 0, got %d", stock.Items["p7"].ReservedAmount)
	}
}

func runRemoveProductFromCart(
	userID string,
	input store_domain.CartProduct,
	cart *store_domain.Cart,
	checkout *store_domain.Checkout,
	stock *store_domain.Stock,
	stockItem store_domain.StockItem,
) error {
	repo := repositoryMock{
		upsertCartFn: func(
			ctx context.Context,
			incomingUserID string,
			item store_domain.CartProduct,
			upsertFn func(cart *store_domain.Cart, checkout *store_domain.Checkout, stock *store_domain.Stock, stockItem store_domain.StockItem) error,
		) error {
			return upsertFn(cart, checkout, stock, stockItem)
		},
	}

	return newServices(repo).RemoveProductFromCart(context.Background(), userID, input)
}
