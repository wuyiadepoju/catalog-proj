package product

import (
	"context"

	"catalog-proj/internal/app/product/usecases/deactivate_product"
	pb "catalog-proj/proto/product/v1"
)

// DeactivateProduct handles the DeactivateProduct gRPC request
func (h *Handler) DeactivateProduct(ctx context.Context, req *pb.DeactivateProductRequest) (*pb.DeactivateProductResponse, error) {
	// 1. Validate
	if req.ProductId == "" {
		return nil, invalidArgumentError("product_id is required")
	}

	// 2. Map proto to use case request
	useCaseReq := &deactivate_product.Request{
		ProductID: req.ProductId,
	}

	// 3. Call use case
	resp, err := h.deactivateProductInteractor.Execute(ctx, useCaseReq)
	if err != nil {
		return nil, MapDomainError(err)
	}

	// 4. Map response to proto
	return &pb.DeactivateProductResponse{
		ProductId: resp.ProductID,
	}, nil
}
