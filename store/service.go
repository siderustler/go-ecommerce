package store

import (
	"context"

	store_domain "github.com/siderustler/go-ecommerce/store/domain"
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

	MergeUserCarts(
		ctx context.Context,
		fromUserID string,
		toUserID string,
		mergeFn func(
			fromCart store_domain.Cart,
			toCart *store_domain.Cart,
			fromCheckout *store_domain.Checkout,
			toCheckout *store_domain.Checkout,
			stock *store_domain.Stock,
		) error,
	) error

	CreateCheckout(
		ctx context.Context,
		userID string,
		insertFn func(cart *store_domain.Cart, stock *store_domain.Stock) (store_domain.Checkout, error),
	) error

	UpdateCheckout(
		ctx context.Context,
		userID string,
		updateFn func(checkout *store_domain.Checkout, stock *store_domain.Stock) error,
	) error

	CreateOrder(
		ctx context.Context,
		order store_domain.Order,
		createFn func(checkout *store_domain.Checkout, stock *store_domain.Stock) error,
	) error

	CheckoutProducts(
		ctx context.Context,
		checkoutID string,
	) ([]store_domain.OrderProduct, error)
}

type Services struct {
	repository Repository
}

func NewServices(repository Repository) *Services {
	return &Services{repository: repository}
}
