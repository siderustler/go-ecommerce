package customer_service_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/cucumber/godog"
	"github.com/google/uuid"
	db_harness "github.com/siderustler/go-ecommerce/adapters/harness_test"
	"github.com/siderustler/go-ecommerce/customer"
	customer_repository "github.com/siderustler/go-ecommerce/customer/repository"
)

type customerFeatureState struct {
	services     *customer.Services
	teardown     func() error
	lastCustomer customer.Customer
	lastErr      error
}

func newCustomerFeatureState() *customerFeatureState {
	return &customerFeatureState{}
}

func (s *customerFeatureState) reset(ctx context.Context) error {
	if err := s.cleanup(); err != nil {
		return err
	}

	teardown, tx, err := db_harness.BeginTestTransaction(ctx, integrationDB)
	if err != nil {
		return fmt.Errorf("begin customer feature transaction: %w", err)
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	txDB := integrationDB.WithTransactionExecutor(tx)
	repo := customer_repository.NewRepository(txDB)

	s.services = customer.NewServices(repo, logger)
	s.teardown = teardown
	s.lastCustomer = customer.Customer{}
	s.lastErr = nil

	return nil
}

func (s *customerFeatureState) cleanup() error {
	if s.teardown == nil {
		return nil
	}
	err := s.teardown()
	s.teardown = nil
	return err
}

func (s *customerFeatureState) givenCustomerDoesNotExistYet(userID string) error {
	return nil
}

func (s *customerFeatureState) givenCustomerAlreadyExistsAsAShallowCustomer(userID string) error {
	return s.services.Command.CreateShallowCustomer.Handle(context.Background(), customer.CreateShallowCustomerCmd{UserID: userID})
}

func (s *customerFeatureState) givenCustomerHasSavedCustomerProfile(userID, name, email, phone, nipCode, company, city, address, postalCode, local string) error {
	profile, err := s.newProfile(userID, name, email, phone, nipCode, company, city, address, postalCode, local)
	if err != nil {
		return err
	}
	return s.services.Command.SaveCustomerProfile.Handle(context.Background(), customer.SaveCustomerProfileCmd{Customer: profile})
}

func (s *customerFeatureState) givenCustomerHasAShippingAddress(userID, city, address, postalCode, local string) error {
	shipping, err := customer.NewShippingAddress(s.shippingID(userID), city, address, postalCode, local)
	if err != nil {
		return err
	}
	return s.services.Command.AddShippingAddress.Handle(context.Background(), customer.AddShippingAddressCmd{UserID: userID, Shipping: shipping})
}

func (s *customerFeatureState) whenLoggedInCustomerIsRecognized(userID string) error {
	s.lastErr = s.services.Command.CreateShallowCustomer.Handle(context.Background(), customer.CreateShallowCustomerCmd{UserID: userID})
	return nil
}

func (s *customerFeatureState) whenBasketCustomerAddsAnItemForTheFirstTime(userID string) error {
	s.lastErr = s.services.Command.CreateShallowCustomer.Handle(context.Background(), customer.CreateShallowCustomerCmd{UserID: userID})
	return nil
}

func (s *customerFeatureState) whenCustomerSavesCustomerProfile(userID, name, email, phone, nipCode, company, city, address, postalCode, local string) error {
	profile, err := s.newProfile(userID, name, email, phone, nipCode, company, city, address, postalCode, local)
	if err != nil {
		return err
	}
	s.lastErr = s.services.Command.SaveCustomerProfile.Handle(context.Background(), customer.SaveCustomerProfileCmd{Customer: profile})
	return nil
}

func (s *customerFeatureState) whenCustomerAddsShippingAddress(userID, city, address, postalCode, local string) error {
	shipping, err := customer.NewShippingAddress(s.shippingID(userID), city, address, postalCode, local)
	if err != nil {
		return err
	}
	s.lastErr = s.services.Command.AddShippingAddress.Handle(context.Background(), customer.AddShippingAddressCmd{UserID: userID, Shipping: shipping})
	return nil
}

func (s *customerFeatureState) whenTheCustomerServiceRetrievesCustomer(userID string) error {
	s.lastCustomer, s.lastErr = s.services.Query.Customer.Handle(context.Background(), customer.CustomerQuery{UserID: userID})
	return nil
}

func (s *customerFeatureState) thenTheCustomerCommandSucceeds() error {
	if s.lastErr != nil {
		return fmt.Errorf("expected command success, got %v", s.lastErr)
	}
	return nil
}

func (s *customerFeatureState) thenTheCustomerServiceRejectsBecauseTheCustomerAlreadyExists() error {
	if !errors.Is(s.lastErr, customer.ErrCustomerAlreadyExists) {
		return fmt.Errorf("expected ErrCustomerAlreadyExists, got %v", s.lastErr)
	}
	return nil
}

func (s *customerFeatureState) thenTheCustomerServiceRejectsBecauseTheCustomerIsMissing() error {
	if !errors.Is(s.lastErr, customer.ErrCustomerNotFound) {
		return fmt.Errorf("expected ErrCustomerNotFound, got %v", s.lastErr)
	}
	return nil
}

func (s *customerFeatureState) thenCustomerCanBeRetrievedAsAShallowCustomer(userID string) error {
	cust, err := s.services.Query.Customer.Handle(context.Background(), customer.CustomerQuery{UserID: userID})
	if err != nil {
		return fmt.Errorf("query customer: %w", err)
	}
	if cust.ID != userID {
		return fmt.Errorf("expected customer id %q, got %q", userID, cust.ID)
	}
	if !cust.Credentials.IsZero() {
		return fmt.Errorf("expected shallow customer credentials to be empty, got %#v", cust.Credentials)
	}
	if !cust.Billing.IsZero() {
		return fmt.Errorf("expected shallow customer billing to be empty, got %#v", cust.Billing)
	}
	if !cust.Shipping.IsZero() {
		return fmt.Errorf("expected shallow customer shipping to be empty, got %#v", cust.Shipping)
	}
	return nil
}

func (s *customerFeatureState) thenTheCustomerProfileContainsCredentialsAndBillingDetails(userID, name, email, phone, nipCode, company, city, address, postalCode, local string) error {
	cust, err := s.services.Query.Customer.Handle(context.Background(), customer.CustomerQuery{UserID: userID})
	if err != nil {
		return fmt.Errorf("query customer: %w", err)
	}
	if err := assertCustomerCredentials(cust, userID, name, email, phone); err != nil {
		return err
	}
	if err := assertCustomerBilling(cust, nipCode, company, city, address, postalCode, local); err != nil {
		return err
	}
	return nil
}

func (s *customerFeatureState) thenTheCustomerProfileContainsShippingAddress(userID, city, address, postalCode, local string) error {
	cust, err := s.services.Query.Customer.Handle(context.Background(), customer.CustomerQuery{UserID: userID})
	if err != nil {
		return fmt.Errorf("query customer: %w", err)
	}
	return assertCustomerShipping(cust, city, address, postalCode, local)
}

func (s *customerFeatureState) thenTheCustomerIsMissing(userID string) error {
	if s.lastErr != nil {
		return fmt.Errorf("expected missing customer query without error, got %v", s.lastErr)
	}
	if !s.lastCustomer.IsZero() {
		return fmt.Errorf("expected missing customer %q to be zero, got %#v", userID, s.lastCustomer)
	}
	return nil
}

func (s *customerFeatureState) thenTheRetrievedCustomerIsAShallowCustomer(userID string) error {
	if s.lastErr != nil {
		return fmt.Errorf("expected query success, got %v", s.lastErr)
	}
	if s.lastCustomer.ID != userID {
		return fmt.Errorf("expected customer id %q, got %q", userID, s.lastCustomer.ID)
	}
	if !s.lastCustomer.Credentials.IsZero() || !s.lastCustomer.Billing.IsZero() || !s.lastCustomer.Shipping.IsZero() {
		return fmt.Errorf("expected shallow customer, got %#v", s.lastCustomer)
	}
	return nil
}

func (s *customerFeatureState) thenTheRetrievedCustomerContainsCredentialsAndBillingDetails(userID, name, email, phone, nipCode, company, city, address, postalCode, local string) error {
	if s.lastErr != nil {
		return fmt.Errorf("expected query success, got %v", s.lastErr)
	}
	if err := assertCustomerCredentials(s.lastCustomer, userID, name, email, phone); err != nil {
		return err
	}
	if err := assertCustomerBilling(s.lastCustomer, nipCode, company, city, address, postalCode, local); err != nil {
		return err
	}
	return nil
}

func (s *customerFeatureState) thenTheRetrievedCustomerContainsAFullProfile(userID, name, email, phone, nipCode, company, billingCity, billingAddress, billingPostalCode, billingLocal, shippingCity, shippingAddress, shippingPostalCode, shippingLocal string) error {
	if s.lastErr != nil {
		return fmt.Errorf("expected query success, got %v", s.lastErr)
	}
	if err := assertCustomerCredentials(s.lastCustomer, userID, name, email, phone); err != nil {
		return err
	}
	if err := assertCustomerBilling(s.lastCustomer, nipCode, company, billingCity, billingAddress, billingPostalCode, billingLocal); err != nil {
		return err
	}
	return assertCustomerShipping(s.lastCustomer, shippingCity, shippingAddress, shippingPostalCode, shippingLocal)
}

func (s *customerFeatureState) newProfile(userID, name, email, phone, nipCode, company, city, address, postalCode, local string) (customer.Customer, error) {
	credentials, err := customer.NewCredentials(name, email, phone)
	if err != nil {
		return customer.Customer{}, err
	}
	billing, err := customer.NewBilling(s.billingID(userID), nipCode, company, city, address, postalCode, local)
	if err != nil {
		return customer.Customer{}, err
	}
	return customer.NewCustomer(userID, credentials, billing, customer.ShippingAddress{}), nil
}

func (s *customerFeatureState) billingID(userID string) string {
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte("billing:"+userID)).String()
}

