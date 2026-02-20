package contracts

import (
	"context"

	"catalog-proj/internal/app/product/queries/get_product"
)

// ReadModel defines the interface for read-only product queries
// These queries bypass the domain layer and return DTOs directly
// Note: ListProducts has its own ReadModel interface in list_products package to avoid import cycles
type ReadModel interface {
	// GetProduct retrieves a single product by ID
	GetProduct(ctx context.Context, id string) (*get_product.DTO, error)
}
