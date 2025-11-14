package services

import (
	"context"
)

func (s *Services) GetBasket(ctx context.Context, session string) ([]BasketProduct, error) {
	return []BasketProduct{
		NewBasketProduct("1", "essa", "/public/products/essa/1.webp", 1.99, 1.00, 1),
		NewBasketProduct("2", "dwa", "/public/products/essa/1.webp", 2.99, 2.00, 2),
		NewBasketProduct("3", "trzy", "/public/products/essa/1.webp", 3.99, 3.12, 3),
	}, nil
}
