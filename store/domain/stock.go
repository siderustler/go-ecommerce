package store_domain

import (
	"errors"
	"fmt"
)

var (
	ErrReserveAmountNonPositive      = errors.New("amount to reserve must be greater than zero")
	ErrReserveAmountExceedsAvailable = errors.New("amount to reserve is greater than available amount")
	ErrReleaseAmountNonPositive      = errors.New("amount to release must be greater than zero")
	ErrReleaseAmountExceedsReserved  = errors.New("requested amount to reserve is greater than actual reserved amount")
	ErrDecreaseAmountNonPositive     = errors.New("requested count must be greater than zero")
	ErrDecreaseAmountExceedsAvailabe = errors.New("requested count is greater than actual available amount")
	ErrRemoveAmountNonPositive       = errors.New("requested count must be greater than zero")
	ErrRemoveAmountExceedsReserved   = errors.New("requested count is greater than actual reserved amount")
)

type StockItem struct {
	ProductID       string
	AvailableAmount int
	ReservedAmount  int
}

func NewStockItem(productID string, availableAmount int, reservedAmount int) StockItem {
	return StockItem{
		ProductID:       productID,
		AvailableAmount: availableAmount,
		ReservedAmount:  reservedAmount,
	}
}

func (s StockItem) IsZero() bool {
	return s == StockItem{}
}

func (s *StockItem) ReserveItem(reserveAmount int) error {
	if reserveAmount <= 0 {
		return ErrReserveAmountNonPositive
	}
	if s.AvailableAmount < reserveAmount {
		return ErrReserveAmountExceedsAvailable
	}
	s.AvailableAmount -= reserveAmount
	s.ReservedAmount += reserveAmount
	return nil
}

func (s *StockItem) ReleaseItemReservation(reserveAmount int) error {
	if reserveAmount <= 0 {
		return ErrReleaseAmountNonPositive
	}
	if s.ReservedAmount < reserveAmount {
		return fmt.Errorf(
			"%w: actual: %d, request to reserve: %d",
			ErrReleaseAmountExceedsReserved,
			s.ReservedAmount,
			reserveAmount,
		)
	}
	s.AvailableAmount += reserveAmount
	s.ReservedAmount -= reserveAmount
	return nil
}

func (s *StockItem) DecreaseAvailableAmount(count int) error {
	if count <= 0 {
		return ErrDecreaseAmountNonPositive
	}
	if s.AvailableAmount < count {
		return ErrDecreaseAmountExceedsAvailabe
	}
	s.AvailableAmount -= count

	return nil
}

func (s *StockItem) RemoveItem(count int) error {
	if count <= 0 {
		return ErrRemoveAmountNonPositive
	}
	if s.ReservedAmount < count {
		return ErrRemoveAmountExceedsReserved
	}
	s.ReservedAmount -= count
	return nil
}

func (s StockItem) IsAvailable() bool {
	return s.AvailableAmount > 0
}

type Stock struct {
	Items map[string]StockItem
}
