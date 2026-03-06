package store_domain_test

import (
	"testing"

	store_domain "github.com/siderustler/go-ecommerce/store/domain"
)

type expected struct {
	err    bool
	entity store_domain.StockItem
}

func TestReserveItem(t *testing.T) {
	type reserveItemTest struct {
		name          string
		entity        store_domain.StockItem
		reserveAmount int
		expected      expected
	}

	tests := []reserveItemTest{
		{
			name:          "reserve all available items in stock",
			entity:        store_domain.NewStockItem("1", 5, 3),
			reserveAmount: 5,
			expected: expected{
				entity: store_domain.NewStockItem("1", 0, 8),
			},
		},
		{
			name:          "reserve more items than stock have",
			entity:        store_domain.NewStockItem("1", 5, 3),
			reserveAmount: 6,
			expected: expected{
				err:    true,
				entity: store_domain.NewStockItem("1", 5, 3),
			},
		},
		{
			name:          "reserve zero items",
			entity:        store_domain.NewStockItem("1", 5, 3),
			reserveAmount: 0,
			expected: expected{
				err:    true,
				entity: store_domain.NewStockItem("1", 5, 3),
			},
		},
		{
			name:          "reserve negative items",
			entity:        store_domain.NewStockItem("1", 5, 3),
			reserveAmount: -1,
			expected: expected{
				err:    true,
				entity: store_domain.NewStockItem("1", 5, 3),
			},
		},
		{
			name:          "reserve some available items",
			entity:        store_domain.NewStockItem("1", 5, 3),
			reserveAmount: 2,
			expected: expected{
				entity: store_domain.NewStockItem("1", 3, 5),
			},
		},
	}
	for _, test := range tests {
		actual := test.entity
		err := actual.ReserveItem(test.reserveAmount)
		isError := err != nil
		if isError != test.expected.err {
			t.Fatalf(
				"test %s failed: expected error: %t actual error: %t",
				test.name,
				test.expected.err,
				isError,
			)
		}
		expected := test.expected.entity
		isAvailableAmountCorrect := expected.AvailableAmount == actual.AvailableAmount
		isReservedAmountCorrect := expected.ReservedAmount == actual.ReservedAmount
		if !isAvailableAmountCorrect {
			t.Fatalf(
				"test %s failed: expected available amount: %d actual available amount: %d",
				test.name,
				expected.AvailableAmount,
				actual.AvailableAmount,
			)
		}
		if !isReservedAmountCorrect {
			t.Fatalf(
				"test %s failed: expected reserved amount: %d actual reserved amount: %d",
				test.name,
				expected.ReservedAmount,
				actual.ReservedAmount,
			)
		}
	}
}

func TestReleaseItemReservation(t *testing.T) {
	type releaseItemReservationTest struct {
		name          string
		expected      expected
		releaseAmount int
		entity        store_domain.StockItem
	}

	tests := []releaseItemReservationTest{
		{
			name:          "release item reservations",
			entity:        store_domain.NewStockItem("1", 5, 3),
			releaseAmount: 2,
			expected: expected{
				entity: store_domain.NewStockItem("1", 7, 1),
			},
		},
		{
			name:          "release all reserved items",
			entity:        store_domain.NewStockItem("1", 5, 3),
			releaseAmount: 3,
			expected: expected{
				entity: store_domain.NewStockItem("1", 8, 0),
			},
		},
		{
			name:          "release more items than actually reserved in stock",
			entity:        store_domain.NewStockItem("1", 5, 3),
			releaseAmount: 4,
			expected: expected{
				err:    true,
				entity: store_domain.NewStockItem("1", 5, 3),
			},
		},
		{
			name:          "release zero items",
			entity:        store_domain.NewStockItem("1", 5, 3),
			releaseAmount: 0,
			expected: expected{
				err:    true,
				entity: store_domain.NewStockItem("1", 5, 3),
			},
		},
		{
			name:          "release negative items",
			entity:        store_domain.NewStockItem("1", 5, 3),
			releaseAmount: -1,
			expected: expected{
				err:    true,
				entity: store_domain.NewStockItem("1", 5, 3),
			},
		},
	}
	for _, test := range tests {
		actual := test.entity
		err := actual.ReleaseItemReservation(test.releaseAmount)
		isError := err != nil
		if isError != test.expected.err {
			t.Fatalf(
				"test %s failed: expected error: %t actual error: %t",
				test.name,
				test.expected.err,
				isError,
			)
		}
		expected := test.expected.entity
		isAvailableAmountCorrect := expected.AvailableAmount == actual.AvailableAmount
		isReservedAmountCorrect := expected.ReservedAmount == actual.ReservedAmount
		if !isAvailableAmountCorrect {
			t.Fatalf(
				"test %s failed: expected available amount: %d actual available amount: %d",
				test.name,
				expected.AvailableAmount,
				actual.AvailableAmount)
		}
		if !isReservedAmountCorrect {
			t.Fatalf(
				"test %s failed: expected reserved amount: %d actual reserved amount: %d",
				test.name,
				expected.ReservedAmount,
				actual.ReservedAmount,
			)
		}
	}
}

