package customer

import "context"

type AddShippingAddressCmd struct {
	UserID   string
	Shipping ShippingAddress
}

func (s *Services) addShippingAddress(ctx context.Context, cmd AddShippingAddressCmd) error {
	return s.repository.UpdateShippingAddress(ctx, cmd.UserID, cmd.Shipping)
}
