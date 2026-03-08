package store

import "context"

type CartCountQuery struct {
	UserID string
}

func (s *Services) cartCount(ctx context.Context, query CartCountQuery) (int, error) {
	return s.repository.CartCount(ctx, query.UserID)
}
