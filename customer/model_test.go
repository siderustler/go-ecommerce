package customer_test

import (
	"errors"
	"testing"

	"github.com/siderustler/go-ecommerce/customer"
)

func TestNewCredentials(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		nameInput    string
		emailInput   string
		phoneInput   string
		expected     customer.Credentials
		expectedErrs []error
	}{
		{
			name:       "valid credentials",
			nameInput:  "Alice",
			emailInput: "alice@example.com",
			phoneInput: "+48123",
			expected: customer.Credentials{
				Name:  "Alice",
				Email: "alice@example.com",
				Phone: "+48123",
			},
		},
		{
			name:       "missing name",
			nameInput:  "",
			emailInput: "alice@example.com",
			phoneInput: "+48123",
			expected: customer.Credentials{
				Name:  "",
				Email: "alice@example.com",
				Phone: "+48123",
			},
			expectedErrs: []error{customer.ErrNameEmpty},
		},
		{
			name:       "missing email",
			nameInput:  "Alice",
			emailInput: "",
			phoneInput: "+48123",
			expected: customer.Credentials{
				Name:  "Alice",
				Email: "",
				Phone: "+48123",
			},
			expectedErrs: []error{customer.ErrEmailEmpty},
		},
		{
			name:       "missing phone",
			nameInput:  "Alice",
			emailInput: "alice@example.com",
			phoneInput: "",
			expected: customer.Credentials{
				Name:  "Alice",
				Email: "alice@example.com",
				Phone: "",
			},
			expectedErrs: []error{customer.ErrPhoneEmpty},
		},
		{
			name:       "missing all fields",
			nameInput:  "",
			emailInput: "",
			phoneInput: "",
			expected:   customer.Credentials{},
			expectedErrs: []error{
				customer.ErrNameEmpty,
				customer.ErrEmailEmpty,
				customer.ErrPhoneEmpty,
			},
		},
		{
			name:       "blank fields",
			nameInput:  " ",
			emailInput: "   ",
			phoneInput: " ",
			expected: customer.Credentials{
				Name:  " ",
				Email: "   ",
				Phone: " ",
			},
			expectedErrs: []error{
				customer.ErrNameEmpty,
				customer.ErrEmailEmpty,
				customer.ErrPhoneEmpty,
			},
		},
		{
			name:       "values with surrounding spaces are preserved",
			nameInput:  " Alice ",
			emailInput: " alice@example.com ",
			phoneInput: " +48123 ",
			expected: customer.Credentials{
				Name:  " Alice ",
				Email: " alice@example.com ",
				Phone: " +48123 ",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			actual, err := customer.NewCredentials(test.nameInput, test.emailInput, test.phoneInput)

			if len(test.expectedErrs) == 0 && err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
			if len(test.expectedErrs) > 0 && err == nil {
				t.Fatal("expected error, got nil")
			}
			for _, expectedErr := range test.expectedErrs {
				if !errors.Is(err, expectedErr) {
					t.Fatalf("expected error %v to be present in tree: %v", expectedErr, err)
				}
			}
			if actual != test.expected {
				t.Fatalf("unexpected credentials: got %#v want %#v", actual, test.expected)
			}
		})
	}
}

func TestCredentialsIsZero(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    customer.Credentials
		expected bool
	}{
		{
			name: "all fields present",
			value: customer.Credentials{
				Name:  "Alice",
				Email: "alice@example.com",
				Phone: "+48123",
			},
			expected: false,
		},
		{
			name: "empty name",
			value: customer.Credentials{
				Email: "alice@example.com",
				Phone: "+48123",
			},
			expected: true,
		},
		{
			name: "empty email",
			value: customer.Credentials{
				Name:  "Alice",
				Phone: "+48123",
			},
			expected: true,
		},
		{
			name: "empty phone",
			value: customer.Credentials{
				Name:  "Alice",
				Email: "alice@example.com",
			},
			expected: true,
		},
		{
			name:     "all empty",
			value:    customer.Credentials{},
			expected: true,
		},
		{
			name: "blank fields",
			value: customer.Credentials{
				Name:  " ",
				Email: "   ",
				Phone: " ",
			},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if actual := test.value.IsZero(); actual != test.expected {
				t.Fatalf("unexpected IsZero result: got %t want %t", actual, test.expected)
			}
		})
	}
}

