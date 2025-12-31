package store

import (
	"context"

	store_domain "github.com/siderustler/go-ecommerce/store/domain"
)

func (s Services) Cart(ctx context.Context, userID string) (store_domain.Cart, error) {
	return s.repository.Cart(ctx, userID)
}
