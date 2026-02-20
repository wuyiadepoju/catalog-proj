package repo

import (
	"context"
	"fmt"
	"math/big"

	"catalog-proj/internal/app/product/domain"
	"catalog-proj/internal/models/m_product"

	"cloud.google.com/go/spanner"
	"google.golang.org/grpc/codes"
)

// SpannerProductRepository implements ProductRepository using Spanner
type SpannerProductRepository struct {
	client *spanner.Client
}

// NewSpannerProductRepository creates a new Spanner product repository
func NewSpannerProductRepository(client *spanner.Client) *SpannerProductRepository {
	return &SpannerProductRepository{
		client: client,
	}
}

// InsertMut creates a Spanner insert mutation for a new product
func (r *SpannerProductRepository) InsertMut(product *domain.Product) *spanner.Mutation {
	model := r.domainToModel(product)
	return model.InsertMut()
}

// UpdateMut creates a Spanner update mutation using the product's change tracker
func (r *SpannerProductRepository) UpdateMut(product *domain.Product) *spanner.Mutation {
	model := r.domainToModel(product)
	changes := product.Changes()

	// Build columns list based on dirty fields
	columns := []string{"product_id"} // Always include product_id as primary key

	if changes.Dirty(domain.FieldName) {
		columns = append(columns, "name")
	}
	if changes.Dirty(domain.FieldDescription) {
		columns = append(columns, "description")
	}
	if changes.Dirty(domain.FieldCategory) {
		columns = append(columns, "category")
	}
	// Base price changes require both numerator and denominator
	if changes.Dirty("base_price") {
		columns = append(columns, "base_price_numerator", "base_price_denominator")
	}
	if changes.Dirty(domain.FieldDiscount) {
		// When discount changes, update all discount-related fields
		columns = append(columns, "discount_id", "discount_amount", "discount_start_date", "discount_end_date")
	}
	if changes.Dirty(domain.FieldStatus) {
		columns = append(columns, "status")
	}
	if changes.Dirty(domain.FieldArchivedAt) {
		columns = append(columns, "archived_at")
	}
	// Always update UpdatedAt
	columns = append(columns, "updated_at")

	return model.UpdateMut(columns)
}

// Load retrieves a product by ID from Spanner and maps it to domain model
func (r *SpannerProductRepository) Load(ctx context.Context, id string) (*domain.Product, error) {
	columns := m_product.AllColumns()
	row, err := r.client.Single().ReadRow(ctx, m_product.TableName, spanner.Key{id}, columns)
	if err != nil {
		// Check if error is "not found" - Spanner returns codes.NotFound
		if spanner.ErrCode(err) == codes.NotFound {
			return nil, domain.ErrProductNotFound
		}
		return nil, fmt.Errorf("failed to load product: %w", err)
	}

	model := &m_product.Product{}
	if err := row.ToStruct(model); err != nil {
		return nil, fmt.Errorf("failed to parse product row: %w", err)
	}

	return r.modelToDomain(model)
}

// domainToModel converts a domain Product to a database model
func (r *SpannerProductRepository) domainToModel(product *domain.Product) *m_product.Product {
	model := &m_product.Product{
		ProductID:   product.ID(),
		Name:        product.Name(),
		Description: product.Description(),
		Category:    product.Category(),
		Status:      string(product.Status()),
		CreatedAt:   product.CreatedAt(),
		UpdatedAt:   product.UpdatedAt(),
	}

	// Convert base price: domain.Money is *big.Rat, convert to numerator/denominator
	if basePrice := product.BasePrice(); basePrice != nil {
		// basePrice is *domain.Money
		// domain.Money is *big.Rat, so *basePrice gives us Money (which is *big.Rat)
		// We need to convert to *big.Rat to call Num() and Denom()
		money := *basePrice      // This gives us domain.Money which is *big.Rat
		rat := (*big.Rat)(money) // Convert domain.Money to *big.Rat
		model.BasePriceNumerator = rat.Num().Int64()
		model.BasePriceDenominator = rat.Denom().Int64()
	} else {
		// Default to 0/1 if no price
		model.BasePriceNumerator = 0
		model.BasePriceDenominator = 1
	}

	// Convert discount
	if discount := product.Discount(); discount != nil {
		model.DiscountID = &discount.ID
		if discount.Amount != nil {
			// discount.Amount is *domain.Money
			// domain.Money is *big.Rat, so *discount.Amount is Money which is *big.Rat
			money := *discount.Amount
			rat := (*big.Rat)(money) // Convert domain.Money to *big.Rat
			model.DiscountAmount = rat
		}
		model.DiscountStartDate = &discount.StartDate
		model.DiscountEndDate = &discount.EndDate
	}

	// Handle archivedAt
	if archivedAt := product.ArchivedAt(); archivedAt != nil {
		model.ArchivedAt = archivedAt
	}

	return model
}

// modelToDomain converts a database model to a domain Product
func (r *SpannerProductRepository) modelToDomain(model *m_product.Product) (*domain.Product, error) {
	// Convert base price: numerator/denominator to *big.Rat
	var basePrice *domain.Money
	if model.BasePriceDenominator != 0 {
		price := domain.NewMoneyFromFraction(model.BasePriceNumerator, model.BasePriceDenominator)
		basePrice = &price
	}

	// Convert discount
	var discount *domain.Discount
	if model.DiscountID != nil && model.DiscountStartDate != nil && model.DiscountEndDate != nil {
		var discountAmount *domain.Money
		if model.DiscountAmount != nil {
			amount := domain.Money(model.DiscountAmount)
			discountAmount = &amount
		}

		discount = &domain.Discount{
			ID:        *model.DiscountID,
			Amount:    discountAmount,
			StartDate: *model.DiscountStartDate,
			EndDate:   *model.DiscountEndDate,
		}
	}

	// Convert status
	status := domain.ProductStatus(model.Status)
	if status != domain.ProductStatusActive && status != domain.ProductStatusInactive {
		status = domain.ProductStatusInactive // Default to inactive if invalid
	}

	// Reconstruct product using factory method
	product := domain.ReconstructProduct(
		model.ProductID,
		model.Name,
		model.Description,
		model.Category,
		basePrice,
		discount,
		status,
		model.ArchivedAt,
		model.CreatedAt,
		model.UpdatedAt,
	)

	return product, nil
}
