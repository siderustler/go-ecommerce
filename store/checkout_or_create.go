package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	store_domain "github.com/siderustler/go-ecommerce/store/domain"
)

func (s Services) CheckoutOrCreate(ctx context.Context, userID string) (store_domain.Checkout, error) {
	checkout, err := s.repository.CheckoutOrCreate(
		ctx,
		userID,
		func(cart *store_domain.Cart, stock *store_domain.Stock) (store_domain.Checkout, error) {
			if cart.IsZero() {
				return store_domain.Checkout{}, errors.New("cart not exists")
			}
			for productID, cartProduct := range cart.Products {
				stockItem, exists := stock.Items[productID]
				if !exists {
					return store_domain.Checkout{}, errors.New("item in cart not exists in stock")
				}
				if err := stockItem.ReserveItem(cartProduct.Count); err != nil {
					return store_domain.Checkout{}, fmt.Errorf("reserving item in stock: %w", err)
				}
				stock.Items[productID] = stockItem
			}
			return store_domain.NewCheckout(
				uuid.NewString(),
				userID,
				cart.Products,
				time.Now().UTC().Format(time.RFC3339),
				store_domain.CheckoutPending,
			), nil
		},
	)
	if err != nil {
		return store_domain.Checkout{}, fmt.Errorf("checkout or create: %w", err)
	}
	return checkout, nil
}
