package product_test

import (
	"testing"

	"github.com/siderustler/go-ecommerce/product"
)

func TestSpecifySort(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want product.SortSpecifier
	}{
		{name: "price asc", raw: "price-asc", want: product.PriceAsc},
		{name: "price desc", raw: "price-desc", want: product.PriceDesc},
		{name: "name desc", raw: "name-desc", want: product.NameDesc},
		{name: "fallback name asc", raw: "unknown", want: product.NameAsc},
	}

	for _, test := range tests {
		got := product.SpecifySort(test.raw)
		if got != test.want {
			t.Fatalf("test %s failed: expected sort: %s actual sort: %s", test.name, test.want, got)
		}
	}
}

func TestProductPrice(t *testing.T) {
	tests := []struct {
		name   string
		entity product.Product
		expect int
	}{
		{
			name:   "returns discount price when set",
			entity: product.NewPromoProduct("1", "name", "img", 1200, 1500),
			expect: 1200,
		},
		{
			name:   "returns regular price when no discount",
			entity: product.NewProduct("1", "name", "img", 1500),
			expect: 1500,
		},
	}

	for _, test := range tests {
		got := test.entity.ProductPrice()
		if got != test.expect {
			t.Fatalf("test %s failed: expected price: %d actual price: %d", test.name, test.expect, got)
		}
	}
}
