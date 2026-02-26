package product

import (
	"context"
	"strings"

	"catalog-proj/internal/app/product/usecases/create_product"
	pb "catalog-proj/proto/product/v1"
)

// CreateProduct handles the CreateProduct gRPC request
func (h *Handler) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.CreateProductResponse, error) {
	// 1. Validate
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, invalidArgumentError("name is required and cannot be empty")
	}
	if len(name) > 255 {
		return nil, invalidArgumentError("name exceeds maximum length of 255 characters")
	}

	description := strings.TrimSpace(req.Description)
	if description == "" {
		return nil, invalidArgumentError("description is required and cannot be empty")
	}
	if len(description) > 1000 {
		return nil, invalidArgumentError("description exceeds maximum length of 1000 characters")
	}

	category := strings.TrimSpace(req.Category)
	if category == "" {
		return nil, invalidArgumentError("category is required and cannot be empty")
	}
	if len(category) > 100 {
		return nil, invalidArgumentError("category exceeds maximum length of 100 characters")
	}

	if req.BasePrice == nil {
		return nil, invalidArgumentError("base_price is required")
	}
	if req.BasePrice.Amount <= 0 {
		return nil, invalidArgumentError("base_price must be positive")
	}

	// 2. Map proto to use case request
	basePrice := ProtoMoneyToDomain(req.BasePrice)
	useCaseReq := &create_product.Request{
		Name:        name,
		Description: description,
		Category:    category,
		BasePrice:   basePrice,
	}

	// 3. Call use case
	resp, err := h.createProductInteractor.Execute(ctx, useCaseReq)
	if err != nil {
		return nil, MapDomainError(err)
	}

	// 4. Map response to proto
	return &pb.CreateProductResponse{
		ProductId: resp.ProductID,
	}, nil
}
