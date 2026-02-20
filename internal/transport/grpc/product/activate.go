package product

import (
	"context"

	"catalog-proj/internal/app/product/usecases/activate_product"
	pb "catalog-proj/proto/product/v1"
)

// ActivateProduct handles the ActivateProduct gRPC request
func (h *Handler) ActivateProduct(ctx context.Context, req *pb.ActivateProductRequest) (*pb.ActivateProductResponse, error) {
	// 1. Validate
	if req.ProductId == "" {
		return nil, invalidArgumentError("product_id is required")
	}

	// 2. Map proto to use case request
	useCaseReq := &activate_product.Request{
		ProductID: req.ProductId,
	}

	// 3. Call use case
	resp, err := h.activateProductInteractor.Execute(ctx, useCaseReq)
	if err != nil {
		return nil, MapDomainError(err)
	}

	// 4. Map response to proto
	return &pb.ActivateProductResponse{
		ProductId: resp.ProductID,
	}, nil
}
