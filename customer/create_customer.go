package customer

import "context"

func (s Services) CreateCustomer(ctx context.Context, customer Customer) error {
	return s.repository.CreateCustomer(ctx, customer)
}
