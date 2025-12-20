package customer

import (
	"context"
)

type repository interface {
	CustomerByID(ctx context.Context, userID string) (Customer, error)
	UpdateShippingAddress(ctx context.Context, userID string, shipping ShippingAddress) error
	CreateCustomer(ctx context.Context, customer Customer) error
}

type Services struct {
	repository repository
}

func NewServices(repository repository) *Services {
	return &Services{
		repository: repository,
	}
}
