package product

import (
	"context"
	"strings"

	"catalog-proj/internal/app/product/usecases/update_product"
	pb "catalog-proj/proto/product/v1"
)

// UpdateProduct handles the UpdateProduct gRPC request
func (h *Handler) UpdateProduct(ctx context.Context, req *pb.UpdateProductRequest) (*pb.UpdateProductResponse, error) {
	// 1. Validate
	if req.ProductId == "" {
		return nil, invalidArgumentError("product_id is required")
	}

	// Validate that at least one field is being updated
	if req.Name == nil && req.Description == nil && req.Category == nil {
		return nil, invalidArgumentError("at least one field (name, description, or category) must be provided")
	}

	// 2. Map proto to use case request
	useCaseReq := &update_product.Request{
		ProductID: req.ProductId,
	}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, invalidArgumentError("name cannot be empty")
		}
		if len(name) > 255 {
			return nil, invalidArgumentError("name exceeds maximum length of 255 characters")
		}
		useCaseReq.Name = &name
	}
	if req.Description != nil {
		description := strings.TrimSpace(*req.Description)
		if description == "" {
			return nil, invalidArgumentError("description cannot be empty")
		}
		if len(description) > 1000 {
			return nil, invalidArgumentError("description exceeds maximum length of 1000 characters")
		}
		useCaseReq.Description = &description
	}
	if req.Category != nil {
		category := strings.TrimSpace(*req.Category)
		if category == "" {
			return nil, invalidArgumentError("category cannot be empty")
		}
		if len(category) > 100 {
			return nil, invalidArgumentError("category exceeds maximum length of 100 characters")
		}
		useCaseReq.Category = &category
	}

	// 3. Call use case
	resp, err := h.updateProductInteractor.Execute(ctx, useCaseReq)
	if err != nil {
		return nil, MapDomainError(err)
	}

	// 4. Map response to proto
	return &pb.UpdateProductResponse{
		ProductId: resp.ProductID,
	}, nil
}
