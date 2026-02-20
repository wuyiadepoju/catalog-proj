package m_product

import (
	"math/big"
	"time"

	"cloud.google.com/go/spanner"
)

// Product represents the database model for products
type Product struct {
	ProductID            string     `spanner:"product_id"`
	Name                 string     `spanner:"name"`
	Description          string     `spanner:"description"`
	Category             string     `spanner:"category"`
	BasePriceNumerator   int64      `spanner:"base_price_numerator"`
	BasePriceDenominator int64      `spanner:"base_price_denominator"`
	DiscountID           *string    `spanner:"discount_id"`
	DiscountAmount       *big.Rat   `spanner:"discount_amount"` // Stored as NUMERIC in Spanner
	DiscountStartDate    *time.Time `spanner:"discount_start_date"`
	DiscountEndDate      *time.Time `spanner:"discount_end_date"`
	Status               string     `spanner:"status"`
	ArchivedAt           *time.Time `spanner:"archived_at"`
	CreatedAt            time.Time  `spanner:"created_at"`
	UpdatedAt            time.Time  `spanner:"updated_at"`
}

// InsertMut creates a Spanner insert mutation for a product
func (p *Product) InsertMut() *spanner.Mutation {
	return spanner.Insert(
		TableName,
		[]string{
			ProductID, Name, Description, Category, BasePriceNumerator, BasePriceDenominator,
			DiscountID, DiscountAmount, DiscountStartDate, DiscountEndDate,
			Status, ArchivedAt, CreatedAt, UpdatedAt,
		},
		[]interface{}{
			p.ProductID, p.Name, p.Description, p.Category, p.BasePriceNumerator, p.BasePriceDenominator,
			p.DiscountID, p.DiscountAmount, p.DiscountStartDate, p.DiscountEndDate,
			p.Status, p.ArchivedAt, p.CreatedAt, p.UpdatedAt,
		},
	)
}

// UpdateMut creates a Spanner update mutation for a product
// Only updates fields that are provided (non-nil for optional fields)
// Note: columns must include ProductID as the first column (primary key)
func (p *Product) UpdateMut(columns []string) *spanner.Mutation {
	values := make([]interface{}, 0, len(columns))
	for _, col := range columns {
		switch col {
		case ProductID:
			values = append(values, p.ProductID)
		case Name:
			values = append(values, p.Name)
		case Description:
			values = append(values, p.Description)
		case Category:
			values = append(values, p.Category)
		case BasePriceNumerator:
			values = append(values, p.BasePriceNumerator)
		case BasePriceDenominator:
			values = append(values, p.BasePriceDenominator)
		case DiscountID:
			values = append(values, p.DiscountID)
		case DiscountAmount:
			values = append(values, p.DiscountAmount)
		case DiscountStartDate:
			values = append(values, p.DiscountStartDate)
		case DiscountEndDate:
			values = append(values, p.DiscountEndDate)
		case Status:
			values = append(values, p.Status)
		case ArchivedAt:
			values = append(values, p.ArchivedAt)
		case UpdatedAt:
			values = append(values, p.UpdatedAt)
		}
	}

	return spanner.Update(
		TableName,
		columns,
		values,
	)
}

// DeleteMut creates a Spanner delete mutation for a product
func (p *Product) DeleteMut() *spanner.Mutation {
	return spanner.Delete(TableName, spanner.Key{p.ProductID})
}

// TableName is the Spanner table name for products
const TableName = "products"

// AllColumns returns all column names for the products table
func AllColumns() []string {
	return []string{
		ProductID, Name, Description, Category, BasePriceNumerator, BasePriceDenominator,
		DiscountID, DiscountAmount, DiscountStartDate, DiscountEndDate,
		Status, ArchivedAt, CreatedAt, UpdatedAt,
	}
}
