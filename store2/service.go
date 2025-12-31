package store

import (
	"context"

	store_domain "github.com/siderustler/go-ecommerce/store2/domain"
)

type Repository interface {
	InsertStockItem(
		ctx context.Context,
		stockItem store_domain.StockItem,
		product store_domain.Product,
	) error
	UpdateStockItem(
		ctx context.Context,
		stockItem store_domain.StockItem,
		updateFn func(stockItem *store_domain.StockItem) error,
	)

	Cart(
		ctx context.Context,
		userID string,
	) (store_domain.Cart, error)
	CartCount(
		ctx context.Context,
		userID string,
	) (int, error)

	UpsertCart(
		ctx context.Context,
		userID string,
		item store_domain.CartProduct,
		upsertFn func(cart *store_domain.Cart, checkout *store_domain.Checkout, stock *store_domain.Stock, stockItem store_domain.StockItem) error,
	) error

	CreateCheckout(
		ctx context.Context,
		userID string,
		insertFn func(cart *store_domain.Cart, stock *store_domain.Stock) (store_domain.Checkout, error),
	) error

	UpdateCheckout(
		ctx context.Context,
		checkoutID string,
		updateFn func(checkout *store_domain.Checkout, stock *store_domain.Stock) error,
	) error
}

type Services struct {
	repository Repository
}

func NewServices(repository Repository) *Services {
	return &Services{repository: repository}
}
