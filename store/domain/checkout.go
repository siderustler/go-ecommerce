package store_domain

import (
	"errors"
	"time"
)

type checkoutTimeNowProvider func() time.Time

var checkoutTimeNowProviderDefault checkoutTimeNowProvider = func() time.Time {
	return time.Now().UTC()
}

var ErrCheckoutInvalidated = errors.New("checkout already invalidated")

type CheckoutStatus string

var (
	CheckoutFinalized   = CheckoutStatus("FINALIZED")
	CheckoutPending     = CheckoutStatus("PENDING")
	CheckoutInvalidated = CheckoutStatus("INVALIDATED")
)

func CheckoutStatusFromRawString(stat string) CheckoutStatus {
	switch stat {
	case string(CheckoutFinalized):
		return CheckoutFinalized
	case string(CheckoutPending):
		return CheckoutPending
	default:
		return CheckoutInvalidated
	}
}

type Checkout struct {
	ID        string
	UserID    string
	Items     map[string]CartProduct
	CreatedAt string
	Status    CheckoutStatus
	timeNow   checkoutTimeNowProvider
}

func NewCheckout(id, userID string, items map[string]CartProduct, createdAt string, status CheckoutStatus, timers ...checkoutTimeNowProvider) Checkout {
	t := checkoutTimeNowProviderDefault
	if len(timers) > 0 {
		t = timers[0]
	}
	return Checkout{
		ID:        id,
		UserID:    userID,
		Items:     items,
		Status:    status,
		CreatedAt: createdAt,
		timeNow:   t,
	}
}

func (c Checkout) IsExpired() bool {
	parsedTime, err := time.Parse(time.RFC3339, c.CreatedAt)
	if err != nil {
		return true
	}
	expiryTime := 15 * time.Minute
	expiredTime := parsedTime.Add(expiryTime)
	return c.timeNow().After(expiredTime)
}

func (c Checkout) IsZero() bool {
	return c.ID == ""
}

func (c *Checkout) Invalidate() error {
	if c.Status == CheckoutInvalidated {
		return ErrCheckoutInvalidated
	}
	c.Status = CheckoutInvalidated
	return nil
}

func (c *Checkout) Finalize() {
	c.Status = CheckoutFinalized
}
