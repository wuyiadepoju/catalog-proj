package services

import (
	"catalog-proj/internal/app/product/domain"
	"time"
)

type PricingCalculator struct{}

func NewPricingCalculator() *PricingCalculator {
	return &PricingCalculator{}
}

// CalculateEffectivePrice calculates the effective price of a product
// If discount is valid at the given time, applies the percentage discount
// Otherwise returns the base price
func (pc *PricingCalculator) CalculateEffectivePrice(product *domain.Product, now time.Time) *domain.Money {
	basePrice := product.BasePrice()
	if basePrice == nil {
		return nil
	}

	discount := product.Discount()
	if discount == nil || discount.Amount == nil {
		// No discount, return base price
		return basePrice
	}

	// Check if discount is valid at the given time
	if !discount.IsValidAt(now) {
		// Discount not valid, return base price
		return basePrice
	}

	// Apply percentage discount: effectivePrice = basePrice * (1 - discountPercentage)
	// discount.Amount represents the percentage as a decimal (e.g., 0.10 = 10%)
	// basePrice is *Money (which is **big.Rat), so *basePrice is Money (*big.Rat)
	// discount.Amount is *Money (which is *big.Rat)
	basePriceValue := *basePrice
	discountPercentage := *discount.Amount

	// Calculate (1 - discountPercentage) using domain operations
	one := domain.NewMoney(100) // 1.00 as Money
	multiplier := domain.Subtract(one, discountPercentage)

	// Calculate basePrice * multiplier using domain operations
	effectivePrice := domain.Multiply(basePriceValue, multiplier)

	// Return as *domain.Money
	return &effectivePrice
}
