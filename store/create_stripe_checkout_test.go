package store

import (
	"testing"

	"github.com/siderustler/go-ecommerce/product"
	store_domain "github.com/siderustler/go-ecommerce/store/domain"
)

func TestMapCartProductsToStripeLineItems(t *testing.T) {
	cartProducts := map[string]store_domain.CartProduct{
		"p1": store_domain.NewCartProduct("p1", 2),
	}
	products := map[string]product.Product{
		"p1": product.NewPromoProduct("p1", "Product 1", "/img.jpg", 1200, 1500),
	}

	lineItems := mapCartProductsToStripeLineItems(cartProducts, products)
	if len(lineItems) != 1 {
		t.Fatalf("expected 1 line item, got %d", len(lineItems))
	}

	lineItem := lineItems[0]
	if lineItem.Quantity == nil || *lineItem.Quantity != 2 {
		t.Fatalf("expected quantity 2, got %+v", lineItem.Quantity)
	}
	if lineItem.PriceData == nil || lineItem.PriceData.UnitAmountDecimal == nil {
		t.Fatalf("expected unit amount to be set")
	}
	if *lineItem.PriceData.UnitAmountDecimal != 120000 {
		t.Fatalf("expected unit amount decimal 120000, got %.0f", *lineItem.PriceData.UnitAmountDecimal)
	}
	if lineItem.PriceData.ProductData == nil || lineItem.PriceData.ProductData.Name == nil || *lineItem.PriceData.ProductData.Name != "Product 1" {
		t.Fatalf("expected product name Product 1")
	}
}
