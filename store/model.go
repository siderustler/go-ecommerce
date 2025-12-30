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
	ProductID  string
	Amount     int
	ReservedAt string
}

func NewReservation(productID, reservedAt string, amount int) Reservation {
	return Reservation{
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

type CheckoutStatus string

var (
	CheckoutPending     CheckoutStatus = "PENDING"
	CheckoutFinalized   CheckoutStatus = "FINALIZED"
	CheckoutInvalidated CheckoutStatus = "INVALIDATED"
)

func CheckoutStatusFromRawString(rawStatus string) CheckoutStatus {
	switch rawStatus {
	case string(CheckoutPending):
		return CheckoutPending
	case string(CheckoutFinalized):
		return CheckoutFinalized
	default:
		return CheckoutInvalidated
	}
}

type Reservations map[ProductID]Reservation
type Stock map[ProductID]StockItem

func (b BasketProducts) MapToReservations() Reservations {
	reservations := make(Reservations, len(b))
	reservationTime := time.Now().UTC().Format(time.RFC3339)
	for _, product := range b {
		reservations[ProductID(product.ProductID)] = NewReservation(product.ProductID, reservationTime, product.Count)
	}
	return reservations
}

type Checkout struct {
	ID             string
	BasketID       string
	CreatedAt      string
	Status         CheckoutStatus
	BasketProducts BasketProducts
}

func NewCheckout(id string, basketID string, createdAt string, status string, basketProducts BasketProducts) Checkout {
	return Checkout{
		ID:             id,
		BasketID:       basketID,
		CreatedAt:      createdAt,
		Status:         CheckoutStatusFromRawString(status),
		BasketProducts: basketProducts,
	}
}

func (c Checkout) IsInvalidated() bool {
	return c.Status == CheckoutInvalidated
}

func (c *Checkout) MarkPending() {
	c.Status = CheckoutPending
}

func (c *Checkout) Invalidate() {
	c.Status = CheckoutInvalidated
}

func (c *Checkout) Finalize() {
	c.Status = CheckoutFinalized
}
