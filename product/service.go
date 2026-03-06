package product

import "context"

type repository interface {
	Products(
		ctx context.Context,
		offset int,
		limit int,
		filter Filter,
	) (products []Product, allProductsCount int, err error)
	ProductsByIDs(
		ctx context.Context,
		ids []string,
	) (map[string]Product, error)
	Promotions(
		ctx context.Context,
		offset int,
		limit int,
	) (promos []Product, promosCount int, err error)
}

type Services struct {
	repository repository
}

func NewServices(repository repository) *Services {
	return &Services{
		repository: repository,
	}
}
