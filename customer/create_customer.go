package customer

import "context"

type CreateCustomerCmd struct {
	Customer Customer
}

func (s *Services) createCustomer(ctx context.Context, cmd CreateCustomerCmd) error {
	return s.repository.CreateCustomer(ctx, cmd.Customer)
}
