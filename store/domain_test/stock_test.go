package store_domain_test

import (
	"errors"
	"testing"

	store_domain "github.com/siderustler/go-ecommerce/store/domain"
)

func TestReserveItem(t *testing.T) {
	tests := []struct {
		name          string
		entity        store_domain.StockItem
		reserveAmount int
		expectedError error
		expectedState store_domain.StockItem
	}{
		{
			name:          "reserve some available items",
			entity:        store_domain.NewStockItem("1", 5, 3),
			reserveAmount: 2,
			expectedState: store_domain.NewStockItem("1", 3, 5),
		},
		{
			name:          "reserve all available items in stock",
			entity:        store_domain.NewStockItem("1", 5, 3),
			reserveAmount: 5,
			expectedState: store_domain.NewStockItem("1", 0, 8),
		},
		{
			name:          "reserve more items than available",
			entity:        store_domain.NewStockItem("1", 5, 3),
			reserveAmount: 6,
			expectedError: store_domain.ErrReserveAmountExceedsAvailable,
			expectedState: store_domain.NewStockItem("1", 5, 3),
		},
		{
			name:          "reserve zero items",
			entity:        store_domain.NewStockItem("1", 5, 3),
			reserveAmount: 0,
			expectedError: store_domain.ErrReserveAmountNonPositive,
			expectedState: store_domain.NewStockItem("1", 5, 3),
		},
		{
			name:          "reserve negative items",
			entity:        store_domain.NewStockItem("1", 5, 3),
			reserveAmount: -1,
			expectedError: store_domain.ErrReserveAmountNonPositive,
			expectedState: store_domain.NewStockItem("1", 5, 3),
		},
	}

	for _, test := range tests {
		actual := test.entity

		err := actual.ReserveItem(test.reserveAmount)

		assertErrorIs(t, test.name, err, test.expectedError)
		assertStockItem(t, test.name, actual, test.expectedState)
	}
}

func TestReleaseItemReservation(t *testing.T) {
	tests := []struct {
		name          string
		entity        store_domain.StockItem
		releaseAmount int
		expectedError error
		expectedState store_domain.StockItem
	}{
		{
			name:          "release some reserved items",
			entity:        store_domain.NewStockItem("1", 5, 3),
			releaseAmount: 2,
			expectedState: store_domain.NewStockItem("1", 7, 1),
		},
		{
			name:          "release all reserved items",
			entity:        store_domain.NewStockItem("1", 5, 3),
			releaseAmount: 3,
			expectedState: store_domain.NewStockItem("1", 8, 0),
		},
		{
			name:          "release more items than actually reserved in stock",
			entity:        store_domain.NewStockItem("1", 5, 3),
			releaseAmount: 4,
			expectedError: store_domain.ErrReleaseAmountExceedsReserved,
			expectedState: store_domain.NewStockItem("1", 5, 3),
		},
		{
			name:          "release zero items",
			entity:        store_domain.NewStockItem("1", 5, 3),
			releaseAmount: 0,
			expectedError: store_domain.ErrReleaseAmountNonPositive,
			expectedState: store_domain.NewStockItem("1", 5, 3),
		},
		{
			name:          "release negative items",
			entity:        store_domain.NewStockItem("1", 5, 3),
			releaseAmount: -1,
			expectedError: store_domain.ErrReleaseAmountNonPositive,
			expectedState: store_domain.NewStockItem("1", 5, 3),
		},
	}

	for _, test := range tests {
		actual := test.entity

		err := actual.ReleaseItemReservation(test.releaseAmount)

		assertErrorIs(t, test.name, err, test.expectedError)
		assertStockItem(t, test.name, actual, test.expectedState)
	}
}

func TestDecreaseAvailableAmount(t *testing.T) {
	tests := []struct {
		name          string
		entity        store_domain.StockItem
		count         int
		expectedError error
		expectedState store_domain.StockItem
	}{
		{
			name:          "decrease available amount",
			entity:        store_domain.NewStockItem("1", 5, 3),
			count:         2,
			expectedState: store_domain.NewStockItem("1", 3, 3),
		},
		{
			name:          "decrease all available items",
			entity:        store_domain.NewStockItem("1", 5, 3),
			count:         5,
			expectedState: store_domain.NewStockItem("1", 0, 3),
		},
		{
			name:          "decrease more items than available",
			entity:        store_domain.NewStockItem("1", 5, 3),
			count:         6,
			expectedError: store_domain.ErrDecreaseAmountExceedsAvailabe,
			expectedState: store_domain.NewStockItem("1", 5, 3),
		},
		{
			name:          "decrease zero items",
			entity:        store_domain.NewStockItem("1", 5, 3),
			count:         0,
			expectedError: store_domain.ErrDecreaseAmountNonPositive,
			expectedState: store_domain.NewStockItem("1", 5, 3),
		},
		{
			name:          "decrease negative items",
			entity:        store_domain.NewStockItem("1", 5, 3),
			count:         -1,
			expectedError: store_domain.ErrDecreaseAmountNonPositive,
			expectedState: store_domain.NewStockItem("1", 5, 3),
		},
	}

	for _, test := range tests {
		actual := test.entity

		err := actual.DecreaseAvailableAmount(test.count)

		assertErrorIs(t, test.name, err, test.expectedError)
		assertStockItem(t, test.name, actual, test.expectedState)
	}
}

