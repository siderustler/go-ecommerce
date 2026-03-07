package customer

import "context"

func (s Services) CustomerOrCreate(ctx context.Context, userID string) (Customer, error) {
	return s.repository.CustomerOrCreate(ctx, userID)
}
