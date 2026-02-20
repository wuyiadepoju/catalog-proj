package services

import (
	"catalog-proj/internal/app/product/domain"
	"math/big"
	"testing"
	"time"
)

func TestPricingCalculator_CalculateEffectivePrice(t *testing.T) {
	calculator := NewPricingCalculator()
	now := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

	t.Run("returns base price when no discount", func(t *testing.T) {
		basePrice := domain.NewMoney(10000, "USD") // $100.00
		product := domain.NewProduct("product-1", "Test", "Desc", "Cat", basePrice)
		product.Activate()

		result := calculator.CalculateEffectivePrice(product, now)
		if result == nil {
			t.Fatal("CalculateEffectivePrice() returned nil")
		}
		if (*result).Cmp(*basePrice) != 0 {
			t.Errorf("CalculateEffectivePrice() = %s, want %s", (*result).String(), (*basePrice).String())
		}
	})

	t.Run("returns base price when discount is nil", func(t *testing.T) {
		basePrice := domain.NewMoney(10000, "USD")
		product := domain.NewProduct("product-2", "Test", "Desc", "Cat", basePrice)
		product.Activate()

		result := calculator.CalculateEffectivePrice(product, now)
		if result == nil {
			t.Fatal("CalculateEffectivePrice() returned nil")
		}
		if (*result).Cmp(*basePrice) != 0 {
			t.Errorf("CalculateEffectivePrice() = %s, want %s", (*result).String(), (*basePrice).String())
		}
	})

	t.Run("returns base price when discount is not valid", func(t *testing.T) {
		basePrice := domain.NewMoney(10000, "USD")
		product := domain.NewProduct("product-3", "Test", "Desc", "Cat", basePrice)
		product.Activate()

		invalidDiscount := &domain.Discount{
			ID:        "discount-1",
			Amount:    domain.NewMoney(10, 100), // 10%
			StartDate: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2024, 2, 28, 23, 59, 59, 0, time.UTC),
		}
		product.ApplyDiscount(invalidDiscount, time.Date(2024, 2, 15, 12, 0, 0, 0, time.UTC))

		result := calculator.CalculateEffectivePrice(product, now)
		if result == nil {
			t.Fatal("CalculateEffectivePrice() returned nil")
		}
		if (*result).Cmp(*basePrice) != 0 {
			t.Errorf("CalculateEffectivePrice() = %s, want %s (base price when discount invalid)", (*result).String(), (*basePrice).String())
		}
	})

	t.Run("applies 10% discount correctly", func(t *testing.T) {
		basePrice := domain.NewMoney(10000, "USD") // $100.00
		product := domain.NewProduct("product-4", "Test", "Desc", "Cat", basePrice)
		product.Activate()

		discount := &domain.Discount{
			ID:        "discount-2",
			Amount:    domain.NewMoney(10, 100), // 10% = 0.10
			StartDate: startDate,
			EndDate:   endDate,
		}
		product.ApplyDiscount(discount, now)

		result := calculator.CalculateEffectivePrice(product, now)
		if result == nil {
			t.Fatal("CalculateEffectivePrice() returned nil")
		}

		// Expected: $100.00 * (1 - 0.10) = $90.00
		expected := big.NewRat(9000, 100)
		if (*result).Cmp(expected) != 0 {
			t.Errorf("CalculateEffectivePrice() = %s, want %s", (*result).String(), expected.String())
		}
	})

	t.Run("applies 25% discount correctly", func(t *testing.T) {
		basePrice := domain.NewMoney(20000, "USD") // $200.00
		product := domain.NewProduct("product-5", "Test", "Desc", "Cat", basePrice)
		product.Activate()

		discount := &domain.Discount{
			ID:        "discount-3",
			Amount:    domain.NewMoney(25, 100), // 25% = 0.25
			StartDate: startDate,
			EndDate:   endDate,
		}
		product.ApplyDiscount(discount, now)

		result := calculator.CalculateEffectivePrice(product, now)
		if result == nil {
			t.Fatal("CalculateEffectivePrice() returned nil")
		}

		// Expected: $200.00 * (1 - 0.25) = $150.00
		expected := big.NewRat(15000, 100)
		if (*result).Cmp(expected) != 0 {
			t.Errorf("CalculateEffectivePrice() = %s, want %s", (*result).String(), expected.String())
		}
	})

	t.Run("applies 50% discount correctly", func(t *testing.T) {
		basePrice := domain.NewMoney(5000, "USD") // $50.00
		product := domain.NewProduct("product-6", "Test", "Desc", "Cat", basePrice)
		product.Activate()

		discount := &domain.Discount{
			ID:        "discount-4",
			Amount:    domain.NewMoney(50, 100), // 50% = 0.50
			StartDate: startDate,
			EndDate:   endDate,
		}
		product.ApplyDiscount(discount, now)

		result := calculator.CalculateEffectivePrice(product, now)
		if result == nil {
			t.Fatal("CalculateEffectivePrice() returned nil")
		}

		// Expected: $50.00 * (1 - 0.50) = $25.00
		expected := big.NewRat(2500, 100)
		if (*result).Cmp(expected) != 0 {
			t.Errorf("CalculateEffectivePrice() = %s, want %s", (*result).String(), expected.String())
		}
	})

	t.Run("returns nil when base price is nil", func(t *testing.T) {
		product := domain.NewProduct("product-7", "Test", "Desc", "Cat", nil)

		result := calculator.CalculateEffectivePrice(product, now)
		if result != nil {
			t.Errorf("CalculateEffectivePrice() = %v, want nil", result)
		}
	})

	t.Run("handles discount with zero amount", func(t *testing.T) {
		basePrice := domain.NewMoney(10000, "USD") // $100.00
		product := domain.NewProduct("product-8", "Test", "Desc", "Cat", basePrice)
		product.Activate()

		discount := &domain.Discount{
			ID:        "discount-5",
			Amount:    domain.NewMoney(0, 100), // 0% = 0.00
			StartDate: startDate,
			EndDate:   endDate,
		}
		product.ApplyDiscount(discount, now)

		result := calculator.CalculateEffectivePrice(product, now)
		if result == nil {
			t.Fatal("CalculateEffectivePrice() returned nil")
		}

		// Expected: $100.00 * (1 - 0.00) = $100.00
		expected := big.NewRat(10000, 100)
		if (*result).Cmp(expected) != 0 {
			t.Errorf("CalculateEffectivePrice() = %s, want %s", (*result).String(), expected.String())
		}
	})

}
