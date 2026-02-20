package product

import (
	"context"

	"catalog-proj/internal/app/product/usecases/create_product"
	pb "catalog-proj/proto/product/v1"
)

// CreateProduct handles the CreateProduct gRPC request
func (h *Handler) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.CreateProductResponse, error) {
	// 1. Validate
	if req.Name == "" {
		return nil, invalidArgumentError("name is required")
	}
	if req.Description == "" {
		return nil, invalidArgumentError("description is required")
	}
	if req.Category == "" {
		return nil, invalidArgumentError("category is required")
	}
	if req.BasePrice == nil {
		return nil, invalidArgumentError("base_price is required")
	}

	// 2. Map proto to use case request
	basePrice := ProtoMoneyToDomain(req.BasePrice)
	useCaseReq := &create_product.Request{
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
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
