package services

import (
	"context"
)

func (s *Services) GetBasket(ctx context.Context, session string) ([]BasketProduct, error) {
	return []BasketProduct{
		NewBasketProduct("1", "Podkaszarka elektryczna DAEWOO DATR 800E 550W", "/public/products/essa/1.webp", 1.99, 1.00, 1),
		NewBasketProduct("2", "Podkaszarka elektryczna DAEWOO DATR 800E 550W", "/public/products/essa/1.webp", 2.99, 2.00, 2),
		NewBasketProduct("3", "Podkaszarka elektryczna DAEWOO DATR 800E 550W", "/public/products/essa/1.webp", 3.99, 3.12, 3),
		NewBasketProduct("4", "Podkaszarka elektryczna DAEWOO DATR 800E 550W", "/public/products/essa/1.webp", 3.99, 3.12, 3),
	}, nil
}
