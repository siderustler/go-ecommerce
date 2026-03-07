package store_service_test

import (
	"context"
	"testing"

	store_domain "github.com/siderustler/go-ecommerce/store/domain"
)

func TestCreateOrder_ErrorsWhenInactivatingCartFails(t *testing.T) {
	checkout := store_domain.NewCheckout("ch1", "u1", map[string]store_domain.CartProduct{}, "", store_domain.CheckoutPending)
	cart := store_domain.NewCart("c1", "u1", map[string]store_domain.CartProduct{}, "", store_domain.CartInactive)
	stock := store_domain.Stock{Items: map[string]store_domain.StockItem{}}
	products := []store_domain.Product{}

	_, _, _, err := runCreateOrder(&cart, &checkout, &stock, products)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestCreateOrder_ErrorsWhenRemovingReservedItemFails(t *testing.T) {
	checkout := store_domain.NewCheckout("ch2", "u2", map[string]store_domain.CartProduct{"p2": store_domain.NewCartProduct("p2", 2)}, "", store_domain.CheckoutPending)
	cart := store_domain.NewCart("c2", "u2", map[string]store_domain.CartProduct{"p2": store_domain.NewCartProduct("p2", 2)}, "", store_domain.CartActive)
	stock := store_domain.Stock{Items: map[string]store_domain.StockItem{"p2": store_domain.NewStockItem("p2", 10, 1)}}
	products := []store_domain.Product{{ID: "p2", Name: "P2", ActualPrice: 1}}

	_, _, _, err := runCreateOrder(&cart, &checkout, &stock, products)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestCreateOrder_FinalizesCheckoutAndUsesDiscountPrice(t *testing.T) {
	checkout := store_domain.NewCheckout("ch4", "u4", map[string]store_domain.CartProduct{"p4": store_domain.NewCartProduct("p4", 1), "missing": store_domain.NewCartProduct("missing", 1)}, "", store_domain.CheckoutPending)
	cart := store_domain.NewCart("c4", "u4", map[string]store_domain.CartProduct{"p4": store_domain.NewCartProduct("p4", 1)}, "", store_domain.CartActive)
	stock := store_domain.Stock{Items: map[string]store_domain.StockItem{"p4": store_domain.NewStockItem("p4", 10, 1)}}
	products := []store_domain.Product{{ID: "p4", Name: "P4", ActualPrice: 10, DiscountPrice: 7}}

	order, outCart, outCheckout, err := runCreateOrder(&cart, &checkout, &stock, products)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if outCart.Status != store_domain.CartInactive {
		t.Fatalf("expected cart inactive, got %s", outCart.Status)
	}
	if outCheckout.Status != store_domain.CheckoutFinalized {
		t.Fatalf("expected checkout finalized, got %s", outCheckout.Status)
	}
	if stock.Items["p4"].ReservedAmount != 0 {
		t.Fatalf("expected reserved amount 0, got %d", stock.Items["p4"].ReservedAmount)
	}
	if len(order.Products) != 1 {
		t.Fatalf("expected one order product, got %d", len(order.Products))
	}
	if order.Products[0].ItemPrice != 7 {
		t.Fatalf("expected discount price 7, got %v", order.Products[0].ItemPrice)
	}
}

func runCreateOrder(
	cart *store_domain.Cart,
	checkout *store_domain.Checkout,
	stock *store_domain.Stock,
	products []store_domain.Product,
) (store_domain.Order, store_domain.Cart, store_domain.Checkout, error) {
	capturedOrder := store_domain.Order{}
	repo := repositoryMock{
		createOrderFn: func(
			ctx context.Context,
			checkoutID string,
			createFn func(cart *store_domain.Cart, checkout *store_domain.Checkout, stock *store_domain.Stock, products []store_domain.Product) (store_domain.Order, error),
		) error {
			order, err := createFn(cart, checkout, stock, products)
			capturedOrder = order
			return err
		},
	}

	err := newServices(repo).CreateOrder(context.Background(), checkout.ID, "now")
	return capturedOrder, *cart, *checkout, err
}
