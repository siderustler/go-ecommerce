package customer

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrCityEmpty        = errors.New("city is empty")
	ErrAddressEmpty     = errors.New("address is empty")
	ErrPostalCodeEmpty  = errors.New("postal code is empty")
	ErrLocalNumberEmpty = errors.New("local number is empty")
	ErrNipCodeEmpty     = errors.New("nip code is empty")
	ErrCompanyNameEmpty = errors.New("company name is empty")
	ErrNameEmpty        = errors.New("name is empty")
	ErrPhoneEmpty       = errors.New("phone is empty")
	ErrEmailEmpty       = errors.New("email is empty")
)

type address struct {
	City       string
	Address    string
	PostalCode string
	Local      string
}

func (a address) isZero() bool {
	return a.City == "" || a.Address == "" || a.PostalCode == "" || a.Local == ""
}

func newAddress(city, addr, postalCode, local string) (address, error) {
	var err error
	if strings.Trim(city, " ") == "" {
		err = errors.Join(err, ErrCityEmpty)
	}
	if strings.Trim(addr, " ") == "" {
		err = errors.Join(err, ErrAddressEmpty)
	}
	if strings.Trim(postalCode, " ") == "" {
		err = errors.Join(err, ErrPostalCodeEmpty)
	}
	if strings.Trim(local, " ") == "" {
		err = errors.Join(err, ErrLocalNumberEmpty)
	}
	return address{
		City:       city,
		Address:    addr,
		PostalCode: postalCode,
		Local:      local,
	}, err
}

type ShippingAddress struct {
	ID      string
	Address address
}

func NewShippingAddress(id, city, address, postalCode, local string) (ShippingAddress, error) {
	shipping, err := newAddress(city, address, postalCode, local)
	if err != nil {
		err = fmt.Errorf("validating shipping address: %w", err)
	}
	return ShippingAddress{
		ID:      id,
		Address: shipping,
	}, err
}

func (s ShippingAddress) IsZero() bool {
	return s.ID == ""
}

type Billing struct {
	ID      string
	NIPCode string
	Company string
	Address address
}

func NewBilling(id, nipCode, company, city, address, postalCode, local string) (Billing, error) {
	billingAddress, err := newAddress(city, address, postalCode, local)
	if err != nil {
		err = fmt.Errorf("validating billing address: %w", err)
	}
	if strings.Trim(nipCode, " ") == "" {
		err = errors.Join(err, ErrNipCodeEmpty)
	}
	if strings.Trim(company, " ") == "" {
		err = errors.Join(err, ErrCompanyNameEmpty)
	}
	return Billing{
		ID:      id,
		NIPCode: nipCode,
		Company: company,
		Address: billingAddress,
	}, err
}

type Credentials struct {
	Name  string
	Email string
	Phone string
}

func NewCredentials(name, email, phone string) (Credentials, error) {
	var err error
	if strings.Trim(name, " ") == "" {
		err = errors.Join(err, ErrNameEmpty)
	}
	if strings.Trim(email, " ") == "" {
		err = errors.Join(err, ErrEmailEmpty)
	}
	if strings.Trim(phone, " ") == "" {
		err = errors.Join(err, ErrPhoneEmpty)
	}
	return Credentials{Name: name, Email: email, Phone: phone}, err
}

type Customer struct {
	ID          string
	Credentials Credentials
	Billing     Billing
	Shipping    ShippingAddress
}

func (c Customer) IsZero() bool {
	return c.Billing.IsZero() || c.Shipping.IsZero() || c.Credentials.IsZero()
}

func (c Credentials) IsZero() bool {
	return c.Email == "" || c.Name == "" || c.Phone == ""
}

func (b Billing) IsZero() bool {
	return b.ID == ""
}

func NewCustomer(id string, credentials Credentials, billing Billing, shipping ShippingAddress) Customer {
	return Customer{
		ID:          id,
		Credentials: credentials,
		Billing:     billing,
		Shipping:    shipping,
	}
}
