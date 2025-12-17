package basket

import "context"

func (s *Services) BasketByUserID(ctx context.Context, userID string) (Basket, error) {
	return Basket{
		ID:         "",
		CustomerID: userID,
		Products: BasketProducts{
			"1": NewBasketProduct("1", 1),
			"2": NewBasketProduct("2", 2),
			"3": NewBasketProduct("3", 3),
			"4": NewBasketProduct("4", 3),
		},
	}, nil
}
