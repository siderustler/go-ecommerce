package store_domain

import (
	"errors"
	"fmt"
	"time"
)

type CartStatus string

var (
	CartActive   = CartStatus("ACTIVE")
	CartInactive = CartStatus("INACTIVE")
)

type Cart struct {
	ID             string
	CustomerID     string
	Products       map[string]CartProduct
	LastModifiedAt string
	Status         CartStatus
}

func CartStatusFromRawString(stat string) CartStatus {
	switch stat {
	case string(CartActive):
		return CartActive
	default:
		return CartInactive
	}
}

func (b Cart) IsZero() bool {
	return b.ID == ""
}

func (b *Cart) AddProduct(cartProduct CartProduct) error {
	if b.Status == CartInactive {
		return errors.New("unable to add product to inactive cart")
	}
	b.LastModifiedAt = time.Now().UTC().Format(time.RFC3339)
	product, inCart := b.Products[cartProduct.ProductID]
	if !inCart {
		b.Products[cartProduct.ProductID] = NewCartProduct(cartProduct.ProductID, cartProduct.Count)
		return nil
	}

	product.Count += cartProduct.Count
	b.Products[cartProduct.ProductID] = product
	return nil
}

func (b *Cart) RemoveProduct(cartProduct CartProduct) error {
	if b.Status == CartInactive {
		return errors.New("unable to remove product from inactive cart")
	}
	product, inCart := b.Products[cartProduct.ProductID]
	if !inCart {
		return errors.New("product not in Cart")
	}
	b.LastModifiedAt = time.Now().UTC().Format(time.RFC3339)
	product.Count -= cartProduct.Count
	b.Products[cartProduct.ProductID] = product
	return nil
}

func (b *Cart) MergeCart(cart Cart) error {
	for _, product := range cart.Products {
		err := b.AddProduct(product)
		if err != nil {
			return fmt.Errorf("adding product to cart: %w", err)
		}
	}
	return nil
}

func (b *Cart) Inactivate() error {
	if b.Status != CartActive {
		return errors.New("status is not active")
	}
	b.Status = CartInactive
	return nil
}

func NewCart(id, customerID string, products map[string]CartProduct, lastModifiedAt string, cartStatus CartStatus) Cart {
	return Cart{
		ID:             id,
		CustomerID:     customerID,
		Products:       products,
		LastModifiedAt: lastModifiedAt,
		Status:         cartStatus,
	}
}

type CartProduct struct {
	ProductID string
	Count     int
}

func NewCartProduct(id string, count int) CartProduct {
	return CartProduct{
		ProductID: id,
		Count:     count,
	}
}
