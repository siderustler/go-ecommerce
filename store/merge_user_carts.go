package store

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	store_domain "github.com/siderustler/go-ecommerce/store/domain"
)

func (s Services) MergeUserCarts(ctx context.Context, fromID string, toID string) error {
	mergeFn := func(
		fromCart store_domain.Cart, toCart *store_domain.Cart,
		fromCheckout *store_domain.Checkout, toCheckout *store_domain.Checkout,
		stock *store_domain.Stock,
	) error {
		if fromCart.IsZero() {
			return nil
		}
		if toCart.IsZero() {
			toCart.ID = uuid.NewString()
			toCart.CustomerID = toID
			toCart.Status = store_domain.CartActive
		}
		err := toCart.MergeCart(fromCart)
		if err != nil {
			return fmt.Errorf("merging carts: %w", err)
		}
		err = fromCheckout.Invalidate()
		if err != nil {
			return fmt.Errorf("invalidating from checkout: %w", err)
		}
		err = toCheckout.Invalidate()
		if err != nil {
			return fmt.Errorf("invalidating to checkout: %w", err)
		}
		for productID, item := range stock.Items {
			fromCheckoutItem, ok := fromCheckout.Items[productID]
			if ok {
				err = item.ReleaseItemReservation(fromCheckoutItem.Count)
				if err != nil {
					return fmt.Errorf("releasing merging checkout stock item reservation: %w", err)
				}
			}
			toCheckoutItem, ok := toCheckout.Items[productID]
			if ok {
				err = item.ReleaseItemReservation(toCheckoutItem.Count)
				if err != nil {
					return fmt.Errorf("releasing merge checkout stock item reservation: %w", err)
				}
			}
			stock.Items[productID] = item
		}
		return nil
	}
	return s.repository.MergeUserCarts(ctx, fromID, toID, mergeFn)
}
