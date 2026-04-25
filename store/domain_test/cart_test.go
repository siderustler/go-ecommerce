package store_domain_test

import (
	"errors"
	"testing"

	store_domain "github.com/siderustler/go-ecommerce/store/domain"
)

const fixedLastModifiedAt = "2026-04-25T10:11:12Z"

func fixedTimeNow() string {
	return fixedLastModifiedAt
}

func TestCartStatusFromRawString(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		expected store_domain.CartStatus
	}{
		{
			name:     "active status",
			raw:      "ACTIVE",
			expected: store_domain.CartActive,
		},
		{
			name:     "inactive status",
			raw:      "INACTIVE",
			expected: store_domain.CartInactive,
		},
		{
			name:     "unknown status falls back to inactive",
			raw:      "UNKNOWN",
			expected: store_domain.CartInactive,
		},
	}

	for _, test := range tests {
		actual := store_domain.CartStatusFromRawString(test.raw)
		if actual != test.expected {
			t.Fatalf("test %s failed: expected status: %s actual status: %s", test.name, test.expected, actual)
		}
	}
}

func TestCartAddProduct(t *testing.T) {
	tests := []struct {
		name                    string
		entity                  store_domain.Cart
		cartProduct             store_domain.CartProduct
		expectedError           error
		expectedProducts        map[string]store_domain.CartProduct
		expectLastModifiedAtSet bool
	}{
		{
			name:                    "add product to empty cart",
			entity:                  store_domain.NewCart("cart-1", "user-1", map[string]store_domain.CartProduct{}, "", store_domain.CartActive, fixedTimeNow),
			cartProduct:             store_domain.NewCartProduct("prod-1", 1),
			expectedProducts:        map[string]store_domain.CartProduct{"prod-1": store_domain.NewCartProduct("prod-1", 1)},
			expectLastModifiedAtSet: true,
		},
		{
			name: "add product already existing in cart",
			entity: store_domain.NewCart(
				"cart-1",
				"user-1",
				map[string]store_domain.CartProduct{
					"prod-1": store_domain.NewCartProduct("prod-1", 1),
					"prod-2": store_domain.NewCartProduct("prod-2", 4),
				},
				"before-update",
				store_domain.CartActive,
				fixedTimeNow,
			),
			cartProduct:             store_domain.NewCartProduct("prod-1", 2),
			expectedProducts:        map[string]store_domain.CartProduct{"prod-1": store_domain.NewCartProduct("prod-1", 3), "prod-2": store_domain.NewCartProduct("prod-2", 4)},
			expectLastModifiedAtSet: true,
		},
		{
			name: "reject adding to inactive cart",
			entity: store_domain.NewCart(
				"cart-1",
				"user-1",
				map[string]store_domain.CartProduct{"prod-1": store_domain.NewCartProduct("prod-1", 1)},
				"still-old",
				store_domain.CartInactive,
				fixedTimeNow,
			),
			cartProduct:      store_domain.NewCartProduct("prod-2", 2),
			expectedError:    store_domain.ErrAddProductToInactiveCart,
			expectedProducts: map[string]store_domain.CartProduct{"prod-1": store_domain.NewCartProduct("prod-1", 1)},
		},
		{
			name: "reject non positive count",
			entity: store_domain.NewCart(
				"cart-1",
				"user-1",
				map[string]store_domain.CartProduct{"prod-1": store_domain.NewCartProduct("prod-1", 1)},
				"still-old",
				store_domain.CartActive,
				fixedTimeNow,
			),
			cartProduct:      store_domain.NewCartProduct("prod-2", 0),
			expectedError:    store_domain.ErrNegativeProductCount,
			expectedProducts: map[string]store_domain.CartProduct{"prod-1": store_domain.NewCartProduct("prod-1", 1)},
		},
	}

	for _, test := range tests {
		actual := test.entity
		beforeLastModifiedAt := actual.LastModifiedAt

		err := actual.AddProduct(test.cartProduct)
		assertErrorMessage(t, test.name, err, test.expectedError)
		assertCartProducts(t, test.name, actual.Products, test.expectedProducts)
		if test.expectLastModifiedAtSet {
			assertLastModifiedAtUpdated(t, test.name, beforeLastModifiedAt, actual, fixedLastModifiedAt)
			continue
		}
		assertLastModifiedAtUnchanged(t, test.name, beforeLastModifiedAt, actual)
	}
}

