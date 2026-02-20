package product

import (
	"catalog-proj/internal/app/product/queries/get_product"
	"catalog-proj/internal/app/product/queries/list_products"
	"catalog-proj/internal/app/product/usecases/activate_product"
	"catalog-proj/internal/app/product/usecases/apply_discount"
	"catalog-proj/internal/app/product/usecases/archive_product"
	"catalog-proj/internal/app/product/usecases/create_product"
	"catalog-proj/internal/app/product/usecases/deactivate_product"
	"catalog-proj/internal/app/product/usecases/remove_discount"
	"catalog-proj/internal/app/product/usecases/update_product"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	pb "catalog-proj/proto/product/v1"
)

// Handler implements the ProductService gRPC service
type Handler struct {
	pb.UnimplementedProductServiceServer

	// Use case interactors
	createProductInteractor     *create_product.Interactor
	updateProductInteractor     *update_product.Interactor
	applyDiscountInteractor     *apply_discount.Interactor
	removeDiscountInteractor    *remove_discount.Interactor
	activateProductInteractor   *activate_product.Interactor
	deactivateProductInteractor *deactivate_product.Interactor
	archiveProductInteractor    *archive_product.Interactor

	// Query handlers
	getProductQuery  *get_product.Query
	listProductsQuery *list_products.Query
}

// NewHandler creates a new gRPC handler with all dependencies
func NewHandler(
	createProductInteractor *create_product.Interactor,
	updateProductInteractor *update_product.Interactor,
	applyDiscountInteractor *apply_discount.Interactor,
	removeDiscountInteractor *remove_discount.Interactor,
	activateProductInteractor *activate_product.Interactor,
	deactivateProductInteractor *deactivate_product.Interactor,
	archiveProductInteractor *archive_product.Interactor,
	getProductQuery *get_product.Query,
	listProductsQuery *list_products.Query,
) *Handler {
	return &Handler{
		createProductInteractor:     createProductInteractor,
		updateProductInteractor:     updateProductInteractor,
		applyDiscountInteractor:     applyDiscountInteractor,
		removeDiscountInteractor:    removeDiscountInteractor,
		activateProductInteractor:   activateProductInteractor,
		deactivateProductInteractor: deactivateProductInteractor,
		archiveProductInteractor:    archiveProductInteractor,
		getProductQuery:             getProductQuery,
		listProductsQuery:           listProductsQuery,
	}
}

// invalidArgumentError is a helper to create invalid argument errors
func invalidArgumentError(msg string) error {
	return status.Errorf(codes.InvalidArgument, msg)
}
