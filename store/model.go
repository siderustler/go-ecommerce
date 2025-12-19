package store

import (
	"errors"
	"fmt"
	"time"
)

type ProductID string
type BasketProducts map[ProductID]BasketProduct

type Basket struct {
	ID             string
	CustomerID     string
	Products       BasketProducts
	LastModifiedAt string
}

func (b Basket) IsZero() bool {
	return b.ID == ""
}

func (b *Basket) AddProduct(productID string, count int) {
	product, inBasket := b.Products[ProductID(productID)]
	if !inBasket {
		b.Products[ProductID(productID)] = NewBasketProduct(productID, 1)
		return
	}
	b.LastModifiedAt = time.Now().UTC().String()
	product.Count += count
	b.Products[ProductID(productID)] = product
}

func (b *Basket) RemoveProduct(productID string, count int) error {
	product, inBasket := b.Products[ProductID(productID)]
	if !inBasket {
		return errors.New("product not in basket")
	}
	b.LastModifiedAt = time.Now().UTC().String()
	product.Count -= count
	isBasketProductDeletable := product.Count <= 0
	if isBasketProductDeletable {
		delete(b.Products, ProductID(productID))
		return nil
	}
	b.Products[ProductID(productID)] = product
	return nil
}

func NewBasket(id, customerID string, products BasketProducts, lastModifiedAt string) Basket {
	return Basket{
		ID:             id,
		CustomerID:     customerID,
		Products:       products,
		LastModifiedAt: lastModifiedAt,
	}
}

type BasketProduct struct {
	ProductID string
	Count     int
}

func NewBasketProduct(id string, count int) BasketProduct {
	return BasketProduct{
		ProductID: id,
		Count:     count,
	}
}

type StockItem struct {
	ProductID       string
	AvailableAmount int
	ReservedAmount  int
}

func NewStockItem(productID string, available, reserved int) StockItem {
	return StockItem{
		ProductID:       productID,
		AvailableAmount: available,
		ReservedAmount:  reserved,
	}
}

func (s *StockItem) ReserveItem(reserveAmount int) error {
	if s.AvailableAmount < reserveAmount {
		return errors.New("reserved amount is greater than available amount")
	}
	s.AvailableAmount -= reserveAmount
	s.ReservedAmount += reserveAmount
	return nil
}

func (s *StockItem) ReleaseItemReservation(reserveAmount int) error {
	if s.ReservedAmount < reserveAmount {
		return errors.New("requested amount to reserve is greater than actual reserved amount")
	}
	s.AvailableAmount += reserveAmount
	s.ReservedAmount -= reserveAmount
	return nil
}

func (s *StockItem) RemoveItem(count int) error {
	if s.AvailableAmount < count {
		return errors.New("requested count is greater than actual available amount")
	}
	if s.ReservedAmount < count {
		return errors.New("requested count is greater than actual reserved amount")
	}
	s.AvailableAmount -= count
	s.ReservedAmount -= count
	return nil
}

type Reservation struct {
	BasketID   string
	ProductID  string
	Amount     int
	ReservedAt string
}

func NewReservation(basketID, productID, reservedAt string, amount int) Reservation {
	return Reservation{
		BasketID:   basketID,
		ProductID:  productID,
		ReservedAt: reservedAt,
		Amount:     amount,
	}
}

func (r Reservation) IsExpired() (bool, error) {
	parsedTime, err := time.Parse(time.RFC3339, r.ReservedAt)
	if err != nil {
		return false, fmt.Errorf("parsing time from db: %w", err)
	}
	timeToExpire := 15 * parsedTime.Minute()
	expiryTime := parsedTime.Add(time.Duration(timeToExpire))
	isExpired := time.Now().UTC().After(expiryTime)

	return isExpired, nil
}

type Checkout struct {
	ID        string
	CreatedAt string
}

func (c Checkout) IsExpired(lastBasketModifyTime string) (bool, error) {
	if c.CreatedAt == "" {
		return false, nil
	}
	parsedTime, err := time.Parse(time.RFC3339, c.CreatedAt)
	if err != nil {
		return true, fmt.Errorf("parsing checkout time: %w", err)
	}
	parsedBasketModifyTime, err := time.Parse(time.RFC3339, lastBasketModifyTime)
	if err != nil {
		return true, fmt.Errorf("parsing basket modify time: %w", err)
	}
	isExpired := parsedTime.Before(parsedBasketModifyTime)
	return isExpired, nil
}
