package store

import "context"

type repository interface {
	UpsertBasket(ctx context.Context, basket Basket) error
	Basket(ctx context.Context) (Basket, error)
}

type Services struct {
	repository repository
}

func NewServices(repository repository) *Services {
	return &Services{repository: repository}
}
