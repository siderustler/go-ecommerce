package customer

import (
	"context"
	"log/slog"

	"github.com/siderustler/go-ecommerce/common/service_logger"
)

type repository interface {
	CustomerByID(ctx context.Context, userID string) (Customer, error)
	CreateShallowCustomer(ctx context.Context, userID string) error
	SaveCustomerProfile(ctx context.Context, customer Customer) error
	UpdateShippingAddress(ctx context.Context, userID string, shipping ShippingAddress) error
}

type Services struct {
	repository repository
	Command    Command
	Query      Query
}

type Command struct {
	CreateShallowCustomer service_logger.Command[CreateShallowCustomerCmd]
	SaveCustomerProfile   service_logger.Command[SaveCustomerProfileCmd]
	AddShippingAddress    service_logger.Command[AddShippingAddressCmd]
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
		CreateShallowCustomer: service_logger.NewCommandLoggerDecorator(s.createShallowCustomer, logger),
		SaveCustomerProfile:   service_logger.NewCommandLoggerDecorator(s.saveCustomerProfile, logger),
		AddShippingAddress:    service_logger.NewCommandLoggerDecorator(s.addShippingAddress, logger),
	}
	s.Query = Query{
		Customer: service_logger.NewQueryLoggerDecorator(s.customer, logger),
	}
	return s
}
