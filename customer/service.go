package customer

import (
	"context"
	"log/slog"

	"github.com/siderustler/go-ecommerce/common/service_logger"
)

type repository interface {
	CustomerByID(ctx context.Context, userID string) (Customer, error)
	CustomerOrCreate(ctx context.Context, userID string) (Customer, error)
	UpdateShippingAddress(ctx context.Context, userID string, shipping ShippingAddress) error
	CreateCustomer(ctx context.Context, customer Customer) error
}

type Services struct {
	repository repository
	Command    Command
	Query      Query
}

type Command struct {
	CustomerOrCreate   service_logger.CommandResult[CustomerOrCreateCmd, Customer]
	CreateCustomer     service_logger.Command[CreateCustomerCmd]
	AddShippingAddress service_logger.Command[AddShippingAddressCmd]
}

type Query struct {
	Customer service_logger.Query[CustomerQuery, Customer]
}

func NewServices(repository repository, logger *slog.Logger) *Services {
	if logger == nil {
		logger = slog.Default()
	}

	s := &Services{repository: repository}
	s.Command = Command{
		CustomerOrCreate:   service_logger.NewCommandResultLoggerDecorator(s.customerOrCreate, logger),
		CreateCustomer:     service_logger.NewCommandLoggerDecorator(s.createCustomer, logger),
		AddShippingAddress: service_logger.NewCommandLoggerDecorator(s.addShippingAddress, logger),
	}
	s.Query = Query{
		Customer: service_logger.NewQueryLoggerDecorator(s.customer, logger),
	}
	return s
}
