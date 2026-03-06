package store_domain_test

import (
	"testing"

	store_domain "github.com/siderustler/go-ecommerce/store/domain"
)

func TestCartAddProduct(t *testing.T) {
	type expected struct {
		err          bool
		productCount int
	}
	type addProductTest struct {
		name        string
		entity      store_domain.Cart
		cartProduct store_domain.CartProduct
		expected    expected
	}

	tests := []addProductTest{
		{
			name:        "add product to empty cart",
			entity:      store_domain.NewCart("cart-1", "user-1", map[string]store_domain.CartProduct{}, "", store_domain.CartActive),
			cartProduct: store_domain.NewCartProduct("prod-1", 1),
			expected: expected{
				productCount: 1,
			},
		},
		{
			name: "add product already existing in cart",
			entity: store_domain.NewCart(
				"cart-1",
				"user-1",
				map[string]store_domain.CartProduct{"prod-1": store_domain.NewCartProduct("prod-1", 1)},
				"",
				store_domain.CartActive,
			),
			cartProduct: store_domain.NewCartProduct("prod-1", 2),
			expected: expected{
				productCount: 3,
			},
		},
		{
			name:        "reject non positive count",
			entity:      store_domain.NewCart("cart-1", "user-1", map[string]store_domain.CartProduct{}, "", store_domain.CartActive),
			cartProduct: store_domain.NewCartProduct("prod-1", 0),
			expected: expected{
				err:          true,
				productCount: 0,
			},
		},
	}

	for _, test := range tests {
		actual := test.entity
		err := actual.AddProduct(test.cartProduct)
		isError := err != nil
		if isError != test.expected.err {
			t.Fatalf("test %s failed: expected error: %t actual error: %t", test.name, test.expected.err, isError)
		}

		product, exists := actual.Products[test.cartProduct.ProductID]
		if test.expected.productCount == 0 && exists {
			t.Fatalf("test %s failed: expected product to not exist in cart", test.name)
		}
		if test.expected.productCount > 0 {
			if !exists {
				t.Fatalf("test %s failed: expected product to exist in cart", test.name)
			}
			if product.Count != test.expected.productCount {
				t.Fatalf(
					"test %s failed: expected product count: %d actual product count: %d",
					test.name,
					test.expected.productCount,
					product.Count,
				)
			}
		}
	}
}

func TestCartRemoveProduct(t *testing.T) {
	type expected struct {
		err          bool
		productCount int
	}
	type removeProductTest struct {
		name        string
		entity      store_domain.Cart
		cartProduct store_domain.CartProduct
		expected    expected
	}

	tests := []removeProductTest{
		{
			name: "remove product count from cart",
			entity: store_domain.NewCart(
				"cart-1",
				"user-1",
				map[string]store_domain.CartProduct{"prod-1": store_domain.NewCartProduct("prod-1", 3)},
				"",
				store_domain.CartActive,
			),
			cartProduct: store_domain.NewCartProduct("prod-1", 1),
			expected: expected{
				productCount: 2,
			},
		},
		{
			name: "remove product completely from cart",
			entity: store_domain.NewCart(
				"cart-1",
				"user-1",
				map[string]store_domain.CartProduct{"prod-1": store_domain.NewCartProduct("prod-1", 1)},
				"",
				store_domain.CartActive,
			),
			cartProduct: store_domain.NewCartProduct("prod-1", 1),
			expected: expected{
				productCount: 0,
			},
		},
		{
			name: "reject removing more than exists",
			entity: store_domain.NewCart(
				"cart-1",
				"user-1",
				map[string]store_domain.CartProduct{"prod-1": store_domain.NewCartProduct("prod-1", 1)},
				"",
				store_domain.CartActive,
			),
			cartProduct: store_domain.NewCartProduct("prod-1", 2),
			expected: expected{
				err:          true,
				productCount: 1,
			},
		},
		{
			name: "reject non positive remove count",
			entity: store_domain.NewCart(
				"cart-1",
				"user-1",
				map[string]store_domain.CartProduct{"prod-1": store_domain.NewCartProduct("prod-1", 1)},
				"",
				store_domain.CartActive,
			),
			cartProduct: store_domain.NewCartProduct("prod-1", 0),
			expected: expected{
				err:          true,
				productCount: 1,
			},
		},
	}

	for _, test := range tests {
		actual := test.entity
		err := actual.RemoveProduct(test.cartProduct)
		isError := err != nil
		if isError != test.expected.err {
			t.Fatalf("test %s failed: expected error: %t actual error: %t", test.name, test.expected.err, isError)
		}

		product, exists := actual.Products[test.cartProduct.ProductID]
		if test.expected.productCount == 0 && exists {
			t.Fatalf("test %s failed: expected product to not exist in cart", test.name)
		}
		if test.expected.productCount > 0 {
			if !exists {
				t.Fatalf("test %s failed: expected product to exist in cart", test.name)
			}
			if product.Count != test.expected.productCount {
				t.Fatalf(
					"test %s failed: expected product count: %d actual product count: %d",
					test.name,
					test.expected.productCount,
					product.Count,
				)
			}
		}
	}
}
