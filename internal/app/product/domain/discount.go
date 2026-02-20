package domain

import "time"

type Discount struct {
	ID        string
	Amount    *Money
	StartDate time.Time
	EndDate   time.Time
}

func (d *Discount) IsValidAt(now time.Time) bool {
	// Discount is valid if now is >= StartDate and < EndDate (inclusive start, exclusive end)
	return !now.Before(d.StartDate) && now.Before(d.EndDate)
}
