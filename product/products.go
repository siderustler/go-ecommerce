package product

import (
	"context"
)

func (s Services) Products(ctx context.Context, page int, filter Filter) ([]Product, error) {
	return s.repository.Products(ctx, filter)
}
