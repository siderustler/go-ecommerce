package store

import "context"

func (s Services) CartCount(ctx context.Context, userID string) (int, error) {
	return s.repository.CartCount(ctx, userID)
}
