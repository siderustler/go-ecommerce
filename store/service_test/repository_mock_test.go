package store_service_test

import (
	"context"
	"errors"

	"github.com/siderustler/go-ecommerce/store"
	store_domain "github.com/siderustler/go-ecommerce/store/domain"
)

type repositoryMock struct {
	insertStockItemFn func(ctx context.Context, stockItem store_domain.StockItem, product store_domain.Product) error
	updateStockItemFn func(ctx context.Context, stockItem store_domain.StockItem, updateFn func(stockItem *store_domain.StockItem) error)
	cartFn            func(ctx context.Context, userID string) (store_domain.Cart, error)
	cartCountFn       func(ctx context.Context, userID string) (int, error)
	upsertCartFn      func(
		ctx context.Context,
		userID string,
		item store_domain.CartProduct,
		upsertFn func(cart *store_domain.Cart, checkout *store_domain.Checkout, stock *store_domain.Stock, stockItem store_domain.StockItem) error,
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
	createCheckoutFn func(
		ctx context.Context,
		userID string,
		insertFn func(cart *store_domain.Cart, stock *store_domain.Stock) (store_domain.Checkout, error),
	) error
	updateCheckoutFn func(
		ctx context.Context,
		userID string,
		updateFn func(checkout *store_domain.Checkout, stock *store_domain.Stock) error,
	) error
	createOrderFn func(
		ctx context.Context,
		checkoutID string,
		createFn func(cart *store_domain.Cart, checkout *store_domain.Checkout, stock *store_domain.Stock, products []store_domain.Product) (store_domain.Order, error),
	) error
	checkoutByUserIDFn func(ctx context.Context, userID string) (store_domain.Checkout, error)
}

func (m repositoryMock) InsertStockItem(ctx context.Context, stockItem store_domain.StockItem, product store_domain.Product) error {
	if m.insertStockItemFn == nil {
		return errors.New("InsertStockItem not configured")
	}
	return m.insertStockItemFn(ctx, stockItem, product)
}

func (m repositoryMock) UpdateStockItem(ctx context.Context, stockItem store_domain.StockItem, updateFn func(stockItem *store_domain.StockItem) error) {
	if m.updateStockItemFn != nil {
		m.updateStockItemFn(ctx, stockItem, updateFn)
	}
}

func (m repositoryMock) Cart(ctx context.Context, userID string) (store_domain.Cart, error) {
	if m.cartFn == nil {
		return store_domain.Cart{}, errors.New("Cart not configured")
	}
	return m.cartFn(ctx, userID)
}

func (m repositoryMock) CartCount(ctx context.Context, userID string) (int, error) {
	if m.cartCountFn == nil {
		return 0, errors.New("CartCount not configured")
	}
	return m.cartCountFn(ctx, userID)
}

func (m repositoryMock) UpsertCart(
	ctx context.Context,
	userID string,
	item store_domain.CartProduct,
	upsertFn func(cart *store_domain.Cart, checkout *store_domain.Checkout, stock *store_domain.Stock, stockItem store_domain.StockItem) error,
) error {
	if m.upsertCartFn == nil {
		return errors.New("UpsertCart not configured")
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
		return errors.New("MergeUserCarts not configured")
	}
	return m.mergeUserCartsFn(ctx, fromUserID, toUserID, mergeFn)
}

func (m repositoryMock) CreateCheckout(
	ctx context.Context,
	userID string,
	insertFn func(cart *store_domain.Cart, stock *store_domain.Stock) (store_domain.Checkout, error),
) error {
	if m.createCheckoutFn == nil {
		return errors.New("CreateCheckout not configured")
	}
	return m.createCheckoutFn(ctx, userID, insertFn)
}

func (m repositoryMock) UpdateCheckout(
	ctx context.Context,
	userID string,
	updateFn func(checkout *store_domain.Checkout, stock *store_domain.Stock) error,
) error {
	if m.updateCheckoutFn == nil {
		return errors.New("UpdateCheckout not configured")
	}
	return m.updateCheckoutFn(ctx, userID, updateFn)
}

func (m repositoryMock) CreateOrder(
	ctx context.Context,
	checkoutID string,
	createFn func(cart *store_domain.Cart, checkout *store_domain.Checkout, stock *store_domain.Stock, products []store_domain.Product) (store_domain.Order, error),
) error {
	if m.createOrderFn == nil {
		return errors.New("CreateOrder not configured")
	}
	return m.createOrderFn(ctx, checkoutID, createFn)
}

func (m repositoryMock) CheckoutByUserID(ctx context.Context, userID string) (store_domain.Checkout, error) {
	if m.checkoutByUserIDFn == nil {
		return store_domain.Checkout{}, errors.New("CheckoutByUserID not configured")
	}
	return m.checkoutByUserIDFn(ctx, userID)
}

func newServices(r repositoryMock) *store.Services {
	return store.NewServices(r)
}