func TestNewShippingAddress(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		id           string
		city         string
		address      string
		postalCode   string
		local        string
		expected     customer.ShippingAddress
		expectedErrs []error
	}{
		{
			name:       "valid shipping address",
			id:         "shipping-1",
			city:       "Warsaw",
			address:    "Main",
			postalCode: "00-001",
			local:      "1A",
			expected: mustNewShippingAddress(
				t,
				"shipping-1",
				"Warsaw",
				"Main",
				"00-001",
				"1A",
			),
		},
		{
			name:       "missing city",
			id:         "shipping-1",
			city:       "",
			address:    "Main",
			postalCode: "00-001",
			local:      "1A",
			expected: mustShippingAddressPreservingInput(
				t,
				"shipping-1",
				"",
				"Main",
				"00-001",
				"1A",
			),
			expectedErrs: []error{customer.ErrCityEmpty},
		},
		{
			name:       "missing address",
			id:         "shipping-1",
			city:       "Warsaw",
			address:    "",
			postalCode: "00-001",
			local:      "1A",
			expected: mustShippingAddressPreservingInput(
				t,
				"shipping-1",
				"Warsaw",
				"",
				"00-001",
				"1A",
			),
			expectedErrs: []error{customer.ErrAddressEmpty},
		},
		{
			name:       "missing postal code",
			id:         "shipping-1",
			city:       "Warsaw",
			address:    "Main",
			postalCode: "",
			local:      "1A",
			expected: mustShippingAddressPreservingInput(
				t,
				"shipping-1",
				"Warsaw",
				"Main",
				"",
				"1A",
			),
			expectedErrs: []error{customer.ErrPostalCodeEmpty},
		},
		{
			name:       "missing local number",
			id:         "shipping-1",
			city:       "Warsaw",
			address:    "Main",
			postalCode: "00-001",
			local:      "",
			expected: mustShippingAddressPreservingInput(
				t,
				"shipping-1",
				"Warsaw",
				"Main",
				"00-001",
				"",
			),
			expectedErrs: []error{customer.ErrLocalNumberEmpty},
		},
		{
			name:       "all address fields missing",
			id:         "shipping-1",
			city:       "",
			address:    "",
			postalCode: "",
			local:      "",
			expected: mustShippingAddressPreservingInput(
				t,
				"shipping-1",
				"",
				"",
				"",
				"",
			),
			expectedErrs: []error{
				customer.ErrCityEmpty,
				customer.ErrAddressEmpty,
				customer.ErrPostalCodeEmpty,
				customer.ErrLocalNumberEmpty,
			},
		},
		{
			name:       "blank fields",
			id:         "shipping-1",
			city:       " ",
			address:    " ",
			postalCode: "   ",
			local:      " ",
			expected: mustShippingAddressPreservingInput(
				t,
				"shipping-1",
				" ",
				" ",
				"   ",
				" ",
			),
			expectedErrs: []error{
				customer.ErrCityEmpty,
				customer.ErrAddressEmpty,
				customer.ErrPostalCodeEmpty,
				customer.ErrLocalNumberEmpty,
			},
		},
		{
			name:       "empty spaces are trimmed",
			id:         "shipping-1",
			city:       " Warsaw ",
			address:    " Main ",
			postalCode: " 00-001 ",
			local:      " 1A ",
			expected:   mustNewShippingAddress(t, "shipping-1", "Warsaw", "Main", "00-001", "1A"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			actual, err := customer.NewShippingAddress(test.id, test.city, test.address, test.postalCode, test.local)

			if len(test.expectedErrs) == 0 && err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
			if len(test.expectedErrs) > 0 && err == nil {
				t.Fatal("expected error, got nil")
			}
			for _, expectedErr := range test.expectedErrs {
				if !errors.Is(err, expectedErr) {
					t.Fatalf("expected error %v to be present in tree: %v", expectedErr, err)
				}
			}
			if actual != test.expected {
				t.Fatalf("unexpected shipping address: got %#v want %#v", actual, test.expected)
			}
		})
	}
}

func TestShippingAddressIsZero(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    customer.ShippingAddress
		expected bool
	}{
		{
			name:     "id present",
			value:    mustNewShippingAddress(t, "shipping-1", "Warsaw", "Main", "00-001", "1A"),
			expected: false,
		},
		{
			name:     "empty id",
			value:    mustNewShippingAddress(t, "", "Warsaw", "Main", "00-001", "1A"),
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if actual := test.value.IsZero(); actual != test.expected {
				t.Fatalf("unexpected IsZero result: got %t want %t", actual, test.expected)
			}
		})
	}
}

