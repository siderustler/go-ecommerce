package store

import (
	"context"
	"log/slog"

	"github.com/siderustler/go-ecommerce/common/service_logger"
	store_domain "github.com/siderustler/go-ecommerce/store/domain"
	"github.com/stripe/stripe-go/v84"
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
			fromCart *store_domain.Cart,
			toCart *store_domain.Cart,
			fromCheckout *store_domain.Checkout,
			toCheckout *store_domain.Checkout,
			stock *store_domain.Stock,
		) error,
	) error

	CheckoutOrCreate(
		ctx context.Context,
		userID string,
		insertFn func(cart *store_domain.Cart, stock *store_domain.Stock) (store_domain.Checkout, error),
	) (store_domain.Checkout, error)

	UpdateCheckout(
		ctx context.Context,
		userID string,
		updateFn func(checkout *store_domain.Checkout, stock *store_domain.Stock) error,
	) error

	CreateOrder(
		ctx context.Context,
		checkoutID string,
		createFn func(cart *store_domain.Cart, checkout *store_domain.Checkout, stock *store_domain.Stock, products []store_domain.Product) (store_domain.Order, error),
	) error

	CheckoutByUserID(
		ctx context.Context,
		userID string,
	) (store_domain.Checkout, error)
}

type Services struct {
	repository Repository
	Command    Command
	Query      Query
}

type Command struct {
	AddProductToCart      service_logger.Command[AddProductToCartCmd]
	RemoveProductFromCart service_logger.Command[RemoveProductFromCartCmd]
	CreateOrder           service_logger.Command[CreateOrderCmd]
	InvalidateCheckout    service_logger.Command[InvalidateCheckoutCmd]
	MergeUserCarts        service_logger.Command[MergeUserCartsCmd]
	CreateStripeCheckout  service_logger.CommandResult[CreateStripeCheckoutCmd, *stripe.CheckoutSession]
	CheckoutOrCreate      service_logger.CommandResult[CheckoutOrCreateCmd, store_domain.Checkout]
}

type Query struct {
	Cart             service_logger.Query[CartQuery, store_domain.Cart]
	CartCount        service_logger.Query[CartCountQuery, int]
	CheckoutByUserID service_logger.Query[CheckoutByUserIDQuery, store_domain.Checkout]
}

func NewServices(repository Repository, logger *slog.Logger) *Services {
	if logger == nil {
		logger = slog.Default()
	}

	s := &Services{repository: repository}
	s.Command = Command{
		AddProductToCart:      service_logger.NewCommandLoggerDecorator(s.addProductToCart, logger),
		RemoveProductFromCart: service_logger.NewCommandLoggerDecorator(s.removeProductFromCart, logger),
		CreateOrder:           service_logger.NewCommandLoggerDecorator(s.createOrder, logger),
		InvalidateCheckout:    service_logger.NewCommandLoggerDecorator(s.invalidateCheckout, logger),
		MergeUserCarts:        service_logger.NewCommandLoggerDecorator(s.mergeUserCarts, logger),
		CheckoutOrCreate:      service_logger.NewCommandResultLoggerDecorator(s.checkoutOrCreate, logger),
		CreateStripeCheckout:  service_logger.NewCommandResultLoggerDecorator(s.createStripeCheckout, logger),
	}
	s.Query = Query{
		Cart:             service_logger.NewQueryLoggerDecorator(s.cart, logger),
		CartCount:        service_logger.NewQueryLoggerDecorator(s.cartCount, logger),
		CheckoutByUserID: service_logger.NewQueryLoggerDecorator(s.checkoutByUserID, logger),
	}

	return s
}
