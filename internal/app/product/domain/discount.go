package domain

import (
	"math/big"
	"strings"
	"time"
)

type Discount struct {
	ID        string
	Amount    *Money
	StartDate time.Time
	EndDate   time.Time
}

// Validate validates the discount structure
func (d *Discount) Validate() error {
	if strings.TrimSpace(d.ID) == "" {
		return ErrInvalidDiscountID
	}

	if d.Amount == nil {
		return ErrInvalidDiscountAmount
	}

	// Validate discount amount is between 0 and 100% (0.00 to 1.00)
	// Amount is stored as a decimal (e.g., 0.20 = 20%)
	// d.Amount is *Money, and Money is *big.Rat, so *d.Amount is Money (*big.Rat)
	// We need to cast to *big.Rat to use Cmp method
	amount := (*big.Rat)(*d.Amount)
	zero := big.NewRat(0, 1)
	one := big.NewRat(1, 1)

	// amount is *big.Rat, so we can call Cmp directly
	if amount.Cmp(zero) < 0 || amount.Cmp(one) > 0 {
		return ErrInvalidDiscountAmount
	}

	if !d.StartDate.Before(d.EndDate) {
		return ErrInvalidDiscountDateRange
	}

	return nil
}

func (d *Discount) IsValidAt(now time.Time) bool {
	// Discount is valid if now is >= StartDate and < EndDate (inclusive start, exclusive end)
	return !now.Before(d.StartDate) && now.Before(d.EndDate)
}