func TestNewBilling(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		id           string
		nipCode      string
		company      string
		city         string
		address      string
		postalCode   string
		local        string
		expected     customer.Billing
		expectedErrs []error
	}{
		{
			name:       "valid billing",
			id:         "billing-1",
			nipCode:    "1234567890",
			company:    "Acme",
			city:       "Warsaw",
			address:    "Main",
			postalCode: "00-001",
			local:      "1A",
			expected: mustNewBilling(
				t,
				"billing-1",
				"1234567890",
				"Acme",
				"Warsaw",
				"Main",
				"00-001",
				"1A",
			),
		},
		{
			name:       "missing nip code",
			id:         "billing-1",
			nipCode:    "",
			company:    "Acme",
			city:       "Warsaw",
			address:    "Main",
			postalCode: "00-001",
			local:      "1A",
			expected: mustBillingPreservingInput(
				t,
				"billing-1",
				"",
				"Acme",
				"Warsaw",
				"Main",
				"00-001",
				"1A",
			),
			expectedErrs: []error{customer.ErrNipCodeEmpty},
		},
		{
			name:       "missing company",
			id:         "billing-1",
			nipCode:    "1234567890",
			company:    "",
			city:       "Warsaw",
			address:    "Main",
			postalCode: "00-001",
			local:      "1A",
			expected: mustBillingPreservingInput(
				t,
				"billing-1",
				"1234567890",
				"",
				"Warsaw",
				"Main",
				"00-001",
				"1A",
			),
			expectedErrs: []error{customer.ErrCompanyNameEmpty},
		},
		{
			name:       "missing city",
			id:         "billing-1",
			nipCode:    "1234567890",
			company:    "Acme",
			city:       "",
			address:    "Main",
			postalCode: "00-001",
			local:      "1A",
			expected: mustBillingPreservingInput(
				t,
				"billing-1",
				"1234567890",
				"Acme",
				"",
				"Main",
				"00-001",
				"1A",
			),
			expectedErrs: []error{customer.ErrCityEmpty},
		},
		{
			name:       "missing address",
			id:         "billing-1",
			nipCode:    "1234567890",
			company:    "Acme",
			city:       "Warsaw",
			address:    "",
			postalCode: "00-001",
			local:      "1A",
			expected: mustBillingPreservingInput(
				t,
				"billing-1",
				"1234567890",
				"Acme",
				"Warsaw",
				"",
				"00-001",
				"1A",
			),
			expectedErrs: []error{customer.ErrAddressEmpty},
		},
		{
			name:       "missing postal code",
			id:         "billing-1",
			nipCode:    "1234567890",
			company:    "Acme",
			city:       "Warsaw",
			address:    "Main",
			postalCode: "",
			local:      "1A",
			expected: mustBillingPreservingInput(
				t,
				"billing-1",
				"1234567890",
				"Acme",
				"Warsaw",
				"Main",
				"",
				"1A",
			),
			expectedErrs: []error{customer.ErrPostalCodeEmpty},
		},
		{
			name:       "missing local number",
			id:         "billing-1",
			nipCode:    "1234567890",
			company:    "Acme",
			city:       "Warsaw",
			address:    "Main",
			postalCode: "00-001",
			local:      "",
			expected: mustBillingPreservingInput(
				t,
				"billing-1",
				"1234567890",
				"Acme",
				"Warsaw",
				"Main",
				"00-001",
				"",
			),
			expectedErrs: []error{customer.ErrLocalNumberEmpty},
		},
		{
			name:       "joined nip and company errors",
			id:         "billing-1",
			nipCode:    "",
			company:    "",
			city:       "Warsaw",
			address:    "Main",
			postalCode: "00-001",
			local:      "1A",
			expected: mustBillingPreservingInput(
				t,
				"billing-1",
				"",
				"",
				"Warsaw",
				"Main",
				"00-001",
				"1A",
			),
			expectedErrs: []error{
				customer.ErrNipCodeEmpty,
				customer.ErrCompanyNameEmpty,
			},
		},
		{
			name:       "joined billing and address errors",
			id:         "billing-1",
			nipCode:    "",
			company:    "Acme",
			city:       "",
			address:    "Main",
			postalCode: "00-001",
			local:      "1A",
			expected: mustBillingPreservingInput(
				t,
				"billing-1",
				"",
				"Acme",
				"",
				"Main",
				"00-001",
				"1A",
			),
			expectedErrs: []error{
				customer.ErrNipCodeEmpty,
				customer.ErrCityEmpty,
			},
		},
		{
			name:       "all invalid",
			id:         "billing-1",
			nipCode:    " ",
			company:    " ",
			city:       " ",
			address:    " ",
			postalCode: " ",
			local:      " ",
			expected: mustBillingPreservingInput(
				t,
				"billing-1",
				" ",
				" ",
				" ",
				" ",
				" ",
				" ",
			),
			expectedErrs: []error{
				customer.ErrNipCodeEmpty,
				customer.ErrCompanyNameEmpty,
				customer.ErrCityEmpty,
				customer.ErrAddressEmpty,
				customer.ErrPostalCodeEmpty,
				customer.ErrLocalNumberEmpty,
			},
		},
		{
			name:       "blank fields are trimmed",
			id:         "billing-1",
			nipCode:    " 1234567890 ",
			company:    " Acme ",
			city:       " Warsaw ",
			address:    " Main ",
			postalCode: " 00-001 ",
			local:      " 1A ",
			expected:   mustNewBilling(t, "billing-1", "1234567890", "Acme", "Warsaw", "Main", "00-001", "1A"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			actual, err := customer.NewBilling(
				test.id,
				test.nipCode,
				test.company,
				test.city,
				test.address,
				test.postalCode,
				test.local,
			)

			if len(test.expectedErrs) == 0 && err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
			if len(test.expectedErrs) > 0 && err == nil {
				t.Fatal("expected error, got nil")
			}
			for _, expectedErr := range test.expectedErrs {
				if !errors.Is(err, expectedErr) {
					t.Fatalf("expected error %v to be present in tree: %v", expectedErr, err)
				}
			}
			if actual != test.expected {
				t.Fatalf("unexpected billing: got %#v want %#v", actual, test.expected)
			}
		})
	}
}

