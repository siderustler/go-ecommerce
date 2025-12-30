package store

import (
	"context"

	store_domain "github.com/siderustler/go-ecommerce/store2/domain"
)

func (s Services) Cart(ctx context.Context, userID string) (store_domain.Cart, error) {
	return store_domain.Cart{}, nil
}
