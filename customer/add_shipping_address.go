package customer

import "context"

func (s Services) AddShippingAddress(ctx context.Context, userID string, shipping ShippingAddress) error {
	return s.repository.UpdateShippingAddress(ctx, userID, shipping)
}