func TestRemoveItem(t *testing.T) {
	tests := []struct {
		name          string
		entity        store_domain.StockItem
		count         int
		expectedError error
		expectedState store_domain.StockItem
	}{
		{
			name:          "remove reserved items",
			entity:        store_domain.NewStockItem("1", 5, 3),
			count:         2,
			expectedState: store_domain.NewStockItem("1", 5, 1),
		},
		{
			name:          "remove all reserved items",
			entity:        store_domain.NewStockItem("1", 5, 3),
			count:         3,
			expectedState: store_domain.NewStockItem("1", 5, 0),
		},
		{
			name:          "remove more items than reserved",
			entity:        store_domain.NewStockItem("1", 5, 3),
			count:         4,
			expectedError: store_domain.ErrRemoveAmountExceedsReserved,
			expectedState: store_domain.NewStockItem("1", 5, 3),
		},
		{
			name:          "remove zero items",
			entity:        store_domain.NewStockItem("1", 5, 3),
			count:         0,
			expectedError: store_domain.ErrRemoveAmountNonPositive,
			expectedState: store_domain.NewStockItem("1", 5, 3),
		},
		{
			name:          "remove negative items",
			entity:        store_domain.NewStockItem("1", 5, 3),
			count:         -1,
			expectedError: store_domain.ErrRemoveAmountNonPositive,
			expectedState: store_domain.NewStockItem("1", 5, 3),
		},
	}

	for _, test := range tests {
		actual := test.entity

		err := actual.RemoveItem(test.count)

		assertErrorIs(t, test.name, err, test.expectedError)
		assertStockItem(t, test.name, actual, test.expectedState)
	}
}

func TestStockItemIsAvailable(t *testing.T) {
	tests := []struct {
		name   string
		entity store_domain.StockItem
		want   bool
	}{
		{
			name:   "available when amount is positive",
			entity: store_domain.NewStockItem("1", 1, 10),
			want:   true,
		},
		{
			name:   "not available when amount is zero",
			entity: store_domain.NewStockItem("1", 0, 0),
			want:   false,
		},
		{
			name:   "not available when amount is negative",
			entity: store_domain.NewStockItem("1", -1, 0),
			want:   false,
		},
	}

	for _, test := range tests {
		got := test.entity.IsAvailable()
		if got != test.want {
			t.Fatalf("test %s failed: expected availability %t actual %t", test.name, test.want, got)
		}
	}
}

func assertErrorIs(t *testing.T, testName string, err error, expected error) {
	t.Helper()

	if expected == nil && err != nil {
		t.Fatalf("test %s failed: expected no error, got: %v", testName, err)
	}
	if expected != nil {
		if err == nil {
			t.Fatalf("test %s failed: expected error: %v actual error: <nil>", testName, expected)
		}
		if !errors.Is(err, expected) {
			t.Fatalf("test %s failed: expected error: %v actual error: %v", testName, expected, err)
		}
	}
}

func assertStockItem(t *testing.T, testName string, actual, expected store_domain.StockItem) {
	t.Helper()

	if actual.ProductID != expected.ProductID {
		t.Fatalf("test %s failed: expected product id: %s actual product id: %s", testName, expected.ProductID, actual.ProductID)
	}
	if actual.AvailableAmount != expected.AvailableAmount {
		t.Fatalf(
			"test %s failed: expected available amount: %d actual available amount: %d",
			testName,
			expected.AvailableAmount,
			actual.AvailableAmount,
		)
	}
	if actual.ReservedAmount != expected.ReservedAmount {
		t.Fatalf(
			"test %s failed: expected reserved amount: %d actual reserved amount: %d",
			testName,
			expected.ReservedAmount,
			actual.ReservedAmount,
		)
	}
}
