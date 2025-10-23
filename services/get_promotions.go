package services

import (
	"context"
)

func (s *Services) GetPromotions(ctx context.Context) ([]Product, error) {
	return []Product{
		NewProduct("1", "essa", "/public/products/essa/1.webp", 1.99),
		NewProduct("2", "dwa", "/public/products/essa/1.webp", 2.99),
		NewProduct("3", "trzy", "/public/products/essa/1.webp", 3.99),
		NewProduct("4", "cztery", "/public/products/essa/1.webp", 4.99),
	}, nil
}
