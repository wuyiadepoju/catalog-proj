package product

import (
	"context"

	pb "catalog-proj/proto/product/v1"
)

// GetProduct handles the GetProduct gRPC request
func (h *Handler) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.GetProductResponse, error) {
	// 1. Validate
	if req.ProductId == "" {
		return nil, invalidArgumentError("product_id is required")
	}

	// 2. Call query (no mapping needed, query handles it)
	dto, err := h.getProductQuery.Execute(ctx, req.ProductId)
	if err != nil {
		return nil, MapDomainError(err)
	}

	// 3. Map DTO to proto
	protoProduct := DTOToProtoProduct(dto)

	// 4. Return response
	return &pb.GetProductResponse{
		Product: protoProduct,
	}, nil
}
