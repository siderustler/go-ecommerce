package basket

import "context"

func (s *Services) GetBasket(ctx context.Context, session string) ([]BasketProduct, error) {
	return []BasketProduct{
		NewBasketProduct("1", 1),
		NewBasketProduct("2", 2),
		NewBasketProduct("3", 3),
		NewBasketProduct("4", 3),
	}, nil
}
