package customer

import "context"

type CustomerOrCreateCmd struct {
	UserID string
}

func (s *Services) customerOrCreate(ctx context.Context, cmd CustomerOrCreateCmd) (Customer, error) {
	return s.repository.CustomerOrCreate(ctx, cmd.UserID)
}
