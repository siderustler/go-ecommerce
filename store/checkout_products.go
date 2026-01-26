package store

import (
	"context"

	store_domain "github.com/siderustler/go-ecommerce/store/domain"
)

func (s Services) CheckoutProducts(ctx context.Context, checkoutID string) ([]store_domain.OrderProduct, error) {
	return s.repository.CheckoutProducts(ctx, checkoutID)
}
