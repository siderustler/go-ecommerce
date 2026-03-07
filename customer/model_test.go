package customer_test

import (
	"errors"
	"testing"

	"github.com/siderustler/go-ecommerce/customer"
)

func TestNewCredentials(t *testing.T) {
	tests := []struct {
		name string
		in   customer.Credentials
		err  error
	}{
		{
			name: "valid credentials",
			in:   customer.Credentials{Name: "A", Email: "a@example.com", Phone: "+48123"},
		},
		{
			name: "missing all fields",
			in:   customer.Credentials{},
			err:  errors.Join(customer.ErrNameEmpty, customer.ErrEmailEmpty, customer.ErrPhoneEmpty),
		},
	}

	for _, test := range tests {
		actual, err := customer.NewCredentials(test.in.Name, test.in.Email, test.in.Phone)
		if test.err == nil && err != nil {
			t.Fatalf("test %s failed: expected no error, got: %v", test.name, err)
		}
		if test.err != nil {
			if !errors.Is(err, customer.ErrNameEmpty) || !errors.Is(err, customer.ErrEmailEmpty) || !errors.Is(err, customer.ErrPhoneEmpty) {
				t.Fatalf("test %s failed: expected joined validation errors, got: %v", test.name, err)
			}
		}
		if actual.Name != test.in.Name || actual.Email != test.in.Email || actual.Phone != test.in.Phone {
			t.Fatalf("test %s failed: unexpected credentials mapping", test.name)
		}
	}
}

func TestNewShippingAddress(t *testing.T) {
	tests := []struct {
		name string
		city string
		err  bool
	}{
		{name: "valid shipping", city: "Warsaw", err: false},
		{name: "empty city", city: "", err: true},
	}

	for _, test := range tests {
		_, err := customer.NewShippingAddress("1", test.city, "Main", "00-001", "1")
		isError := err != nil
		if isError != test.err {
			t.Fatalf("test %s failed: expected error: %t actual error: %t", test.name, test.err, isError)
		}
	}
}