func (s *customerFeatureState) shippingID(userID string) string {
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte("shipping:"+userID)).String()
}

func assertCustomerCredentials(cust customer.Customer, userID, name, email, phone string) error {
	if cust.ID != userID {
		return fmt.Errorf("expected customer id %q, got %q", userID, cust.ID)
	}
	if cust.Credentials.Name != name || cust.Credentials.Email != email || cust.Credentials.Phone != phone {
		return fmt.Errorf("unexpected customer credentials: got %#v", cust.Credentials)
	}
	return nil
}

func assertCustomerBilling(cust customer.Customer, nipCode, company, city, address, postalCode, local string) error {
	if cust.Billing.IsZero() {
		return fmt.Errorf("expected billing details to be present")
	}
	if cust.Billing.NIPCode != nipCode || cust.Billing.Company != company {
		return fmt.Errorf("unexpected billing identity: got %#v", cust.Billing)
	}
	if cust.Billing.Address.City != city || cust.Billing.Address.Address != address || cust.Billing.Address.PostalCode != postalCode || cust.Billing.Address.Local != local {
		return fmt.Errorf("unexpected billing address: got %#v", cust.Billing.Address)
	}
	return nil
}

func assertCustomerShipping(cust customer.Customer, city, address, postalCode, local string) error {
	if cust.Shipping.IsZero() {
		return fmt.Errorf("expected shipping address to be present")
	}
	if cust.Shipping.Address.City != city || cust.Shipping.Address.Address != address || cust.Shipping.Address.PostalCode != postalCode || cust.Shipping.Address.Local != local {
		return fmt.Errorf("unexpected shipping address: got %#v", cust.Shipping.Address)
	}
	return nil
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	state := newCustomerFeatureState()

	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		return ctx, state.reset(ctx)
	})

	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		cleanupErr := state.cleanup()
		if err != nil {
			return ctx, err
		}
		return ctx, cleanupErr
	})

	ctx.Step(`^customer "([^"]*)" does not exist yet$`, state.givenCustomerDoesNotExistYet)
	ctx.Step(`^customer "([^"]*)" already exists as a shallow customer$`, state.givenCustomerAlreadyExistsAsAShallowCustomer)
	ctx.Step(`^customer "([^"]*)" has saved customer profile with credentials "([^"]*)" "([^"]*)" "([^"]*)" and billing details "([^"]*)" "([^"]*)" "([^"]*)" "([^"]*)" "([^"]*)" "([^"]*)"$`, state.givenCustomerHasSavedCustomerProfile)
	ctx.Step(`^customer "([^"]*)" has a shipping address "([^"]*)" "([^"]*)" "([^"]*)" "([^"]*)"$`, state.givenCustomerHasAShippingAddress)
	ctx.Step(`^logged-in customer "([^"]*)" is recognized for the first time$`, state.whenLoggedInCustomerIsRecognized)
	ctx.Step(`^basket customer "([^"]*)" adds an item for the first time$`, state.whenBasketCustomerAddsAnItemForTheFirstTime)
	ctx.Step(`^customer "([^"]*)" saves customer profile with credentials "([^"]*)" "([^"]*)" "([^"]*)" and billing details "([^"]*)" "([^"]*)" "([^"]*)" "([^"]*)" "([^"]*)" "([^"]*)"$`, state.whenCustomerSavesCustomerProfile)
	ctx.Step(`^customer "([^"]*)" adds shipping address "([^"]*)" "([^"]*)" "([^"]*)" "([^"]*)"$`, state.whenCustomerAddsShippingAddress)
	ctx.Step(`^the customer service retrieves customer "([^"]*)"$`, state.whenTheCustomerServiceRetrievesCustomer)
	ctx.Step(`^the customer command succeeds$`, state.thenTheCustomerCommandSucceeds)
	ctx.Step(`^the customer service rejects the request because the customer already exists$`, state.thenTheCustomerServiceRejectsBecauseTheCustomerAlreadyExists)
	ctx.Step(`^the customer service rejects the request because the customer is missing$`, state.thenTheCustomerServiceRejectsBecauseTheCustomerIsMissing)
	ctx.Step(`^customer "([^"]*)" can be retrieved as a shallow customer$`, state.thenCustomerCanBeRetrievedAsAShallowCustomer)
	ctx.Step(`^the customer profile for "([^"]*)" contains credentials "([^"]*)" "([^"]*)" "([^"]*)" and billing details "([^"]*)" "([^"]*)" "([^"]*)" "([^"]*)" "([^"]*)" "([^"]*)"$`, state.thenTheCustomerProfileContainsCredentialsAndBillingDetails)
	ctx.Step(`^the customer profile for "([^"]*)" contains shipping address "([^"]*)" "([^"]*)" "([^"]*)" "([^"]*)"$`, state.thenTheCustomerProfileContainsShippingAddress)
	ctx.Step(`^the customer "([^"]*)" is missing$`, state.thenTheCustomerIsMissing)
	ctx.Step(`^the retrieved customer "([^"]*)" is a shallow customer$`, state.thenTheRetrievedCustomerIsAShallowCustomer)
	ctx.Step(`^the retrieved customer "([^"]*)" contains credentials "([^"]*)" "([^"]*)" "([^"]*)" and billing details "([^"]*)" "([^"]*)" "([^"]*)" "([^"]*)" "([^"]*)" "([^"]*)"$`, state.thenTheRetrievedCustomerContainsCredentialsAndBillingDetails)
	ctx.Step(`^the retrieved customer "([^"]*)" contains credentials "([^"]*)" "([^"]*)" "([^"]*)" billing details "([^"]*)" "([^"]*)" "([^"]*)" "([^"]*)" "([^"]*)" "([^"]*)" and shipping address "([^"]*)" "([^"]*)" "([^"]*)" "([^"]*)"$`, state.thenTheRetrievedCustomerContainsAFullProfile)
}
