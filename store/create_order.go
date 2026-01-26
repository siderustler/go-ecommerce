package store

import (
	"context"
	"fmt"

	store_domain "github.com/siderustler/go-ecommerce/store/domain"
)

func (s Services) CreateOrder(ctx context.Context, order store_domain.Order) error {
	return s.repository.CreateOrder(ctx, order, func(checkout *store_domain.Checkout, stock *store_domain.Stock) error {
		var err error
		if err = checkout.Finalize(); err != nil {
			return fmt.Errorf("finalizing checkout: %w", err)
		}
		for itemID, checkoutItem := range checkout.Items {
			stockItem, ok := stock.Items[itemID]
			if !ok {
				continue
			}
			err = stockItem.RemoveItem(checkoutItem.Count)
			if err != nil {
				return fmt.Errorf("removing item from stock: %w", err)
			}
			stock.Items[itemID] = stockItem
		}
		return nil
	})
}