func TestCartRemoveProduct(t *testing.T) {
	tests := []struct {
		name                    string
		entity                  store_domain.Cart
		cartProduct             store_domain.CartProduct
		expectedError           error
		expectedProducts        map[string]store_domain.CartProduct
		expectLastModifiedAtSet bool
	}{
		{
			name: "remove product count from cart",
			entity: store_domain.NewCart(
				"cart-1",
				"user-1",
				map[string]store_domain.CartProduct{
					"prod-1": store_domain.NewCartProduct("prod-1", 3),
					"prod-2": store_domain.NewCartProduct("prod-2", 4),
				},
				"before-update",
				store_domain.CartActive,
				fixedTimeNow,
			),
			cartProduct:             store_domain.NewCartProduct("prod-1", 1),
			expectedProducts:        map[string]store_domain.CartProduct{"prod-1": store_domain.NewCartProduct("prod-1", 2), "prod-2": store_domain.NewCartProduct("prod-2", 4)},
			expectLastModifiedAtSet: true,
		},
		{
			name: "remove product completely from cart",
			entity: store_domain.NewCart(
				"cart-1",
				"user-1",
				map[string]store_domain.CartProduct{
					"prod-1": store_domain.NewCartProduct("prod-1", 1),
					"prod-2": store_domain.NewCartProduct("prod-2", 4),
				},
				"before-update",
				store_domain.CartActive,
				fixedTimeNow,
			),
			cartProduct:             store_domain.NewCartProduct("prod-1", 1),
			expectedProducts:        map[string]store_domain.CartProduct{"prod-2": store_domain.NewCartProduct("prod-2", 4)},
			expectLastModifiedAtSet: true,
		},
		{
			name: "reject removing from inactive cart",
			entity: store_domain.NewCart(
				"cart-1",
				"user-1",
				map[string]store_domain.CartProduct{"prod-1": store_domain.NewCartProduct("prod-1", 1)},
				"still-old",
				store_domain.CartInactive,
				fixedTimeNow,
			),
			cartProduct:      store_domain.NewCartProduct("prod-1", 1),
			expectedError:    store_domain.ErrRemoveProductFromInactiveCart,
			expectedProducts: map[string]store_domain.CartProduct{"prod-1": store_domain.NewCartProduct("prod-1", 1)},
		},
		{
			name: "reject non positive remove count",
			entity: store_domain.NewCart(
				"cart-1",
				"user-1",
				map[string]store_domain.CartProduct{"prod-1": store_domain.NewCartProduct("prod-1", 1)},
				"still-old",
				store_domain.CartActive,
				fixedTimeNow,
			),
			cartProduct:      store_domain.NewCartProduct("prod-1", 0),
			expectedError:    store_domain.ErrNegativeProductCount,
			expectedProducts: map[string]store_domain.CartProduct{"prod-1": store_domain.NewCartProduct("prod-1", 1)},
		},
		{
			name: "reject removing missing product",
			entity: store_domain.NewCart(
				"cart-1",
				"user-1",
				map[string]store_domain.CartProduct{"prod-1": store_domain.NewCartProduct("prod-1", 1)},
				"still-old",
				store_domain.CartActive,
				fixedTimeNow,
			),
			cartProduct:      store_domain.NewCartProduct("prod-2", 1),
			expectedError:    store_domain.ErrProductNotInCart,
			expectedProducts: map[string]store_domain.CartProduct{"prod-1": store_domain.NewCartProduct("prod-1", 1)},
		},
		{
			name: "reject removing more than exists",
			entity: store_domain.NewCart(
				"cart-1",
				"user-1",
				map[string]store_domain.CartProduct{"prod-1": store_domain.NewCartProduct("prod-1", 1)},
				"still-old",
				store_domain.CartActive,
				fixedTimeNow,
			),
			cartProduct:      store_domain.NewCartProduct("prod-1", 2),
			expectedError:    store_domain.ErrRequestedCountExceedsCart,
			expectedProducts: map[string]store_domain.CartProduct{"prod-1": store_domain.NewCartProduct("prod-1", 1)},
		},
	}

	for _, test := range tests {
		actual := test.entity
		beforeLastModifiedAt := actual.LastModifiedAt

		err := actual.RemoveProduct(test.cartProduct)
		assertErrorMessage(t, test.name, err, test.expectedError)
		assertCartProducts(t, test.name, actual.Products, test.expectedProducts)
		if test.expectLastModifiedAtSet {
			assertLastModifiedAtUpdated(t, test.name, beforeLastModifiedAt, actual, fixedLastModifiedAt)
			continue
		}
		assertLastModifiedAtUnchanged(t, test.name, beforeLastModifiedAt, actual)
	}
}

