package services

import "context"

func (s *Services) AddToBasket(ctx context.Context, sessionID string, productID string, count int) (basketCount int, err error) {
	return 0, nil
}
