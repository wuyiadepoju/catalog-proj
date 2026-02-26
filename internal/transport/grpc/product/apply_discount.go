package product

import (
	"context"
	"strings"

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

	// Validate discount fields
	if strings.TrimSpace(req.Discount.Id) == "" {
		return nil, invalidArgumentError("discount.id is required and cannot be empty")
	}

	if req.Discount.Amount == nil {
		return nil, invalidArgumentError("discount.amount is required")
	}

	if req.Discount.Amount.Amount < 0 || req.Discount.Amount.Amount > 100 {
		return nil, invalidArgumentError("discount.amount must be between 0 and 100 (0-100%)")
	}

	if req.Discount.StartDate == nil {
		return nil, invalidArgumentError("discount.start_date is required")
	}

	if req.Discount.EndDate == nil {
		return nil, invalidArgumentError("discount.end_date is required")
	}

	startDate := req.Discount.StartDate.AsTime()
	endDate := req.Discount.EndDate.AsTime()

	if !startDate.Before(endDate) {
		return nil, invalidArgumentError("discount.start_date must be before end_date")
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