func TestCartMergeCart(t *testing.T) {
	tests := []struct {
		name                    string
		entity                  store_domain.Cart
		other                   store_domain.Cart
		expectedError           error
		expectedProducts        map[string]store_domain.CartProduct
		expectLastModifiedAtSet bool
	}{
		{
			name:                    "merge products into empty cart",
			entity:                  store_domain.NewCart("cart-1", "user-1", map[string]store_domain.CartProduct{}, "", store_domain.CartActive, fixedTimeNow),
			other:                   store_domain.NewCart("cart-2", "user-2", map[string]store_domain.CartProduct{"prod-1": store_domain.NewCartProduct("prod-1", 2)}, "", store_domain.CartActive),
			expectedProducts:        map[string]store_domain.CartProduct{"prod-1": store_domain.NewCartProduct("prod-1", 2)},
			expectLastModifiedAtSet: true,
		},
		{
			name: "merge overlapping products",
			entity: store_domain.NewCart(
				"cart-1",
				"user-1",
				map[string]store_domain.CartProduct{
					"prod-1": store_domain.NewCartProduct("prod-1", 1),
					"prod-2": store_domain.NewCartProduct("prod-2", 2),
				},
				"before-update",
				store_domain.CartActive,
				fixedTimeNow,
			),
			other: store_domain.NewCart(
				"cart-2",
				"user-2",
				map[string]store_domain.CartProduct{
					"prod-1": store_domain.NewCartProduct("prod-1", 3),
					"prod-3": store_domain.NewCartProduct("prod-3", 5),
				},
				"",
				store_domain.CartActive,
			),
			expectedProducts:        map[string]store_domain.CartProduct{"prod-1": store_domain.NewCartProduct("prod-1", 4), "prod-2": store_domain.NewCartProduct("prod-2", 2), "prod-3": store_domain.NewCartProduct("prod-3", 5)},
			expectLastModifiedAtSet: true,
		},
		{
			name: "merge into inactive cart fails",
			entity: store_domain.NewCart(
				"cart-1",
				"user-1",
				map[string]store_domain.CartProduct{"prod-1": store_domain.NewCartProduct("prod-1", 1)},
				"still-old",
				store_domain.CartInactive,
				fixedTimeNow,
			),
			other:            store_domain.NewCart("cart-2", "user-2", map[string]store_domain.CartProduct{"prod-2": store_domain.NewCartProduct("prod-2", 3)}, "", store_domain.CartActive),
			expectedError:    store_domain.ErrAddProductToInactiveCart,
			expectedProducts: map[string]store_domain.CartProduct{"prod-1": store_domain.NewCartProduct("prod-1", 1)},
		},
		{
			name: "merge invalid product count fails fast",
			entity: store_domain.NewCart(
				"cart-1",
				"user-1",
				map[string]store_domain.CartProduct{"prod-1": store_domain.NewCartProduct("prod-1", 1)},
				"still-old",
				store_domain.CartActive,
				fixedTimeNow,
			),
			other: store_domain.NewCart(
				"cart-2",
				"user-2",
				map[string]store_domain.CartProduct{
					"prod-2": store_domain.NewCartProduct("prod-2", 0),
					"prod-3": store_domain.NewCartProduct("prod-3", 4),
				},
				"",
				store_domain.CartActive,
			),
			expectedError:    store_domain.ErrNegativeProductCount,
			expectedProducts: map[string]store_domain.CartProduct{"prod-1": store_domain.NewCartProduct("prod-1", 1)},
		},
	}

	for _, test := range tests {
		actual := test.entity
		beforeLastModifiedAt := actual.LastModifiedAt

		err := actual.MergeCart(test.other)
		assertErrorMessage(t, test.name, err, test.expectedError)
		assertCartProducts(t, test.name, actual.Products, test.expectedProducts)
		if test.expectLastModifiedAtSet {
			assertLastModifiedAtUpdated(t, test.name, beforeLastModifiedAt, actual, fixedLastModifiedAt)
			continue
		}
		assertLastModifiedAtUnchanged(t, test.name, beforeLastModifiedAt, actual)
	}
}