func TestDecreaseAvailableAmount(t *testing.T) {
	type decreaseAvailableAmountTest struct {
		name     string
		count    int
		entity   store_domain.StockItem
		expected expected
	}

	tests := []decreaseAvailableAmountTest{
		{
			name:   "decrease available amount",
			entity: store_domain.NewStockItem("1", 5, 3),
			count:  2,
			expected: expected{
				entity: store_domain.NewStockItem("1", 3, 3),
			},
		},
		{
			name:   "decrease all available items",
			entity: store_domain.NewStockItem("1", 5, 3),
			count:  5,
			expected: expected{
				entity: store_domain.NewStockItem("1", 0, 3),
			},
		},
		{
			name:   "decrease more items than available",
			entity: store_domain.NewStockItem("1", 5, 3),
			count:  6,
			expected: expected{
				err:    true,
				entity: store_domain.NewStockItem("1", 5, 3),
			},
		},
		{
			name:   "decrease zero items",
			entity: store_domain.NewStockItem("1", 5, 3),
			count:  0,
			expected: expected{
				err:    true,
				entity: store_domain.NewStockItem("1", 5, 3),
			},
		},
		{
			name:   "decrease negative items",
			entity: store_domain.NewStockItem("1", 5, 3),
			count:  -1,
			expected: expected{
				err:    true,
				entity: store_domain.NewStockItem("1", 5, 3),
			},
		},
	}

	for _, test := range tests {
		actual := test.entity
		err := actual.DecreaseAvailableAmount(test.count)
		isError := err != nil

		if isError != test.expected.err {
			t.Fatalf(
				"test %s failed: expected error: %t actual error: %t",
				test.name,
				test.expected.err,
				isError,
			)
		}

		expected := test.expected.entity
		isAvailableAmountCorrect := expected.AvailableAmount == actual.AvailableAmount
		isReservedAmountCorrect := expected.ReservedAmount == actual.ReservedAmount

		if !isAvailableAmountCorrect {
			t.Fatalf(
				"test %s failed: expected available amount: %d actual available amount: %d",
				test.name,
				expected.AvailableAmount,
				actual.AvailableAmount,
			)
		}

		if !isReservedAmountCorrect {
			t.Fatalf(
				"test %s failed: expected reserved amount: %d actual reserved amount: %d",
				test.name,
				expected.ReservedAmount,
				actual.ReservedAmount,
			)
		}
	}
}
func TestRemoveItem(t *testing.T) {
	type removeItemTest struct {
		name     string
		count    int
		entity   store_domain.StockItem
		expected expected
	}

	tests := []removeItemTest{
		{
			name:   "remove reserved items",
			entity: store_domain.NewStockItem("1", 5, 3),
			count:  2,
			expected: expected{
				entity: store_domain.NewStockItem("1", 5, 1),
			},
		},
		{
			name:   "remove all reserved items",
			entity: store_domain.NewStockItem("1", 5, 3),
			count:  3,
			expected: expected{
				entity: store_domain.NewStockItem("1", 5, 0),
			},
		},
		{
			name:   "remove more items than reserved",
			entity: store_domain.NewStockItem("1", 5, 3),
			count:  4,
			expected: expected{
				err:    true,
				entity: store_domain.NewStockItem("1", 5, 3),
			},
		},
		{
			name:   "remove zero items",
			entity: store_domain.NewStockItem("1", 5, 3),
			count:  0,
			expected: expected{
				err:    true,
				entity: store_domain.NewStockItem("1", 5, 3),
			},
		},
		{
			name:   "remove negative items",
			entity: store_domain.NewStockItem("1", 5, 3),
			count:  -1,
			expected: expected{
				err:    true,
				entity: store_domain.NewStockItem("1", 5, 3),
			},
		},
	}

	for _, test := range tests {
		actual := test.entity
		err := actual.RemoveItem(test.count)
		isError := err != nil

		if isError != test.expected.err {
			t.Fatalf(
				"test %s failed: expected error: %t actual error: %t",
				test.name,
				test.expected.err,
				isError,
			)
		}

		expected := test.expected.entity
		isAvailableAmountCorrect := expected.AvailableAmount == actual.AvailableAmount
		isReservedAmountCorrect := expected.ReservedAmount == actual.ReservedAmount

		if !isAvailableAmountCorrect {
			t.Fatalf(
				"test %s failed: expected available amount: %d actual available amount: %d",
				test.name,
				expected.AvailableAmount,
				actual.AvailableAmount,
			)
		}

		if !isReservedAmountCorrect {
			t.Fatalf(
				"test %s failed: expected reserved amount: %d actual reserved amount: %d",
				test.name,
				expected.ReservedAmount,
				actual.ReservedAmount,
			)
		}
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
	}

	for _, test := range tests {
		got := test.entity.IsAvailable()
		if got != test.want {
			t.Fatalf("test %s failed: expected availability %t actual %t", test.name, test.want, got)
		}
	}
}
