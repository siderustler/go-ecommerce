package store_domain

import (
	"errors"
	"fmt"
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
		return errors.New("amount to reserve must be greater than zero")
	}
	if s.AvailableAmount < reserveAmount {
		return errors.New("amount to reserve is greater than available amount")
	}
	s.AvailableAmount -= reserveAmount
	s.ReservedAmount += reserveAmount
	return nil
}

func (s *StockItem) ReleaseItemReservation(reserveAmount int) error {
	if reserveAmount <= 0 {
		return errors.New("amount to release must be greater than zero")
	}
	if s.ReservedAmount < reserveAmount {
		return fmt.Errorf("requested amount to reserve is greater than actual reserved amount: actual: %d, request to reserve: %d", s.ReservedAmount, reserveAmount)
	}
	s.AvailableAmount += reserveAmount
	s.ReservedAmount -= reserveAmount
	return nil
}

func (s *StockItem) DecreaseAvailableAmount(count int) error {
	if count <= 0 {
		return errors.New("requested count must be greater than zero")
	}
	if s.AvailableAmount < count {
		return errors.New("requested count is greater than actual available amount")
	}
	s.AvailableAmount -= count

	return nil
}

func (s *StockItem) RemoveItem(count int) error {
	if count <= 0 {
		return errors.New("requested count must be greater than zero")
	}
	if s.ReservedAmount < count {
		return errors.New("requested count is greater than actual reserved amount")
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
