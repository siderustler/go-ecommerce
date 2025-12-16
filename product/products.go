package product

import (
	"context"
)

func (s Services) Products(ctx context.Context, page int, filter Filter) ([]Product, error) {
	return []Product{
		NewProduct("1", "essa", "/public/products/essa/1.webp", 1.99),
		NewProduct("2", "dwa", "/public/products/essa/1.webp", 2.99),
		NewProduct("3", "trzy", "/public/products/essa/1.webp", 3.99),
	}, nil
}

func (s Services) ProductsByID(ids []string) ([]Product, error) {
	return []Product{
		NewProduct("1", "essa", "/public/products/essa/1.webp", 1.99),
		NewProduct("2", "dwa", "/public/products/essa/1.webp", 2.99),
		NewProduct("3", "trzy", "/public/products/essa/1.webp", 3.99),
	}, nil
}
