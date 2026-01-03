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
	if s.AvailableAmount < reserveAmount {
		return errors.New("reserved amount is greater than available amount")
	}
	s.AvailableAmount -= reserveAmount
	s.ReservedAmount += reserveAmount
	return nil
}

func (s *StockItem) ReleaseItemReservation(reserveAmount int) error {
	if s.ReservedAmount < reserveAmount {
		return fmt.Errorf("requested amount to reserve is greater than actual reserved amount: actual: %d, request to reserve: %d", s.ReservedAmount, reserveAmount)
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

func (s StockItem) IsAvailable() bool {
	return s.AvailableAmount > s.ReservedAmount
}

type Stock struct {
	Items map[string]StockItem
}
