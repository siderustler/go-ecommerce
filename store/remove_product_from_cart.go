package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	store_domain "github.com/siderustler/go-ecommerce/store/domain"
)

type RemoveProductFromCartCmd struct {
	UserID      string
	CartProduct store_domain.CartProduct
}

func (s *Services) removeProductFromCart(ctx context.Context, cmd RemoveProductFromCartCmd) error {
	return s.repository.UpsertCart(
		ctx,
		cmd.UserID,
		cmd.CartProduct,
		func(cart *store_domain.Cart, checkout *store_domain.Checkout, stock *store_domain.Stock, stockItem store_domain.StockItem) error {
			if !stockItem.IsAvailable() {
				return errors.New("item is not available in stock")
			}
			if cart.IsZero() {
				cart.ID = uuid.NewString()
				cart.CustomerID = cmd.UserID
				cart.Status = store_domain.CartActive
			}
			err := cart.RemoveProduct(cmd.CartProduct)
			if err != nil {
				return fmt.Errorf("removing product from cart: %w", err)
			}
			if checkout.IsZero() {
				return nil
			}
			err = checkout.Invalidate()
			if err != nil {
				return fmt.Errorf("invalidating existing checkout: %w", err)
			}
			for productID, cartItem := range checkout.Items {
				stockItem, exists := stock.Items[productID]
				if !exists {
					continue
				}
				err := stockItem.ReleaseItemReservation(cartItem.Count)
				if err != nil {
					return fmt.Errorf("releasing item reservation: %w", err)
				}
				stock.Items[productID] = stockItem
			}
			return nil
		})
}
