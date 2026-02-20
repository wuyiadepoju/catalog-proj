package product

import (
	"context"

	"catalog-proj/internal/app/product/queries/list_products"
	pb "catalog-proj/proto/product/v1"
)

// ListProducts handles the ListProducts gRPC request
func (h *Handler) ListProducts(ctx context.Context, req *pb.ListProductsRequest) (*pb.ListProductsResponse, error) {
	// 1. Validate (optional fields, so just validate limit/offset if provided)
	if req.Limit < 0 {
		return nil, invalidArgumentError("limit must be non-negative")
	}
	if req.Offset < 0 {
		return nil, invalidArgumentError("offset must be non-negative")
	}

	// 2. Map proto to query request
	queryReq := &list_products.Request{
		Limit:  int(req.Limit),
		Offset: int(req.Offset),
	}
	if req.Category != nil {
		queryReq.Category = *req.Category
	}
	if req.Status != nil {
		queryReq.Status = *req.Status
	}

	// 3. Call query
	dto, err := h.listProductsQuery.Execute(ctx, queryReq)
	if err != nil {
		return nil, MapDomainError(err)
	}

	// 4. Map DTO to proto
	protoProducts := make([]*pb.Product, 0, len(dto.Products))
	for _, item := range dto.Products {
		protoProducts = append(protoProducts, ListProductItemToProto(item))
	}

	// 5. Return response
	return &pb.ListProductsResponse{
		Products: protoProducts,
		Total:    int32(dto.Total),
	}, nil
}
