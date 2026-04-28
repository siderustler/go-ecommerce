package customer

import "context"

type CreateShallowCustomerCmd struct {
	UserID string
}

func (s *Services) createShallowCustomer(ctx context.Context, cmd CreateShallowCustomerCmd) error {
	return s.repository.CreateShallowCustomer(ctx, cmd.UserID)
}
