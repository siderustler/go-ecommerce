package store_domain

import (
	"errors"
	"fmt"
	"time"
)

type timeNowProvider func() string

var timeNowProviderDefault timeNowProvider = func() string {
	return time.Now().UTC().Format(time.RFC3339)
}

type CartStatus string

var (
	CartActive   = CartStatus("ACTIVE")
	CartInactive = CartStatus("INACTIVE")

	ErrAddProductToInactiveCart      = errors.New("unable to add product to inactive cart")
	ErrRemoveProductFromInactiveCart = errors.New("unable to remove product from inactive cart")
	ErrNegativeProductCount          = errors.New("count must be greater than zero")
	ErrProductNotInCart              = errors.New("product not in Cart")
	ErrRequestedCountExceedsCart     = errors.New("requested count is greater than count in cart")
	ErrCartInactive                  = errors.New("status is not active")
)

type Cart struct {
	ID             string
	CustomerID     string
	Products       map[string]CartProduct
	LastModifiedAt string
	Status         CartStatus
	timeNow        timeNowProvider
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
	if b.timeNow == nil {
		b.timeNow = timeNowProviderDefault
	}
	if b.Status == CartInactive {
		return ErrAddProductToInactiveCart
	}
	if cartProduct.Count <= 0 {
		return ErrNegativeProductCount
	}
	b.LastModifiedAt = b.timeNow()
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
	if b.timeNow == nil {
		b.timeNow = timeNowProviderDefault
	}
	if b.Status == CartInactive {
		return ErrRemoveProductFromInactiveCart
	}
	if cartProduct.Count <= 0 {
		return ErrNegativeProductCount
	}
	product, inCart := b.Products[cartProduct.ProductID]
	if !inCart {
		return ErrProductNotInCart
	}
	if product.Count < cartProduct.Count {
		return ErrRequestedCountExceedsCart
	}
	b.LastModifiedAt = b.timeNow()
	product.Count -= cartProduct.Count
	if product.Count == 0 {
		delete(b.Products, cartProduct.ProductID)
		return nil
	}
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
		return ErrCartInactive
	}
	b.Status = CartInactive
	return nil
}

func NewCart(id, customerID string, products map[string]CartProduct, lastModifiedAt string, cartStatus CartStatus, timers ...timeNowProvider) Cart {
	t := timeNowProviderDefault
	if len(timers) > 0 {
		t = timers[0]
	}
	return Cart{
		ID:             id,
		CustomerID:     customerID,
		Products:       products,
		LastModifiedAt: lastModifiedAt,
		Status:         cartStatus,
		timeNow:        t,
	}
}