func TestBillingIsZero(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    customer.Billing
		expected bool
	}{
		{
			name:     "id present",
			value:    mustNewBilling(t, "billing-1", "1234567890", "Acme", "Warsaw", "Main", "00-001", "1A"),
			expected: false,
		},
		{
			name:     "empty id",
			value:    mustNewBilling(t, "", "1234567890", "Acme", "Warsaw", "Main", "00-001", "1A"),
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if actual := test.value.IsZero(); actual != test.expected {
				t.Fatalf("unexpected IsZero result: got %t want %t", actual, test.expected)
			}
		})
	}
}

func TestCustomerIsZero(t *testing.T) {
	t.Parallel()

	validCredentials, err := customer.NewCredentials("Alice", "alice@example.com", "+48123")
	if err != nil {
		t.Fatalf("creating valid credentials: %v", err)
	}
	validShipping := mustNewShippingAddress(t, "shipping-1", "Warsaw", "Main", "00-001", "1A")
	validBilling := mustNewBilling(t, "billing-1", "1234567890", "Acme", "Warsaw", "Main", "00-001", "1A")

	tests := []struct {
		name     string
		value    customer.Customer
		expected bool
	}{
		{
			name:     "all nested values valid",
			value:    customer.NewCustomer("customer-1", validCredentials, validBilling, validShipping),
			expected: false,
		},
		{
			name: "zero credentials",
			value: customer.NewCustomer(
				"customer-1",
				customer.Credentials{},
				validBilling,
				validShipping,
			),
			expected: true,
		},
		{
			name: "zero billing",
			value: customer.NewCustomer(
				"customer-1",
				validCredentials,
				customer.Billing{},
				validShipping,
			),
			expected: true,
		},
		{
			name: "zero shipping",
			value: customer.NewCustomer(
				"customer-1",
				validCredentials,
				validBilling,
				customer.ShippingAddress{},
			),
			expected: true,
		},
		{
			name: "multiple zero nested values",
			value: customer.NewCustomer(
				"customer-1",
				customer.Credentials{},
				customer.Billing{},
				validShipping,
			),
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if actual := test.value.IsZero(); actual != test.expected {
				t.Fatalf("unexpected IsZero result: got %t want %t", actual, test.expected)
			}
		})
	}
}

func mustNewShippingAddress(t *testing.T, id, city, address, postalCode, local string) customer.ShippingAddress {
	t.Helper()

	shipping, err := customer.NewShippingAddress(id, city, address, postalCode, local)
	if err != nil {
		t.Fatalf("creating shipping address: %v", err)
	}
	return shipping
}

func mustShippingAddressPreservingInput(t *testing.T, id, city, address, postalCode, local string) customer.ShippingAddress {
	t.Helper()

	shipping, err := customer.NewShippingAddress(id, city, address, postalCode, local)
	if err == nil {
		t.Fatalf("expected shipping validation error")
	}
	return shipping
}

func mustNewBilling(t *testing.T, id, nipCode, company, city, address, postalCode, local string) customer.Billing {
	t.Helper()

	billing, err := customer.NewBilling(id, nipCode, company, city, address, postalCode, local)
	if err != nil {
		t.Fatalf("creating billing: %v", err)
	}
	return billing
}

func mustBillingPreservingInput(t *testing.T, id, nipCode, company, city, address, postalCode, local string) customer.Billing {
	t.Helper()

	billing, err := customer.NewBilling(id, nipCode, company, city, address, postalCode, local)
	if err == nil {
		t.Fatalf("expected billing validation error")
	}
	return billing
}
