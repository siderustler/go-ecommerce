package customer

import "context"

type CustomerQuery struct {
	UserID string
}

func (s *Services) customer(ctx context.Context, query CustomerQuery) (Customer, error) {
	return s.repository.CustomerByID(ctx, query.UserID)
}
