package list_products

import (
	"context"
	"fmt"

	"catalog-proj/internal/app/product/domain"
	"catalog-proj/internal/app/product/domain/services"
	"catalog-proj/internal/pkg/clock"
)

// ReadModel defines the interface for reading products (to avoid import cycle)
type ReadModel interface {
	ListProducts(ctx context.Context, req *Request) (*DTO, error)
}

// Query handles the list products query use case
type Query struct {
	readModel  ReadModel
	calculator *services.PricingCalculator
	clock      clock.Clock
}

// NewQuery creates a new list products query
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

// Execute retrieves a list of products and calculates effective prices
func (q *Query) Execute(ctx context.Context, req *Request) (*DTO, error) {
	// 1. Call read model with filters
	dto, err := q.readModel.ListProducts(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}

	// 2. Calculate effective prices for each product
	now := q.clock.Now()
	for i := range dto.Products {
		product := &dto.Products[i]
		
		// Reconstruct domain product from database data (queries should use ReconstructProduct, not NewProduct)
		var basePrice *domain.Money
		if product.BasePrice != nil {
			price := domain.Money(product.BasePrice)
			basePrice = &price
		}
		
		var discount *domain.Discount
		if product.DiscountID != nil && product.DiscountStartDate != nil && product.DiscountEndDate != nil {
			var discountAmount *domain.Money
			if product.DiscountAmount != nil {
				amount := domain.Money(product.DiscountAmount)
				discountAmount = &amount
			}
			
			discount = &domain.Discount{
				ID:        *product.DiscountID,
				Amount:    discountAmount,
				StartDate: *product.DiscountStartDate,
				EndDate:   *product.DiscountEndDate,
			}
		}
		
		status := domain.ProductStatus(product.Status)
		if status != domain.ProductStatusActive && status != domain.ProductStatusInactive {
			status = domain.ProductStatusInactive
		}
		
		domainProduct := domain.ReconstructProduct(
			product.ID,
			product.Name,
			product.Description,
			product.Category,
			basePrice,
			discount,
			status,
			product.ArchivedAt,
			product.CreatedAt,
			product.UpdatedAt,
		)
		
		// Calculate effective price
		effectivePricePtr := q.calculator.CalculateEffectivePrice(domainProduct, now)
		if effectivePricePtr != nil {
			product.EffectivePrice = *effectivePricePtr
		} else if product.BasePrice != nil {
			product.EffectivePrice = product.BasePrice
		}
	}

	// 3. Return paginated DTO
	return dto, nil
}
