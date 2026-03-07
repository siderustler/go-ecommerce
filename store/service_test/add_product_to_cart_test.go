package store_service_test

import (
	"context"
	"testing"

	store_domain "github.com/siderustler/go-ecommerce/store/domain"
)

func TestAddProductToCart_ReturnsErrorWhenRequestedStockUnavailable(t *testing.T) {
	cart := activeCart("c1", "u1")
	checkout := zeroCheckout()
	stock := emptyStock()
	requestedProductStock := store_domain.NewStockItem("p1", 0, 0)

	err := runAddProductToCart("u1", store_domain.NewCartProduct("p1", 1), &cart, &checkout, &stock, requestedProductStock)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestAddProductToCart_CreatesZeroCartAndAddsFirstItem(t *testing.T) {
	cart := store_domain.Cart{}
	checkout := zeroCheckout()
	stock := emptyStock()
	requestedProductStock := store_domain.NewStockItem("p2", 10, 0)

	err := runAddProductToCart("u2", store_domain.NewCartProduct("p2", 2), &cart, &checkout, &stock, requestedProductStock)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cart.ID == "" {
		t.Fatalf("expected cart id to be initialized")
	}
	if cart.CustomerID != "u2" {
		t.Fatalf("expected customer id %q, got %q", "u2", cart.CustomerID)
	}
	if cart.Status != store_domain.CartActive {
		t.Fatalf("expected cart status %q, got %q", store_domain.CartActive, cart.Status)
	}
	if cart.Products["p2"].Count != 2 {
		t.Fatalf("expected product count %d, got %d", 2, cart.Products["p2"].Count)
	}
}

func TestAddProductToCart_ReturnsErrorWhenCartInactive(t *testing.T) {
	cart := inactiveCart("c3", "u3")
	checkout := zeroCheckout()
	stock := emptyStock()
	requestedProductStock := store_domain.NewStockItem("p3", 10, 0)

	err := runAddProductToCart("u3", store_domain.NewCartProduct("p3", 1), &cart, &checkout, &stock, requestedProductStock)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestAddProductToCart_ReturnsErrorWhenCheckoutAlreadyInvalidated(t *testing.T) {
	cart := activeCart("c4", "u4")
	checkout := store_domain.NewCheckout(
		"ch4",
		"u4",
		map[string]store_domain.CartProduct{"p4": store_domain.NewCartProduct("p4", 1)},
		"",
		store_domain.CheckoutInvalidated,
	)
	stock := stockWith(store_domain.NewStockItem("p4", 10, 1))
	requestedProductStock := store_domain.NewStockItem("p4", 10, 0)

	err := runAddProductToCart("u4", store_domain.NewCartProduct("p4", 1), &cart, &checkout, &stock, requestedProductStock)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestAddProductToCart_ReturnsErrorWhenReleasingReservationExceedsReservedAmount(t *testing.T) {
	cart := activeCart("c5", "u5")
	checkout := store_domain.NewCheckout(
		"ch5",
		"u5",
		map[string]store_domain.CartProduct{"p5": store_domain.NewCartProduct("p5", 2)},
		"",
		store_domain.CheckoutPending,
	)
	stock := stockWith(store_domain.NewStockItem("p5", 10, 1))
	requestedProductStock := store_domain.NewStockItem("p5", 10, 0)

	err := runAddProductToCart("u5", store_domain.NewCartProduct("p5", 1), &cart, &checkout, &stock, requestedProductStock)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestAddProductToCart_InvalidatesCheckoutAndSkipsMissingStockEntries(t *testing.T) {
	cart := activeCart("c6", "u6")
	checkout := store_domain.NewCheckout(
		"ch6",
		"u6",
		map[string]store_domain.CartProduct{"missing": store_domain.NewCartProduct("missing", 1)},
		"",
		store_domain.CheckoutPending,
	)
	stock := emptyStock()
	requestedProductStock := store_domain.NewStockItem("p6", 10, 0)

	err := runAddProductToCart("u6", store_domain.NewCartProduct("p6", 1), &cart, &checkout, &stock, requestedProductStock)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if checkout.Status != store_domain.CheckoutInvalidated {
		t.Fatalf("expected checkout status %q, got %q", store_domain.CheckoutInvalidated, checkout.Status)
	}
}

func TestAddProductToCart_InvalidatesCheckoutAndReleasesReservations(t *testing.T) {
	cart := activeCart("c7", "u7")
	checkout := store_domain.NewCheckout(
		"ch7",
		"u7",
		map[string]store_domain.CartProduct{"p7": store_domain.NewCartProduct("p7", 2)},
		"",
		store_domain.CheckoutPending,
	)
	stock := stockWith(store_domain.NewStockItem("p7", 8, 2))
	requestedProductStock := store_domain.NewStockItem("p7", 8, 2)

	err := runAddProductToCart("u7", store_domain.NewCartProduct("p7", 1), &cart, &checkout, &stock, requestedProductStock)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if checkout.Status != store_domain.CheckoutInvalidated {
		t.Fatalf("expected checkout status %q, got %q", store_domain.CheckoutInvalidated, checkout.Status)
	}
	if stock.Items["p7"].ReservedAmount != 0 {
		t.Fatalf("expected reserved amount %d, got %d", 0, stock.Items["p7"].ReservedAmount)
	}
	if stock.Items["p7"].AvailableAmount != 10 {
		t.Fatalf("expected available amount %d, got %d", 10, stock.Items["p7"].AvailableAmount)
	}
}

func TestAddProductToCart_ForwardsInputsToRepository(t *testing.T) {
	expectedUserID := "u-forward"
	expectedProduct := store_domain.NewCartProduct("p-forward", 3)

	var gotUserID string
	var gotProduct store_domain.CartProduct

	repo := repositoryMock{
		upsertCartFn: func(
			ctx context.Context,
			incomingUserID string,
			productToAdd store_domain.CartProduct,
			upsertFn func(cart *store_domain.Cart, checkout *store_domain.Checkout, stock *store_domain.Stock, stockItem store_domain.StockItem) error,
		) error {
			gotUserID = incomingUserID
			gotProduct = productToAdd
			return upsertFn(&store_domain.Cart{}, &store_domain.Checkout{}, &store_domain.Stock{Items: map[string]store_domain.StockItem{}}, store_domain.NewStockItem("p-forward", 10, 0))
		},
	}

	err := newServices(repo).AddProductToCart(context.Background(), expectedUserID, expectedProduct)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if gotUserID != expectedUserID {
		t.Fatalf("expected user id %q, got %q", expectedUserID, gotUserID)
	}
	if gotProduct != expectedProduct {
		t.Fatalf("expected product %+v, got %+v", expectedProduct, gotProduct)
	}
}

func runAddProductToCart(
	userID string,
	productToAdd store_domain.CartProduct,
	cart *store_domain.Cart,
	checkout *store_domain.Checkout,
	checkoutStock *store_domain.Stock,
	requestedProductStock store_domain.StockItem,
) error {
	repo := repositoryMock{
		upsertCartFn: func(
			ctx context.Context,
			incomingUserID string,
			incomingProductToAdd store_domain.CartProduct,
			upsertFn func(cart *store_domain.Cart, checkout *store_domain.Checkout, stock *store_domain.Stock, stockItem store_domain.StockItem) error,
		) error {
			return upsertFn(cart, checkout, checkoutStock, requestedProductStock)
		},
	}

	return newServices(repo).AddProductToCart(context.Background(), userID, productToAdd)
}

func activeCart(id string, userID string) store_domain.Cart {
	return store_domain.Cart{
		ID:         id,
		CustomerID: userID,
		Products:   map[string]store_domain.CartProduct{},
		Status:     store_domain.CartActive,
	}
}

func inactiveCart(id string, userID string) store_domain.Cart {
	return store_domain.Cart{
		ID:         id,
		CustomerID: userID,
		Products:   map[string]store_domain.CartProduct{},
		Status:     store_domain.CartInactive,
	}
}

func zeroCheckout() store_domain.Checkout {
	return store_domain.Checkout{Items: map[string]store_domain.CartProduct{}}
}

func emptyStock() store_domain.Stock {
	return store_domain.Stock{Items: map[string]store_domain.StockItem{}}
}

func stockWith(items ...store_domain.StockItem) store_domain.Stock {
	stock := emptyStock()
	for _, item := range items {
		stock.Items[item.ProductID] = item
	}
	return stock
}
