package store

import (
	"errors"
	"time"
)

type Cart struct {
	ID             string
	CustomerID     string
	Products       map[string]CartProduct
	LastModifiedAt string
}

func (b Cart) IsZero() bool {
	return b.ID == ""
}

func (b *Cart) AddProduct(productID string, count int) {
	product, inCart := b.Products[productID]
	if !inCart {
		b.Products[productID] = NewCartProduct(productID, 1)
		return
	}
	b.LastModifiedAt = time.Now().UTC().String()
	product.Count += count
	b.Products[productID] = product
}

func (b *Cart) RemoveProduct(productID string, count int) error {
	product, inCart := b.Products[productID]
	if !inCart {
		return errors.New("product not in Cart")
	}
	b.LastModifiedAt = time.Now().UTC().String()
	product.Count -= count
	isCartProductDeletable := product.Count <= 0
	if isCartProductDeletable {
		delete(b.Products, productID)
		return nil
	}
	b.Products[productID] = product
	return nil
}

func NewCart(id, customerID string, products map[string]CartProduct, lastModifiedAt string) Cart {
	return Cart{
		ID:             id,
		CustomerID:     customerID,
		Products:       products,
		LastModifiedAt: lastModifiedAt,
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
