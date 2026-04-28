package customer

import "context"

type SaveCustomerProfileCmd struct {
	Customer Customer
}

func (s *Services) saveCustomerProfile(ctx context.Context, cmd SaveCustomerProfileCmd) error {
	return s.repository.SaveCustomerProfile(ctx, cmd.Customer)
}
