package store

import (
	"context"

	store_domain "github.com/siderustler/go-ecommerce/store/domain"
)

type CheckoutByUserIDQuery struct {
	UserID string
}

func (s *Services) checkoutByUserID(ctx context.Context, query CheckoutByUserIDQuery) (store_domain.Checkout, error) {
	return s.repository.CheckoutByUserID(ctx, query.UserID)
}
