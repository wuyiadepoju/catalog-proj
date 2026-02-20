package product

import (
	"context"

	"catalog-proj/internal/app/product/usecases/apply_discount"
	pb "catalog-proj/proto/product/v1"
)

// ApplyDiscount handles the ApplyDiscount gRPC request
func (h *Handler) ApplyDiscount(ctx context.Context, req *pb.ApplyDiscountRequest) (*pb.ApplyDiscountResponse, error) {
	// 1. Validate
	if req.ProductId == "" {
		return nil, invalidArgumentError("product_id is required")
	}
	if req.Discount == nil {
		return nil, invalidArgumentError("discount is required")
	}

	// 2. Map proto to use case request
	discount := ProtoDiscountToDomain(req.Discount)
	useCaseReq := &apply_discount.Request{
		ProductID: req.ProductId,
		Discount:  discount,
	}

	// 3. Call use case
	resp, err := h.applyDiscountInteractor.Execute(ctx, useCaseReq)
	if err != nil {
		return nil, MapDomainError(err)
	}

	// 4. Map response to proto
	return &pb.ApplyDiscountResponse{
		ProductId: resp.ProductID,
	}, nil
}
