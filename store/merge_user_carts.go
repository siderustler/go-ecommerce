package store

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	store_domain "github.com/siderustler/go-ecommerce/store/domain"
)

type MergeUserCartsCmd struct {
	FromID string
	ToID   string
}

func (s *Services) mergeUserCarts(ctx context.Context, cmd MergeUserCartsCmd) error {
	mergeFn := func(
		fromActiveCart *store_domain.Cart, toActiveCart *store_domain.Cart,
		fromPendingCheckout *store_domain.Checkout, toPendingCheckout *store_domain.Checkout,
		stock *store_domain.Stock,
	) error {
		if fromActiveCart.IsZero() {
			return nil
		}
		if toActiveCart.IsZero() {
			*toActiveCart = store_domain.NewCart(
				uuid.NewString(),
				cmd.ToID,
				make(map[string]store_domain.CartProduct),
				"",
				store_domain.CartActive,
			)
		}
		err := toActiveCart.MergeCart(*fromActiveCart)
		if err != nil {
			return fmt.Errorf("merging carts: %w", err)
		}
		err = fromActiveCart.Inactivate()
		if err != nil {
			return fmt.Errorf("inactivating from cart: %w", err)
		}
		if !fromPendingCheckout.IsZero() {
			err = fromPendingCheckout.Invalidate()
			if err != nil {
				return fmt.Errorf("invalidating from checkout: %w", err)
			}
		}
		if !toPendingCheckout.IsZero() {
			err = toPendingCheckout.Invalidate()
			if err != nil {
				return fmt.Errorf("invalidating to checkout: %w", err)
			}
		}
		for productID, item := range stock.Items {
			fromCheckoutItem, ok := fromPendingCheckout.Items[productID]
			if ok {
				err = item.ReleaseItemReservation(fromCheckoutItem.Count)
				if err != nil {
					return fmt.Errorf("releasing merging checkout stock item reservation: %w", err)
				}
			}
			toCheckoutItem, ok := toPendingCheckout.Items[productID]
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
	return s.repository.MergeUserCarts(ctx, cmd.FromID, cmd.ToID, mergeFn)
}
