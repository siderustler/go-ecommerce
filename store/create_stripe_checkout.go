package store

import (
	"context"

	"github.com/siderustler/go-ecommerce/product"
	store_domain "github.com/siderustler/go-ecommerce/store/domain"
	"github.com/stripe/stripe-go/v84"
	"github.com/stripe/stripe-go/v84/checkout/session"
)

func (s Services) CreateStripeCheckout(
	ctx context.Context,
	checkoutID string,
	cartProducts map[string]store_domain.CartProduct,
	products map[string]product.Product,
) (sess *stripe.CheckoutSession, err error) {
	sessionParams := &stripe.CheckoutSessionParams{
		Mode:              stripe.String(string(stripe.CheckoutSessionModePayment)),
		UIMode:            stripe.String("embedded"),
		ReturnURL:         stripe.String("http://localhost:8080/basket/checkout/finalize?session_id={CHECKOUT_SESSION_ID}"),
		LineItems:         mapCartProductsToStripeLineItems(cartProducts, products),
		ClientReferenceID: stripe.String(checkoutID),
	}
	sess, err = session.New(sessionParams)
	if err != nil {
		return nil, err
	}

	return sess, nil
}

func mapCartProductsToStripeLineItems(cartProducts map[string]store_domain.CartProduct, products map[string]product.Product) []*stripe.CheckoutSessionLineItemParams {
	lineItems := make([]*stripe.CheckoutSessionLineItemParams, 0, len(products))
	for productID, cartProduct := range cartProducts {
		product, _ := products[productID]
		unitAmount := float64(product.ProductPrice() * 100)
		lineItem := &stripe.CheckoutSessionLineItemParams{
			PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
				Currency: stripe.String("pln"),
				ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
					Name:   stripe.String(product.Name),
					Images: []*string{stripe.String("http://localhost:8080/" + product.Image)},
				},
				UnitAmountDecimal: &unitAmount,
			},
			Quantity: stripe.Int64(int64(cartProduct.Count)),
		}
		lineItems = append(lineItems, lineItem)
	}
	return lineItems
}
