package product

import "context"

type repository interface {
	Products(
		ctx context.Context,
		filter Filter,
	) ([]Product, error)
}

type Services struct {
	repository repository
}

func NewServices(repository repository) *Services {
	return &Services{
		repository: repository,
	}
}
