package store_service_test

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"maps"

	"github.com/siderustler/go-ecommerce/store"
	store_domain "github.com/siderustler/go-ecommerce/store/domain"
)

type repositoryMock struct {
	upsertCartFn func(
		ctx context.Context,
		userID string,
		item store_domain.CartProduct,
		upsertFn func(
			cart *store_domain.Cart,
			checkout *store_domain.Checkout,
			stock *store_domain.Stock,
			stockItem store_domain.StockItem,
		) error,
	) error
	mergeUserCartsFn func(
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
	) error
	checkoutOrCreateFn func(
		ctx context.Context,
		userID string,
		insertFn func(cart *store_domain.Cart, stock *store_domain.Stock) (store_domain.Checkout, error),
	) (store_domain.Checkout, error)
	updateCheckoutFn func(
		ctx context.Context,
		checkoutID string,
		updateFn func(checkout *store_domain.Checkout, stock *store_domain.Stock) error,
	) error
	createOrderFn func(
		ctx context.Context,
		checkoutID string,
		createFn func(cart *store_domain.Cart, checkout *store_domain.Checkout, stock *store_domain.Stock, products []store_domain.Product) (store_domain.Order, error),
	) error
}

func (m repositoryMock) InsertStockItem(ctx context.Context, stockItem store_domain.StockItem, product store_domain.Product) error {
	return fmt.Errorf("InsertStockItem not configured")
}

func (m repositoryMock) UpdateStockItem(ctx context.Context, stockItem store_domain.StockItem, updateFn func(stockItem *store_domain.StockItem) error) {
}

func (m repositoryMock) Cart(ctx context.Context, userID string) (store_domain.Cart, error) {
	return store_domain.Cart{}, fmt.Errorf("Cart not configured")
}

func (m repositoryMock) CartCount(ctx context.Context, userID string) (int, error) {
	return 0, fmt.Errorf("CartCount not configured")
}

func (m repositoryMock) UpsertCart(
	ctx context.Context,
	userID string,
	item store_domain.CartProduct,
	upsertFn func(
		cart *store_domain.Cart,
		checkout *store_domain.Checkout,
		stock *store_domain.Stock,
		stockItem store_domain.StockItem,
	) error,
) error {
	if m.upsertCartFn == nil {
		return fmt.Errorf("UpsertCart not configured")
	}
	return m.upsertCartFn(ctx, userID, item, upsertFn)
}

func (m repositoryMock) MergeUserCarts(
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
	if m.mergeUserCartsFn == nil {
		return fmt.Errorf("MergeUserCarts not configured")
	}
	return m.mergeUserCartsFn(ctx, fromUserID, toUserID, mergeFn)
}

func (m repositoryMock) CheckoutOrCreate(
	ctx context.Context,
	userID string,
	insertFn func(cart *store_domain.Cart, stock *store_domain.Stock) (store_domain.Checkout, error),
) (store_domain.Checkout, error) {
	if m.checkoutOrCreateFn == nil {
		return store_domain.Checkout{}, fmt.Errorf("CheckoutOrCreate not configured")
	}
	return m.checkoutOrCreateFn(ctx, userID, insertFn)
}

func (m repositoryMock) UpdateCheckout(
	ctx context.Context,
	checkoutID string,
	updateFn func(checkout *store_domain.Checkout, stock *store_domain.Stock) error,
) error {
	if m.updateCheckoutFn == nil {
		return fmt.Errorf("UpdateCheckout not configured")
	}
	return m.updateCheckoutFn(ctx, checkoutID, updateFn)
}

func (m repositoryMock) CreateOrder(
	ctx context.Context,
	checkoutID string,
	createFn func(cart *store_domain.Cart, checkout *store_domain.Checkout, stock *store_domain.Stock, products []store_domain.Product) (store_domain.Order, error),
) error {
	if m.createOrderFn == nil {
		return fmt.Errorf("CreateOrder not configured")
	}
	return m.createOrderFn(ctx, checkoutID, createFn)
}

func (m repositoryMock) CheckoutByUserID(ctx context.Context, userID string) (store_domain.Checkout, error) {
	return store_domain.Checkout{}, fmt.Errorf("CheckoutByUserID not configured")
}

func newServices(repo repositoryMock) *store.Services {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return store.NewServices(repo, logger)
}

func zeroCart() store_domain.Cart {
	return store_domain.Cart{Products: map[string]store_domain.CartProduct{}}
}

func zeroCheckout() store_domain.Checkout {
	return store_domain.Checkout{Items: map[string]store_domain.CartProduct{}}
}

func emptyStock() store_domain.Stock {
	return store_domain.Stock{Items: map[string]store_domain.StockItem{}}
}

func activeCart(userID string, products map[string]store_domain.CartProduct) store_domain.Cart {
	return store_domain.NewCart("cart-"+userID, userID, products, "", store_domain.CartActive)
}

func inactiveCart(userID string, products map[string]store_domain.CartProduct) store_domain.Cart {
	return store_domain.NewCart("cart-"+userID, userID, products, "", store_domain.CartInactive)
}

func pendingCheckout(id, userID string, items map[string]store_domain.CartProduct) store_domain.Checkout {
	return store_domain.NewCheckout(id, userID, items, "", store_domain.CheckoutPending)
}

func invalidatedCheckout(id, userID string, items map[string]store_domain.CartProduct) store_domain.Checkout {
	return store_domain.NewCheckout(id, userID, items, "", store_domain.CheckoutInvalidated)
}

func cloneCart(cart store_domain.Cart) store_domain.Cart {
	products := make(map[string]store_domain.CartProduct, len(cart.Products))
	maps.Copy(products, cart.Products)
	return store_domain.Cart{
		ID:             cart.ID,
		CustomerID:     cart.CustomerID,
		Products:       products,
		LastModifiedAt: cart.LastModifiedAt,
		Status:         cart.Status,
	}
}

func cloneCheckout(checkout store_domain.Checkout) store_domain.Checkout {
	items := make(map[string]store_domain.CartProduct, len(checkout.Items))
	maps.Copy(items, checkout.Items)
	return store_domain.Checkout{
		ID:        checkout.ID,
		UserID:    checkout.UserID,
		Items:     items,
		CreatedAt: checkout.CreatedAt,
		Status:    checkout.Status,
	}
}

func cloneStock(stock store_domain.Stock) store_domain.Stock {
	items := make(map[string]store_domain.StockItem, len(stock.Items))
	maps.Copy(items, stock.Items)
	return store_domain.Stock{Items: items}
}

var _ store.Repository = repositoryMock{}
