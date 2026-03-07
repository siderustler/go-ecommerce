package store

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	store_domain "github.com/siderustler/go-ecommerce/store/domain"
)

func (s Services) CreateOrder(ctx context.Context, checkoutID, orderTime string) error {
	return s.repository.CreateOrder(
		ctx,
		checkoutID,
		func(cart *store_domain.Cart, checkout *store_domain.Checkout, stock *store_domain.Stock, products []store_domain.Product) (store_domain.Order, error) {
			var err error
			if err = cart.Inactivate(); err != nil {
				return store_domain.Order{}, fmt.Errorf("inactivating cart: %w", err)
			}
			for itemID, checkoutItem := range checkout.Items {
				stockItem, ok := stock.Items[itemID]
				if !ok {
					continue
				}
				err = stockItem.RemoveItem(checkoutItem.Count)
				if err != nil {
					return store_domain.Order{}, fmt.Errorf("removing item from stock: %w", err)
				}

				stock.Items[itemID] = stockItem
			}
			checkout.Finalize()
			orderProducts := make([]store_domain.OrderProduct, 0, len(products))
			for _, product := range products {
				count := cart.Products[product.ID].Count
				itemPrice := product.ActualPrice
				if product.DiscountPrice != 0 {
					itemPrice = product.DiscountPrice
				}
				orderProducts = append(orderProducts, store_domain.NewOrderProduct(product.Name, count, itemPrice))
			}
			return store_domain.NewOrder(uuid.NewString(), checkoutID, orderTime, store_domain.OrderPaid, orderProducts), nil
		})
}
