package store

import (
	"context"
	"fmt"

	store_domain "github.com/siderustler/go-ecommerce/store/domain"
)

func (s Services) InvalidateCheckout(ctx context.Context, checkoutID string) error {
	return s.repository.UpdateCheckout(
		ctx,
		checkoutID,
		func(checkout *store_domain.Checkout, stock *store_domain.Stock) error {
			err := checkout.Invalidate()
			if err != nil {
				return fmt.Errorf("invalidating checkout: %w", err)
			}
			for productID, checkoutItem := range checkout.Items {
				stockItem, exists := stock.Items[productID]
				if !exists {
					continue
				}
				err := stockItem.ReleaseItemReservation(checkoutItem.Count)
				if err != nil {
					return fmt.Errorf("removing item from stock: %w", err)
				}
				stock.Items[productID] = stockItem
			}
			return nil
		},
	)
}
