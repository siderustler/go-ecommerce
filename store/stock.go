package store

import "errors"

type StoreStockItem struct {
	ProductID       string
	AvailableAmount int
	ReservedAmount  int
}

type StoreStock map[string]StoreStockItem

func NewStoreStockItem(productID string, availableAmount int, reservedAmount int) StoreStockItem {
	return StoreStockItem{
		ProductID:       productID,
		AvailableAmount: availableAmount,
		ReservedAmount:  reservedAmount,
	}
}

func (s *StoreStockItem) ReserveItem(reserveAmount int) error {
	if s.AvailableAmount < reserveAmount {
		return errors.New("reserved amount is greater than available amount")
	}
	s.AvailableAmount -= reserveAmount
	s.ReservedAmount += reserveAmount
	return nil
}

func (s *StoreStockItem) ReleaseItemReservation(reserveAmount int) error {
	if s.ReservedAmount < reserveAmount {
		return errors.New("requested amount to reserve is greater than actual reserved amount")
	}
	s.AvailableAmount += reserveAmount
	s.ReservedAmount -= reserveAmount
	return nil
}

func (s *StoreStockItem) RemoveItem(count int) error {
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
