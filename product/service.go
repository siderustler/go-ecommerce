package product

import (
	"context"
	"log/slog"

	"github.com/siderustler/go-ecommerce/common/service_logger"
)

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
	Query      Query
}

type Query struct {
	Products          service_logger.Query[ProductsQuery, ProductsResult]
	ProductsByIDs     service_logger.Query[ProductsByIDsQuery, map[string]Product]
	Promotions        service_logger.Query[PromotionsQuery, PromotionsResult]
	GetProductDetails service_logger.Query[ProductDetailsQuery, ProductDetail]
}

func NewServices(repository repository, logger *slog.Logger) *Services {
	if logger == nil {
		logger = slog.Default()
	}

	s := &Services{repository: repository}
	s.Query = Query{
		Products:          service_logger.NewQueryLoggerDecorator(s.products, logger),
		ProductsByIDs:     service_logger.NewQueryLoggerDecorator(s.productsByIDs, logger),
		Promotions:        service_logger.NewQueryLoggerDecorator(s.promotions, logger),
		GetProductDetails: service_logger.NewQueryLoggerDecorator(s.getProductDetails, logger),
	}
	return s
}
