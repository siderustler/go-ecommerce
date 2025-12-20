package customer

import "context"

func (s Services) Customer(ctx context.Context, id string) (Customer, error) {
	return s.repository.CustomerByID(ctx, id)
}
