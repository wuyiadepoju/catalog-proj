package product

import (
	"context"

	"catalog-proj/internal/app/product/usecases/archive_product"
	pb "catalog-proj/proto/product/v1"
)

// ArchiveProduct handles the ArchiveProduct gRPC request
func (h *Handler) ArchiveProduct(ctx context.Context, req *pb.ArchiveProductRequest) (*pb.ArchiveProductResponse, error) {
	// 1. Validate
	if req.ProductId == "" {
		return nil, invalidArgumentError("product_id is required")
	}

	// 2. Map proto to use case request
	useCaseReq := &archive_product.Request{
		ProductID: req.ProductId,
	}

	// 3. Call use case
	resp, err := h.archiveProductInteractor.Execute(ctx, useCaseReq)
	if err != nil {
		return nil, MapDomainError(err)
	}

	// 4. Map response to proto
	return &pb.ArchiveProductResponse{
		ProductId: resp.ProductID,
	}, nil
}
