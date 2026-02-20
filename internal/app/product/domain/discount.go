package domain

import "time"

type Discount struct {
	ID        string
	Amount    *Money
	StartDate time.Time
	EndDate   time.Time
}

func (d *Discount) IsValidAt(now time.Time) bool {
	return now.After(d.StartDate) && now.Before(d.EndDate)
}
