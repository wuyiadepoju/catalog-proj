package product

import (
	"context"

	"catalog-proj/internal/app/product/usecases/update_product"
	pb "catalog-proj/proto/product/v1"
)

// UpdateProduct handles the UpdateProduct gRPC request
func (h *Handler) UpdateProduct(ctx context.Context, req *pb.UpdateProductRequest) (*pb.UpdateProductResponse, error) {
	// 1. Validate
	if req.ProductId == "" {
		return nil, invalidArgumentError("product_id is required")
	}

	// 2. Map proto to use case request
	useCaseReq := &update_product.Request{
		ProductID: req.ProductId,
	}
	if req.Name != nil {
		useCaseReq.Name = req.Name
	}
	if req.Description != nil {
		useCaseReq.Description = req.Description
	}
	if req.Category != nil {
		useCaseReq.Category = req.Category
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
