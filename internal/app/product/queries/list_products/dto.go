package list_products

import (
	"math/big"
	"time"
)

// Request represents the request parameters for listing products
type Request struct {
	Category string
	Status   string
	Limit    int
	Offset   int
}

// ProductItem represents a single product in the list
type ProductItem struct {
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

// DTO represents the data transfer object for list products query result
type DTO struct {
	Products []ProductItem
	Total    int
}
