package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	store_domain "github.com/siderustler/go-ecommerce/store/domain"
)

type AddProductToCartCmd struct {
	UserID       string
	ProductToAdd store_domain.CartProduct
}

func (s *Services) addProductToCart(ctx context.Context, cmd AddProductToCartCmd) error {
	return s.repository.UpsertCart(
		ctx,
		cmd.UserID,
		cmd.ProductToAdd,
		func(cart *store_domain.Cart, checkout *store_domain.Checkout, checkoutStock *store_domain.Stock, requestedProductStock store_domain.StockItem) error {
			if !requestedProductStock.IsAvailable() {
				return errors.New("item is not available in stock")
			}
			if cart.IsZero() {
				newCart := store_domain.NewCart(
					uuid.NewString(),
					cmd.UserID,
					make(map[string]store_domain.CartProduct),
					"",
					store_domain.CartActive,
				)
				*cart = newCart
			}
			err := cart.AddProduct(cmd.ProductToAdd)
			if err != nil {
				return fmt.Errorf("adding product to cart: %w", err)
			}
			if checkout.IsZero() {
				return nil
			}
			err = checkout.Invalidate()
			if err != nil {
				return fmt.Errorf("invalidating existing checkout: %w", err)
			}
			for productID, cartItem := range checkout.Items {
				stockItem, exists := checkoutStock.Items[productID]
				if !exists {
					continue
				}
				err := stockItem.ReleaseItemReservation(cartItem.Count)
				if err != nil {
					return fmt.Errorf("releasing item reservation: %w", err)
				}
				checkoutStock.Items[productID] = stockItem
			}
			return nil
		})
}
