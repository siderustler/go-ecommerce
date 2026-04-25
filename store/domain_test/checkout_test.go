package store_domain_test

import (
	"errors"
	"testing"
	"time"

	store_domain "github.com/siderustler/go-ecommerce/store/domain"
)

var fixedCheckoutNow = time.Date(2026, time.April, 25, 10, 15, 0, 0, time.UTC)

func fixedCheckoutTimeNow() time.Time {
	return fixedCheckoutNow
}

func TestCheckoutStatusFromRawString(t *testing.T) {
	tests := []struct {
		name   string
		raw    string
		status store_domain.CheckoutStatus
	}{
		{name: "finalized", raw: "FINALIZED", status: store_domain.CheckoutFinalized},
		{name: "pending", raw: "PENDING", status: store_domain.CheckoutPending},
		{name: "invalidated", raw: "INVALIDATED", status: store_domain.CheckoutInvalidated},
		{name: "fallback invalidated", raw: "UNKNOWN", status: store_domain.CheckoutInvalidated},
	}

	for _, test := range tests {
		actual := store_domain.CheckoutStatusFromRawString(test.raw)
		if actual != test.status {
			t.Fatalf("test %s failed: expected status: %s actual status: %s", test.name, test.status, actual)
		}
	}
}

func TestCheckoutIsExpired(t *testing.T) {
	tests := []struct {
		name      string
		createdAt string
		expired   bool
	}{
		{name: "invalid timestamp is expired", createdAt: "broken", expired: true},
		{name: "older than 15 minutes is expired", createdAt: fixedCheckoutNow.Add(-16 * time.Minute).Format(time.RFC3339), expired: true},
		{name: "exactly 15 minutes old is not expired", createdAt: fixedCheckoutNow.Add(-15 * time.Minute).Format(time.RFC3339), expired: false},
		{name: "recent checkout is not expired", createdAt: fixedCheckoutNow.Add(-5 * time.Minute).Format(time.RFC3339), expired: false},
	}

	for _, test := range tests {
		actual := store_domain.NewCheckout(
			"1",
			"u1",
			map[string]store_domain.CartProduct{},
			test.createdAt,
			store_domain.CheckoutPending,
			fixedCheckoutTimeNow,
		)
		isExpired := actual.IsExpired()
		if isExpired != test.expired {
			t.Fatalf("test %s failed: expected expired: %t actual expired: %t", test.name, test.expired, isExpired)
		}
	}
}

func TestCheckoutInvalidate(t *testing.T) {
	tests := []struct {
		name          string
		entity        store_domain.Checkout
		expectedError error
		expectedState store_domain.CheckoutStatus
	}{
		{
			name:          "invalidate pending checkout",
			entity:        store_domain.NewCheckout("1", "u1", map[string]store_domain.CartProduct{}, time.Now().UTC().Format(time.RFC3339), store_domain.CheckoutPending),
			expectedState: store_domain.CheckoutInvalidated,
		},
		{
			name:          "invalidating invalidated checkout returns error",
			entity:        store_domain.NewCheckout("1", "u1", map[string]store_domain.CartProduct{}, time.Now().UTC().Format(time.RFC3339), store_domain.CheckoutInvalidated),
			expectedError: store_domain.ErrCheckoutInvalidated,
			expectedState: store_domain.CheckoutInvalidated,
		},
		{
			name:          "invalidating finalized checkout transitions to invalidated",
			entity:        store_domain.NewCheckout("1", "u1", map[string]store_domain.CartProduct{}, time.Now().UTC().Format(time.RFC3339), store_domain.CheckoutFinalized),
			expectedState: store_domain.CheckoutInvalidated,
		},
	}

	for _, test := range tests {
		actual := test.entity
		err := actual.Invalidate()
		if !errors.Is(err, test.expectedError) {
			t.Fatalf("test %s failed: expected error: %v actual error: %v", test.name, test.expectedError, err)
		}
		if actual.Status != test.expectedState {
			t.Fatalf("test %s failed: expected status: %s actual status: %s", test.name, test.expectedState, actual.Status)
		}
	}
}

func TestCheckoutFinalize(t *testing.T) {
	actual := store_domain.NewCheckout("1", "u1", map[string]store_domain.CartProduct{}, time.Now().UTC().Format(time.RFC3339), store_domain.CheckoutPending)
	actual.Finalize()
	if actual.Status != store_domain.CheckoutFinalized {
		t.Fatalf("test finalize checkout failed: expected status: %s actual status: %s", store_domain.CheckoutFinalized, actual.Status)
	}
}
