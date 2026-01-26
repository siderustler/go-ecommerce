package store

import (
	"context"
	"fmt"

	store_domain "github.com/siderustler/go-ecommerce/store/domain"
)

func (s Services) CreateOrder(ctx context.Context, order store_domain.Order) error {
	return s.repository.CreateOrder(
		ctx,
		order,
		func(cart *store_domain.Cart, checkout *store_domain.Checkout, stock *store_domain.Stock) error {
			var err error
			if err = cart.Inactivate(); err != nil {
				return fmt.Errorf("inactivating cart: %w", err)
			}
			for itemID, checkoutItem := range checkout.Items {
				stockItem, ok := stock.Items[itemID]
				if !ok {
					continue
				}
				switch checkout.Status {
				case store_domain.CheckoutInvalidated:
					err = stockItem.DecreaseAvailableAmount(checkoutItem.Count)
					if err != nil {
						return fmt.Errorf("decreasing available amount: %w", err)
					}
				case store_domain.CheckoutPending:
					err = stockItem.RemoveItem(checkoutItem.Count)
					if err != nil {
						return fmt.Errorf("removing item from stock: %w", err)
					}
				}

				stock.Items[itemID] = stockItem
			}
			checkout.Finalize()
			return nil
		})
}