func TestCartInactivate(t *testing.T) {
	tests := []struct {
		name                    string
		entity                  store_domain.Cart
		expectedError           error
		expectedState           store_domain.CartStatus
		expectLastModifiedAtSet bool
	}{
		{
			name:          "active cart becomes inactive",
			entity:        store_domain.NewCart("cart-1", "user-1", map[string]store_domain.CartProduct{}, "", store_domain.CartActive),
			expectedState: store_domain.CartInactive,
		},
		{
			name:          "inactive cart returns error",
			entity:        store_domain.NewCart("cart-1", "user-1", map[string]store_domain.CartProduct{}, "", store_domain.CartInactive),
			expectedError: store_domain.ErrCartInactive,
			expectedState: store_domain.CartInactive,
		},
	}

	for _, test := range tests {
		actual := test.entity
		beforeLastModifiedAt := actual.LastModifiedAt

		err := actual.Inactivate()
		assertErrorMessage(t, test.name, err, test.expectedError)
		if actual.Status != test.expectedState {
			t.Fatalf("test %s failed: expected status: %s actual status: %s", test.name, test.expectedState, actual.Status)
		}
		if test.expectLastModifiedAtSet {
			assertLastModifiedAtUpdated(t, test.name, beforeLastModifiedAt, actual, fixedLastModifiedAt)
			continue
		}
		assertLastModifiedAtUnchanged(t, test.name, beforeLastModifiedAt, actual)
	}
}

func assertErrorMessage(t *testing.T, testName string, err error, expected error) {
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

func assertCartProducts(t *testing.T, testName string, actual, expected map[string]store_domain.CartProduct) {
	t.Helper()

	if len(actual) != len(expected) {
		t.Fatalf("test %s failed: expected product count: %d actual product count: %d", testName, len(expected), len(actual))
	}

	for productID, expectedProduct := range expected {
		actualProduct, exists := actual[productID]
		if !exists {
			t.Fatalf("test %s failed: expected product %s to exist in cart", testName, productID)
		}
		if actualProduct.Count != expectedProduct.Count {
			t.Fatalf(
				"test %s failed: expected product %s count: %d actual product count: %d",
				testName,
				productID,
				expectedProduct.Count,
				actualProduct.Count,
			)
		}
	}
}

func assertLastModifiedAtUpdated(t *testing.T, testName, before string, actual store_domain.Cart, expected string) {
	t.Helper()
	after := actual.LastModifiedAt

	if after == "" {
		t.Fatalf("test %s failed: expected LastModifiedAt to be set", testName)
	}
	if after == before {
		t.Fatalf("test %s failed: expected LastModifiedAt to change", testName)
	}
	if after != expected {
		t.Fatalf("test %s failed: expected LastModifiedAt: %s actual LastModifiedAt: %s", testName, expected, after)
	}
}

func assertLastModifiedAtUnchanged(t *testing.T, testName, before string, actual store_domain.Cart) {
	t.Helper()
	after := actual.LastModifiedAt

	if after != before {
		t.Fatalf("test %s failed: expected LastModifiedAt to remain unchanged", testName)
	}
}
