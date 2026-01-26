package store

import (
	"context"

	store_domain "github.com/siderustler/go-ecommerce/store/domain"
)

func (s Services) CheckoutByUserID(ctx context.Context, userID string) (store_domain.Checkout, error) {
	return s.repository.CheckoutByUserID(ctx, userID)
}
