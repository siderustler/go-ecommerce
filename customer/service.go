package customer

import (
	"context"
)

type repository interface {
	CustomerByID(ctx context.Context, userID string) (Customer, error)
	UpsertCredentials(ctx context.Context, customer Customer) error
	UpsertBillingAddress(ctx context.Context, userID string, billing Billing) error
	UpsertShippingAddress(ctx context.Context, userID string, shipping ShippingAddress) error
	CreateCredentials(ctx context.Context, customer Customer) error
}

type Services struct {
	repository repository
}

func NewServices(repository repository) *Services {
	return &Services{
		repository: repository,
	}
}
