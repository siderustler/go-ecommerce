package product

import (
	"context"
)

type ProductsQuery struct {
	Page   int
	Limit  int
	Filter Filter
}

type ProductsResult struct {
	Products         []Product
	AllProductsCount int
}

func (s *Services) products(ctx context.Context, query ProductsQuery) (ProductsResult, error) {
	offset := (query.Page - 1) * query.Limit
	products, allProductsCount, err := s.repository.Products(ctx, offset, query.Limit, query.Filter)
	if err != nil {
		return ProductsResult{}, err
	}
	return ProductsResult{Products: products, AllProductsCount: allProductsCount}, nil
}

type ProductsByIDsQuery struct {
	IDs []string
}

func (s *Services) productsByIDs(ctx context.Context, query ProductsByIDsQuery) (map[string]Product, error) {
	return s.repository.ProductsByIDs(ctx, query.IDs)
}

type PromotionsQuery struct {
	Page     int
	PageSize int
}

type PromotionsResult struct {
	Promos     []Product
	PromoCount int
}

func (s *Services) promotions(ctx context.Context, query PromotionsQuery) (PromotionsResult, error) {
	page := query.Page
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * query.PageSize
	limit := query.PageSize
	promos, promoCount, err := s.repository.Promotions(ctx, offset, limit)
	if err != nil {
		return PromotionsResult{}, err
	}
	return PromotionsResult{Promos: promos, PromoCount: promoCount}, nil
}
