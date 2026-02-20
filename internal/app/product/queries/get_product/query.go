package get_product

import (
	"context"
	"fmt"
	"math/big"

	"catalog-proj/internal/app/product/domain"
	"catalog-proj/internal/app/product/domain/services"
	"catalog-proj/internal/pkg/clock"
)

// ReadModel defines the interface for reading products (to avoid import cycle)
type ReadModel interface {
	GetProduct(ctx context.Context, id string) (*DTO, error)
}

// Query handles the get product query use case
type Query struct {
	readModel  ReadModel
	calculator *services.PricingCalculator
	clock      clock.Clock
}

// NewQuery creates a new get product query
func NewQuery(
	readModel ReadModel,
	calculator *services.PricingCalculator,
	clock clock.Clock,
) *Query {
	return &Query{
		readModel:  readModel,
		calculator: calculator,
		clock:      clock,
	}
}

// Execute retrieves a product and calculates its effective price
func (q *Query) Execute(ctx context.Context, productID string) (*DTO, error) {
	// 1. Call read model
	dto, err := q.readModel.GetProduct(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	// 2. Calculate effective price using domain service
	// Reconstruct domain product to use the pricing calculator
	now := q.clock.Now()

	var basePrice *domain.Money
	if dto.BasePrice != nil {
		price := domain.Money(dto.BasePrice)
		basePrice = &price
	}

	// Reconstruct product from database data (queries should use ReconstructProduct, not NewProduct)
	var discount *domain.Discount
	if dto.DiscountID != nil && dto.DiscountStartDate != nil && dto.DiscountEndDate != nil {
		var discountAmount *domain.Money
		if dto.DiscountAmount != nil {
			amount := domain.Money(dto.DiscountAmount)
			discountAmount = &amount
		}

		discount = &domain.Discount{
			ID:        *dto.DiscountID,
			Amount:    discountAmount,
			StartDate: *dto.DiscountStartDate,
			EndDate:   *dto.DiscountEndDate,
		}
	}

	status := domain.ProductStatus(dto.Status)
	if status != domain.ProductStatusActive && status != domain.ProductStatusInactive {
		status = domain.ProductStatusInactive
	}

	product := domain.ReconstructProduct(
		dto.ID,
		dto.Name,
		dto.Description,
		dto.Category,
		basePrice,
		discount,
		status,
		dto.ArchivedAt,
		dto.CreatedAt,
		dto.UpdatedAt,
	)

	// Use pricing calculator
	effectivePricePtr := q.calculator.CalculateEffectivePrice(product, now)
	var effectivePrice *big.Rat
	if effectivePricePtr != nil {
		// effectivePricePtr is *domain.Money which is *big.Rat
		// domain.Money is *big.Rat, so *effectivePricePtr gives us *big.Rat
		effectivePrice = *effectivePricePtr
	} else if dto.BasePrice != nil {
		effectivePrice = dto.BasePrice
	}

	// Create response DTO with effective price
	// Build new DTO with all fields including calculated effective price
	return &DTO{
		ID:                dto.ID,
		Name:              dto.Name,
		Description:       dto.Description,
		Category:          dto.Category,
		BasePrice:         dto.BasePrice,
		EffectivePrice:    effectivePrice,
		DiscountID:        dto.DiscountID,
		DiscountAmount:    dto.DiscountAmount,
		DiscountStartDate: dto.DiscountStartDate,
		DiscountEndDate:   dto.DiscountEndDate,
		Status:            dto.Status,
		ArchivedAt:        dto.ArchivedAt,
		CreatedAt:         dto.CreatedAt,
		UpdatedAt:         dto.UpdatedAt,
	}, nil
}
