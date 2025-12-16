package basket

import "context"

func (s Services) DecrementBasketItem(ctx context.Context, basketID string, itemID string) error {
	return nil
}

func (s Services) IncrementBasketItem(ctx context.Context, basketID string, itemID string) error {
	return nil
}

func (s Services) AddToBasket(ctx context.Context, sessionID string, productID string, count int) (basketCount int, err error) {
	return 0, nil
}
