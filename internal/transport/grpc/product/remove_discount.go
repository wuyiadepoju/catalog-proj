package product

import (
	"context"

	"catalog-proj/internal/app/product/usecases/remove_discount"
	pb "catalog-proj/proto/product/v1"
)

// RemoveDiscount handles the RemoveDiscount gRPC request
func (h *Handler) RemoveDiscount(ctx context.Context, req *pb.RemoveDiscountRequest) (*pb.RemoveDiscountResponse, error) {
	// 1. Validate
	if req.ProductId == "" {
		return nil, invalidArgumentError("product_id is required")
	}

	// 2. Map proto to use case request
	useCaseReq := &remove_discount.Request{
		ProductID: req.ProductId,
	}

	// 3. Call use case
	resp, err := h.removeDiscountInteractor.Execute(ctx, useCaseReq)
	if err != nil {
		return nil, MapDomainError(err)
	}

	// 4. Map response to proto
	return &pb.RemoveDiscountResponse{
		ProductId: resp.ProductID,
	}, nil
}
