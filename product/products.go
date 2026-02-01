package product

import (
	"context"
)

func (s Services) Products(ctx context.Context, page int, limit int, filter Filter) ([]Product, error) {
	offset := (page - 1) * limit
	return s.repository.Products(ctx, offset, limit, filter)
}

func (s Services) ProductsByIDs(ctx context.Context, ids []string) (map[string]Product, error) {
	return s.repository.ProductsByIDs(ctx, ids)
}

func (s Services) Promotions(ctx context.Context, page, pageSize int) (promos []Product, promoCount int, err error) {
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize
	limit := pageSize
	return s.repository.Promotions(ctx, offset, limit)
}
