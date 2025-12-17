package basket

import "context"

func (s Services) AddToBasket(ctx context.Context, userID string, basketProduct BasketProduct) (basketCount int, err error) {
	return 1, nil
}
