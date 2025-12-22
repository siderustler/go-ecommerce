package store

import (
	"context"
)

type repository interface {
	InsertStockItem(ctx context.Context, stockItem StockItem) error
	StockItem(ctx context.Context, itemID string) (StockItem, error)
	Checkout(ctx context.Context, checkoutID string) (Checkout, error)
	BasketByUserID(ctx context.Context, userID string) (Basket, error)
	BasketModifyTime(ctx context.Context, basketID string) (string, error)

	UpsertReservations(
		ctx context.Context,
		basketID string,
		productIDs []string,
		reservationTime string,
		upsertFn func(
			stockItem StockItem,
			actualReservation Reservation,
		) (updatedReservation Reservation, updatedStockItem StockItem, err error),
	) error

	UpdateStockItem(
		ctx context.Context,
		itemID string,
		updateFn func(item StockItem) (updatedItem StockItem, err error),
	) error

	UpdateBasket(
		ctx context.Context,
		userID string,
		item BasketProduct,
		onUpdate func(stockItem StockItem) error,
	) error
}

type Services struct {
	repository repository
}

func NewServices(repository repository) *Services {
	return &Services{repository: repository}
}
