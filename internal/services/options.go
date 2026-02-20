package services

import (
	"context"
	"fmt"
	"os"

	"catalog-proj/internal/app/product/queries/get_product"
	"catalog-proj/internal/app/product/queries/list_products"
	domainServices "catalog-proj/internal/app/product/domain/services"
	"catalog-proj/internal/app/product/repo"
	"catalog-proj/internal/app/product/usecases/activate_product"
	"catalog-proj/internal/app/product/usecases/apply_discount"
	"catalog-proj/internal/app/product/usecases/archive_product"
	"catalog-proj/internal/app/product/usecases/create_product"
	"catalog-proj/internal/app/product/usecases/deactivate_product"
	"catalog-proj/internal/app/product/usecases/remove_discount"
	"catalog-proj/internal/app/product/usecases/update_product"
	"catalog-proj/internal/pkg/clock"
	spannerdriver "github.com/wuyiadepoju/commitplan/drivers/spanner"
	"catalog-proj/internal/transport/grpc/product"

	"cloud.google.com/go/spanner"
	"google.golang.org/grpc"
)

// Options holds all service dependencies
type Options struct {
	SpannerClient *spanner.Client
	GRPCServer    *grpc.Server
	ProductHandler *product.Handler
}

// NewOptions creates and wires all dependencies
func NewOptions(ctx context.Context, spannerDatabase string) (*Options, error) {
	// 1. Create Spanner client
	spannerClient, err := createSpannerClient(ctx, spannerDatabase)
	if err != nil {
		return nil, fmt.Errorf("failed to create Spanner client: %w", err)
	}

	// 2. Create clock
	clock := clock.NewRealClock()

	// 3. Create committer
	spannerCommitter := spannerdriver.NewCommitter(spannerClient)

	// 4. Create repositories
	productRepo := repo.NewSpannerProductRepository(spannerClient)
	spannerReadModel := repo.NewSpannerReadModel(spannerClient)

	// 5. Create domain services
	pricingCalculator := domainServices.NewPricingCalculator()

	// 6. Create use cases
	createProductInteractor := create_product.NewInteractor(
		productRepo,
		spannerCommitter,
		clock,
	)

	updateProductInteractor := update_product.NewInteractor(
		productRepo,
		spannerCommitter,
		clock,
	)

	applyDiscountInteractor := apply_discount.NewInteractor(
		productRepo,
		spannerCommitter,
		clock,
	)

	removeDiscountInteractor := remove_discount.NewInteractor(
		productRepo,
		spannerCommitter,
		clock,
	)

	activateProductInteractor := activate_product.NewInteractor(
		productRepo,
		spannerCommitter,
		clock,
	)

	deactivateProductInteractor := deactivate_product.NewInteractor(
		productRepo,
		spannerCommitter,
		clock,
	)

	archiveProductInteractor := archive_product.NewInteractor(
		productRepo,
		spannerCommitter,
		clock,
	)

	// 7. Create queries
	// Note: Each query package has its own ReadModel interface to avoid import cycles
	var readModelForGet get_product.ReadModel = spannerReadModel
	var readModelForList list_products.ReadModel = spannerReadModel
	
	getProductQuery := get_product.NewQuery(
		readModelForGet,
		pricingCalculator,
		clock,
	)

	listProductsQuery := list_products.NewQuery(
		readModelForList,
		pricingCalculator,
		clock,
	)

	// 8. Create gRPC handler
	productHandler := product.NewHandler(
		createProductInteractor,
		updateProductInteractor,
		applyDiscountInteractor,
		removeDiscountInteractor,
		activateProductInteractor,
		deactivateProductInteractor,
		archiveProductInteractor,
		getProductQuery,
		listProductsQuery,
	)

	// 9. Create gRPC server
	grpcServer := grpc.NewServer()

	return &Options{
		SpannerClient:  spannerClient,
		GRPCServer:      grpcServer,
		ProductHandler: productHandler,
	}, nil
}

// createSpannerClient creates a Spanner client
func createSpannerClient(ctx context.Context, database string) (*spanner.Client, error) {
	// Check if using emulator (for local development)
	emulatorHost := os.Getenv("SPANNER_EMULATOR_HOST")
	if emulatorHost != "" {
		// For emulator, database string format: projects/{project}/instances/{instance}/databases/{database}
		// Or we can use a simpler format if emulator is configured
		client, err := spanner.NewClient(ctx, database)
		if err != nil {
			return nil, fmt.Errorf("failed to create Spanner client (emulator): %w", err)
		}
		return client, nil
	}

	// Production Spanner client
	client, err := spanner.NewClient(ctx, database)
	if err != nil {
		return nil, fmt.Errorf("failed to create Spanner client: %w", err)
	}

	return client, nil
}

// Close closes all resources
func (o *Options) Close() error {
	if o.SpannerClient != nil {
		o.SpannerClient.Close()
	}
	return nil
}
