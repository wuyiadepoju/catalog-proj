package contracts

import (
	"context"

	"catalog-proj/internal/app/product/domain"
	"cloud.google.com/go/spanner"
)

// ProductRepository defines the interface for product persistence operations
type ProductRepository interface {
	// InsertMut creates a Spanner insert mutation for a new product
	InsertMut(product *domain.Product) *spanner.Mutation

	// UpdateMut creates a Spanner update mutation for an existing product
	// Uses the product's change tracker to build targeted updates
	UpdateMut(product *domain.Product) *spanner.Mutation

	// Load retrieves a product by ID from Spanner and maps it to domain model
	Load(ctx context.Context, id string) (*domain.Product, error)
}
