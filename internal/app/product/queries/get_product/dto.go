package get_product

import (
	"math/big"
	"time"
)

// DTO represents the data transfer object for a single product query result
type DTO struct {
	ID                string
	Name              string
	Description       string
	Category          string
	BasePrice         *big.Rat
	EffectivePrice    *big.Rat // Calculated price after discount
	DiscountID        *string
	DiscountAmount    *big.Rat
	DiscountStartDate *time.Time
	DiscountEndDate   *time.Time
	Status            string
	ArchivedAt        *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
