package store

import (
	"context"

	store_domain "github.com/siderustler/go-ecommerce/store/domain"
)

type CartQuery struct {
	UserID string
}

func (s *Services) cart(ctx context.Context, query CartQuery) (store_domain.Cart, error) {
	return s.repository.Cart(ctx, query.UserID)
}
