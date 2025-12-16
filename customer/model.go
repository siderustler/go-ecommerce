package customer

type address struct {
	City       string
	Address    string
	PostalCode string
	Local      string
}

func newAddress(city, addr, postalCode, local string) address {
	return address{
		City:       city,
		Address:    addr,
		PostalCode: postalCode,
		Local:      local,
	}
}

type ShippingAddress struct {
	Address address
}

func NewShippingAddress(city, address, postalCode, local string) ShippingAddress {
	shipping := newAddress(city, address, postalCode, local)
	return ShippingAddress{
		Address: shipping,
	}
}

func (s ShippingAddress) isZero() bool {
	return s == ShippingAddress{}
}

type Billing struct {
	NIPCode string
	Company string
	Address address
}

func NewBilling(nipCode string, company string, city, address, postalCode, local string) Billing {
	billingAddress := newAddress(city, address, postalCode, local)
	return Billing{
		NIPCode: nipCode,
		Company: company,
		Address: billingAddress,
	}
}

type Credentials struct {
	Name  string
	Email string
	Phone string
}

func NewCredentials(name, email, phone string) Credentials {
	return Credentials{Name: name, Email: email, Phone: phone}
}

type Customer struct {
	ID          string
	Credentials Credentials
	Billing     Billing
	Shipping    ShippingAddress
}

func NewCustomer(id string, credentials Credentials, billing Billing, shipping ShippingAddress) Customer {
	return Customer{
		ID:          id,
		Credentials: credentials,
		Billing:     billing,
		Shipping:    shipping,
	}
}
