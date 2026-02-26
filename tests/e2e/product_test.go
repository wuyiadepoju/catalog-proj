package e2e

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"strings"
	"testing"
	"time"

	"catalog-proj/internal/app/product/domain"
	domainServices "catalog-proj/internal/app/product/domain/services"
	"catalog-proj/internal/app/product/queries/get_product"
	"catalog-proj/internal/app/product/queries/list_products"
	"catalog-proj/internal/app/product/repo"
	"catalog-proj/internal/app/product/usecases/activate_product"
	"catalog-proj/internal/app/product/usecases/apply_discount"
	"catalog-proj/internal/app/product/usecases/archive_product"
	"catalog-proj/internal/app/product/usecases/create_product"
	"catalog-proj/internal/app/product/usecases/deactivate_product"
	"catalog-proj/internal/app/product/usecases/remove_discount"
	"catalog-proj/internal/app/product/usecases/update_product"
	"catalog-proj/internal/models/m_outbox"
	"catalog-proj/internal/models/m_product"
	"catalog-proj/internal/pkg/clock"
	"catalog-proj/internal/services"

	spannerdriver "github.com/wuyiadepoju/commitplan/drivers/spanner"

	"cloud.google.com/go/spanner"
	admin "cloud.google.com/go/spanner/admin/database/apiv1"
	"cloud.google.com/go/spanner/admin/database/apiv1/databasepb"
	instanceadmin "cloud.google.com/go/spanner/admin/instance/apiv1"
	"cloud.google.com/go/spanner/admin/instance/apiv1/instancepb"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	testProject  = "test-project"
	testInstance = "test-instance"
	emulatorHost = "localhost:9010"
	// baseDiscountDate is the minimum date for discounts (2026-02-25T00:00:00Z)
	baseDiscountDateStr = "2026-02-25T00:00:00Z"
)

// getDiscountTime returns a time that is at least baseDiscountDate
func getDiscountTime() time.Time {
	baseDate, _ := time.Parse(time.RFC3339, baseDiscountDateStr)
	now := time.Now()
	if now.Before(baseDate) {
		return baseDate
	}
	return now
}

// testSetup holds test dependencies
type testSetup struct {
	ctx               context.Context
	cancel            context.CancelFunc
	database          string
	spannerClient     *spanner.Client
	adminClient       *admin.DatabaseAdminClient
	opts              *services.Options
	createProduct     *create_product.Interactor
	updateProduct     *update_product.Interactor
	applyDiscount     *apply_discount.Interactor
	removeDiscount    *remove_discount.Interactor
	activateProduct   *activate_product.Interactor
	deactivateProduct *deactivate_product.Interactor
	archiveProduct    *archive_product.Interactor
	getProductQuery   *get_product.Query
	listProductsQuery *list_products.Query
}

// setupTest creates a test database and initializes all dependencies
func setupTest(t *testing.T) *testSetup {
	// Create context with timeout for setup operations to prevent hanging
	setupCtx, setupCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer setupCancel()

	// Set emulator host
	os.Setenv("SPANNER_EMULATOR_HOST", emulatorHost)
	defer os.Unsetenv("SPANNER_EMULATOR_HOST")

	// Create unique database name for this test
	dbName := fmt.Sprintf("test-db-%s", uuid.New().String()[:8])
	database := fmt.Sprintf("projects/%s/instances/%s/databases/%s", testProject, testInstance, dbName)

	// Create admin client with timeout context
	adminClient, err := admin.NewDatabaseAdminClient(setupCtx)
	if err != nil {
		t.Fatalf("Failed to create admin client: %v. Make sure Spanner emulator is running (docker compose up -d)", err)
	}

	// Create instance if it doesn't exist (for emulator)
	instanceName := fmt.Sprintf("projects/%s/instances/%s", testProject, testInstance)
	projectName := fmt.Sprintf("projects/%s", testProject)

	instanceAdminClient, err := instanceadmin.NewInstanceAdminClient(setupCtx)
	if err != nil {
		t.Fatalf("Failed to create instance admin client: %v", err)
	}
	defer instanceAdminClient.Close()

	_, err = instanceAdminClient.GetInstance(setupCtx, &instancepb.GetInstanceRequest{
		Name: instanceName,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			// Instance doesn't exist, create it
			op, err := instanceAdminClient.CreateInstance(setupCtx, &instancepb.CreateInstanceRequest{
				Parent:     projectName,
				InstanceId: testInstance,
				Instance: &instancepb.Instance{
					DisplayName: testInstance,
				},
			})
			if err != nil {
				t.Fatalf("Failed to create instance: %v", err)
			}
			_, err = op.Wait(setupCtx)
			if err != nil {
				if setupCtx.Err() == context.DeadlineExceeded {
					t.Fatalf("Timeout waiting for instance creation. Is Spanner emulator running? (docker compose up -d)")
				}
				t.Fatalf("Failed to wait for instance creation: %v", err)
			}
		} else {
			t.Fatalf("Failed to check instance existence: %v", err)
		}
	}

	// Create database
	op, err := adminClient.CreateDatabase(setupCtx, &databasepb.CreateDatabaseRequest{
		Parent:          instanceName,
		CreateStatement: fmt.Sprintf("CREATE DATABASE `%s`", dbName),
	})
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	// Wait for database creation with timeout
	db, err := op.Wait(setupCtx)
	if err != nil {
		if setupCtx.Err() == context.DeadlineExceeded {
			t.Fatalf("Timeout waiting for database creation. Is Spanner emulator running? (docker compose up -d)")
		}
		t.Fatalf("Failed to wait for database creation: %v", err)
	}
	database = db.Name

	// Run migrations
	if err := runMigrations(setupCtx, adminClient, database); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Create a background context for test execution (not canceled when setup returns)
	ctx, cancel := context.WithCancel(context.Background())

	// Create Spanner client
	spannerClient, err := spanner.NewClient(ctx, database)
	if err != nil {
		cancel()
		t.Fatalf("Failed to create Spanner client: %v", err)
	}

	// Create service options to get all dependencies
	opts, err := services.NewOptions(ctx, database)
	if err != nil {
		cancel()
		t.Fatalf("Failed to create service options: %v", err)
	}

	// Create use cases and queries directly (same as in services.NewOptions)
	// We'll recreate them here for direct access in tests
	clock := clock.NewRealClock()
	spannerCommitter := spannerdriver.NewCommitter(spannerClient)
	productRepo := repo.NewSpannerProductRepository(spannerClient)
	spannerReadModel := repo.NewSpannerReadModel(spannerClient)
	pricingCalculator := domainServices.NewPricingCalculator()

	createProductUC := create_product.NewInteractor(productRepo, spannerCommitter, clock)
	updateProductUC := update_product.NewInteractor(productRepo, spannerCommitter, clock)
	applyDiscountUC := apply_discount.NewInteractor(productRepo, spannerCommitter, clock)
	removeDiscountUC := remove_discount.NewInteractor(productRepo, spannerCommitter, clock)
	activateProductUC := activate_product.NewInteractor(productRepo, spannerCommitter, clock)
	deactivateProductUC := deactivate_product.NewInteractor(productRepo, spannerCommitter, clock)
	archiveProductUC := archive_product.NewInteractor(productRepo, spannerCommitter, clock)

	var readModelForGet get_product.ReadModel = spannerReadModel
	var readModelForList list_products.ReadModel = spannerReadModel
	getProductQ := get_product.NewQuery(readModelForGet, pricingCalculator, clock)
	listProductsQ := list_products.NewQuery(readModelForList, pricingCalculator, clock)

	return &testSetup{
		ctx:               ctx,
		cancel:            cancel,
		database:          database,
		spannerClient:     spannerClient,
		adminClient:       adminClient,
		opts:              opts,
		createProduct:     createProductUC,
		updateProduct:     updateProductUC,
		applyDiscount:     applyDiscountUC,
		removeDiscount:    removeDiscountUC,
		activateProduct:   activateProductUC,
		deactivateProduct: deactivateProductUC,
		archiveProduct:    archiveProductUC,
		getProductQuery:   getProductQ,
		listProductsQuery: listProductsQ,
	}
}

// teardownTest cleans up test resources
func (ts *testSetup) teardownTest(t *testing.T) {
	// Cancel context first to stop any ongoing operations
	if ts.cancel != nil {
		ts.cancel()
	}

	if ts.spannerClient != nil {
		ts.spannerClient.Close()
	}

	if ts.adminClient != nil {
		// Use a fresh context for cleanup operations
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()

		// Drop database
		if err := ts.adminClient.DropDatabase(cleanupCtx, &databasepb.DropDatabaseRequest{
			Database: ts.database,
		}); err != nil {
			t.Logf("Failed to drop database: %v", err)
		}
		ts.adminClient.Close()
	}

	if ts.opts != nil {
		ts.opts.Close()
	}
}

// runMigrations runs database migrations
func runMigrations(ctx context.Context, adminClient *admin.DatabaseAdminClient, database string) error {
	migrationSQL, err := os.ReadFile("../../migrations/001_initial_schema.sql")
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	statements := parseDDLStatements(string(migrationSQL))

	op, err := adminClient.UpdateDatabaseDdl(ctx, &databasepb.UpdateDatabaseDdlRequest{
		Database:   database,
		Statements: statements,
	})
	if err != nil {
		return fmt.Errorf("failed to start DDL operation: %w", err)
	}

	err = op.Wait(ctx)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("timeout waiting for migrations. Is Spanner emulator running? (docker compose up -d)")
		}
		return fmt.Errorf("failed to wait for migrations: %w", err)
	}
	return nil
}

// stringPtr returns a pointer to the given string
func stringPtr(s string) *string {
	return &s
}

// moneyFromRat creates a *domain.Money from a *big.Rat
func moneyFromRat(r *big.Rat) *domain.Money {
	money := domain.Money(r)
	return &money
}

// parseDDLStatements parses SQL into DDL statements
func parseDDLStatements(sql string) []string {
	var statements []string
	current := ""

	lines := strings.Split(sql, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "--") {
			continue
		}
		current += " " + trimmed
		if strings.HasSuffix(trimmed, ";") {
			stmt := strings.TrimSpace(strings.TrimSuffix(current, ";"))
			if stmt != "" {
				statements = append(statements, stmt)
			}
			current = ""
		}
	}

	return statements
}

// Helper functions for assertions and cleanup

// assertProductState verifies product state in database
func (ts *testSetup) assertProductState(t *testing.T, productID string, expectedStatus string, expectedArchived bool) {
	row, err := ts.spannerClient.Single().ReadRow(ts.ctx, m_product.TableName, spanner.Key{productID}, []string{
		m_product.ProductID, m_product.Status, m_product.ArchivedAt,
	})
	if err != nil {
		t.Fatalf("Failed to read product: %v", err)
	}

	var model m_product.Product
	if err := row.ToStruct(&model); err != nil {
		t.Fatalf("Failed to parse product: %v", err)
	}

	if model.Status != expectedStatus {
		t.Errorf("Expected status %s, got %s", expectedStatus, model.Status)
	}

	if expectedArchived && model.ArchivedAt == nil {
		t.Error("Expected product to be archived, but archived_at is nil")
	}
	if !expectedArchived && model.ArchivedAt != nil {
		t.Error("Expected product not to be archived, but archived_at is set")
	}
}

// assertOutboxEvents verifies outbox events were created
func (ts *testSetup) assertOutboxEvents(t *testing.T, expectedEventNames []string) {
	stmt := spanner.Statement{
		SQL: `SELECT event_type FROM outbox_events ORDER BY created_at`,
	}
	iter := ts.spannerClient.Single().Query(ts.ctx, stmt)
	defer iter.Stop()

	var eventNames []string
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			t.Fatalf("Failed to iterate events: %v", err)
		}
		var eventName string
		if err := row.ColumnByName("event_type", &eventName); err != nil {
			t.Fatalf("Failed to read event_type: %v", err)
		}
		eventNames = append(eventNames, eventName)
	}

	if len(eventNames) != len(expectedEventNames) {
		t.Errorf("Expected %d events, got %d", len(expectedEventNames), len(eventNames))
		return
	}

	for i, expected := range expectedEventNames {
		if i >= len(eventNames) || eventNames[i] != expected {
			t.Errorf("Event %d: expected %s, got %s", i, expected, eventNames[i])
		}
	}
}

// cleanupDatabase deletes all test data
func (ts *testSetup) cleanupDatabase(t *testing.T) {
	// Delete all products
	_, err := ts.spannerClient.Apply(ts.ctx, []*spanner.Mutation{
		spanner.Delete(m_product.TableName, spanner.AllKeys()),
	})
	if err != nil {
		t.Logf("Failed to cleanup products: %v", err)
	}

	// Delete all outbox events
	_, err = ts.spannerClient.Apply(ts.ctx, []*spanner.Mutation{
		spanner.Delete(m_outbox.TableName, spanner.AllKeys()),
	})
	if err != nil {
		t.Logf("Failed to cleanup outbox: %v", err)
	}
}

// Test scenarios

func TestProductCreationFlow(t *testing.T) {
	ts := setupTest(t)
	defer ts.teardownTest(t)
	defer ts.cleanupDatabase(t)

	// Create product
	basePrice := domain.NewMoney(9999) // $99.99 (9999 cents)
	req := &create_product.Request{
		Name:        "Test Product",
		Description: "A test product",
		Category:    "Electronics",
		BasePrice:   &basePrice,
	}

	resp, err := ts.createProduct.Execute(ts.ctx, req)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}

	productID := resp.ProductID
	if productID == "" {
		t.Error("Expected product ID, got empty string")
	}

	// Verify product exists in database (products start as inactive)
	ts.assertProductState(t, productID, string(domain.ProductStatusInactive), false)

	// Verify outbox event
	ts.assertOutboxEvents(t, []string{"product_created"})
}

func TestProductUpdateFlow(t *testing.T) {
	ts := setupTest(t)
	defer ts.teardownTest(t)
	defer ts.cleanupDatabase(t)

	// Create product first
	basePrice := domain.NewMoney(5000) // $50.00 (5000 cents)
	createReq := &create_product.Request{
		Name:        "Original Name",
		Description: "Original Description",
		Category:    "Electronics",
		BasePrice:   &basePrice,
	}

	createResp, err := ts.createProduct.Execute(ts.ctx, createReq)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}

	productID := createResp.ProductID

	// Update product
	updateReq := &update_product.Request{
		ProductID:   productID,
		Name:        stringPtr("Updated Name"),
		Description: stringPtr("Updated Description"),
		Category:    stringPtr("Books"),
	}

	_, err = ts.updateProduct.Execute(ts.ctx, updateReq)
	if err != nil {
		t.Fatalf("Failed to update product: %v", err)
	}

	// Verify product was updated
	row, err := ts.spannerClient.Single().ReadRow(ts.ctx, m_product.TableName, spanner.Key{productID}, []string{
		m_product.Name, m_product.Description, m_product.Category,
	})
	if err != nil {
		t.Fatalf("Failed to read product: %v", err)
	}

	var model m_product.Product
	if err := row.ToStruct(&model); err != nil {
		t.Fatalf("Failed to parse product: %v", err)
	}

	if model.Name != "Updated Name" {
		t.Errorf("Expected name 'Updated Name', got '%s'", model.Name)
	}
	if model.Description != "Updated Description" {
		t.Errorf("Expected description 'Updated Description', got '%s'", model.Description)
	}
	if model.Category != "Books" {
		t.Errorf("Expected category 'Books', got '%s'", model.Category)
	}

	// Verify outbox events
	ts.assertOutboxEvents(t, []string{"product_created", "product_updated"})
}

func TestDiscountApplicationFlow(t *testing.T) {
	ts := setupTest(t)
	defer ts.teardownTest(t)
	defer ts.cleanupDatabase(t)

	// Create product
	basePrice := domain.NewMoney(10000) // $100.00 (10000 cents)
	createReq := &create_product.Request{
		Name:        "Discounted Product",
		Description: "A product with discount",
		Category:    "Electronics",
		BasePrice:   &basePrice,
	}

	createResp, err := ts.createProduct.Execute(ts.ctx, createReq)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}
	productID := createResp.ProductID

	// Activate product first (required for discount)
	_, err = ts.activateProduct.Execute(ts.ctx, &activate_product.Request{ProductID: productID})
	if err != nil {
		t.Fatalf("Failed to activate product: %v", err)
	}

	// Apply discount
	now := time.Now()
	discountAmount := domain.NewMoney(20) // 20% discount (0.20 as decimal)
	discountReq := &apply_discount.Request{
		ProductID: productID,
		Discount: &domain.Discount{
			ID:        "discount-1",
			Amount:    &discountAmount,
			StartDate: now.Add(-1 * time.Hour),
			EndDate:   now.Add(24 * time.Hour),
		},
	}

	_, err = ts.applyDiscount.Execute(ts.ctx, discountReq)
	if err != nil {
		t.Fatalf("Failed to apply discount: %v", err)
	}

	// Verify discount was applied
	row, err := ts.spannerClient.Single().ReadRow(ts.ctx, m_product.TableName, spanner.Key{productID}, []string{
		m_product.DiscountID, m_product.DiscountAmount, m_product.DiscountStartDate, m_product.DiscountEndDate,
	})
	if err != nil {
		t.Fatalf("Failed to read product: %v", err)
	}

	var model m_product.Product
	if err := row.ToStruct(&model); err != nil {
		t.Fatalf("Failed to parse product: %v", err)
	}

	if model.DiscountID == nil || *model.DiscountID != "discount-1" {
		t.Errorf("Expected discount ID 'discount-1', got '%v'", model.DiscountID)
	}

	// Verify outbox events
	ts.assertOutboxEvents(t, []string{"product_created", "product_activated", "discount_applied"})
}

func TestProductActivationFlow(t *testing.T) {
	ts := setupTest(t)
	defer ts.teardownTest(t)
	defer ts.cleanupDatabase(t)

	// Create product (starts as Inactive)
	basePrice := domain.NewMoney(5000) // $50.00 (5000 cents)
	createReq := &create_product.Request{
		Name:        "Product to Activate",
		Description: "A product",
		Category:    "Electronics",
		BasePrice:   &basePrice,
	}

	createResp, err := ts.createProduct.Execute(ts.ctx, createReq)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}

	productID := createResp.ProductID

	// Verify initial state is Inactive (default for new products)
	ts.assertProductState(t, productID, string(domain.ProductStatusInactive), false)

	// Activate product
	_, err = ts.activateProduct.Execute(ts.ctx, &activate_product.Request{ProductID: productID})
	if err != nil {
		t.Fatalf("Failed to activate product: %v", err)
	}

	// Verify status changed to Active
	ts.assertProductState(t, productID, string(domain.ProductStatusActive), false)

	// Deactivate product
	_, err = ts.deactivateProduct.Execute(ts.ctx, &deactivate_product.Request{ProductID: productID})
	if err != nil {
		t.Fatalf("Failed to deactivate product: %v", err)
	}

	// Verify status changed to Inactive
	ts.assertProductState(t, productID, string(domain.ProductStatusInactive), false)

	// Verify outbox events
	ts.assertOutboxEvents(t, []string{
		"product_created",
		"product_activated",
		"product_deactivated",
	})
}

func TestBusinessRuleValidation(t *testing.T) {
	ts := setupTest(t)
	defer ts.teardownTest(t)
	defer ts.cleanupDatabase(t)

	// Create product
	createReq := &create_product.Request{
		Name:        "Product for Validation",
		Description: "A product",
		Category:    "Electronics",
		BasePrice:   moneyFromRat(big.NewRat(5000, 100)), // $50.00
	}

	createResp, err := ts.createProduct.Execute(ts.ctx, createReq)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}

	productID := createResp.ProductID

	// Try to apply discount to Inactive product (should fail)
	now := getDiscountTime()
	discountAmount := moneyFromRat(big.NewRat(10, 100)) // 10% discount (0.10 as decimal)
	discountReq := &apply_discount.Request{
		ProductID: productID,
		Discount: &domain.Discount{
			ID:        "discount-1",
			Amount:    discountAmount,
			StartDate: now.Add(-1 * time.Hour),
			EndDate:   now.Add(24 * time.Hour),
		},
	}

	_, err = ts.applyDiscount.Execute(ts.ctx, discountReq)
	if err == nil {
		t.Error("Expected error when applying discount to Inactive product, got nil")
	}

	// Activate product
	_, err = ts.activateProduct.Execute(ts.ctx, &activate_product.Request{ProductID: productID})
	if err != nil {
		t.Fatalf("Failed to activate product: %v", err)
	}

	// Now apply discount should work
	_, err = ts.applyDiscount.Execute(ts.ctx, discountReq)
	if err != nil {
		t.Fatalf("Failed to apply discount to active product: %v", err)
	}

	// Try to deactivate product with active discount (should fail)
	_, err = ts.deactivateProduct.Execute(ts.ctx, &deactivate_product.Request{ProductID: productID})
	if err == nil {
		t.Error("Expected error when deactivating product with active discount, got nil")
	}

	// Remove discount first
	_, err = ts.removeDiscount.Execute(ts.ctx, &remove_discount.Request{ProductID: productID})
	if err != nil {
		t.Fatalf("Failed to remove discount: %v", err)
	}

	// Now deactivate should work
	_, err = ts.deactivateProduct.Execute(ts.ctx, &deactivate_product.Request{ProductID: productID})
	if err != nil {
		t.Fatalf("Failed to deactivate product after removing discount: %v", err)
	}
}

func TestOutboxEventCreation(t *testing.T) {
	ts := setupTest(t)
	defer ts.teardownTest(t)
	defer ts.cleanupDatabase(t)

	// Create product
	createReq := &create_product.Request{
		Name:        "Product for Events",
		Description: "A product",
		Category:    "Electronics",
		BasePrice:   moneyFromRat(big.NewRat(5000, 100)),
	}

	createResp, err := ts.createProduct.Execute(ts.ctx, createReq)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}

	productID := createResp.ProductID

	// Verify outbox event was created
	// Use TO_JSON_STRING to convert JSON column to string for reading
	stmt := spanner.Statement{
		SQL: `SELECT event_type, TO_JSON_STRING(payload) as payload_str FROM outbox_events WHERE event_type = @eventType`,
		Params: map[string]interface{}{
			"eventType": "product_created",
		},
	}
	iter := ts.spannerClient.Single().Query(ts.ctx, stmt)
	defer iter.Stop()

	var found bool
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			t.Fatalf("Failed to iterate events: %v", err)
		}
		found = true

		var eventType, payloadStr string
		if err := row.ColumnByName("event_type", &eventType); err != nil {
			t.Fatalf("Failed to read event_type: %v", err)
		}
		if err := row.ColumnByName("payload_str", &payloadStr); err != nil {
			t.Fatalf("Failed to read payload: %v", err)
		}

		// Verify event payload contains product ID
		if !strings.Contains(payloadStr, productID) {
			t.Errorf("Event payload should contain product ID %s, got: %s", productID, payloadStr)
		}
	}

	if !found {
		t.Error("Expected product_created in outbox, but not found")
	}
}

func TestListProductsWithFilters(t *testing.T) {
	ts := setupTest(t)
	defer ts.teardownTest(t)
	defer ts.cleanupDatabase(t)

	// Create multiple products
	products := []struct {
		name     string
		category string
		price    *big.Rat
	}{
		{"Electronics Product 1", "Electronics", big.NewRat(10000, 100)},
		{"Electronics Product 2", "Electronics", big.NewRat(20000, 100)},
		{"Book Product 1", "Books", big.NewRat(3000, 100)},
		{"Book Product 2", "Books", big.NewRat(4000, 100)},
	}

	var productIDs []string
	for _, p := range products {
		req := &create_product.Request{
			Name:        p.name,
			Description: "Test product",
			Category:    p.category,
			BasePrice:   moneyFromRat(p.price),
		}
		resp, err := ts.createProduct.Execute(ts.ctx, req)
		if err != nil {
			t.Fatalf("Failed to create product %s: %v", p.name, err)
		}

		productIDs = append(productIDs, resp.ProductID)

		// Activate products
		_, err = ts.activateProduct.Execute(ts.ctx, &activate_product.Request{ProductID: resp.ProductID})
		if err != nil {
			t.Fatalf("Failed to activate product %s: %v", resp.ProductID, err)
		}
	}

	// Test filter by category
	listReq := &list_products.Request{
		Category: "Electronics",
		Limit:    10,
		Offset:   0,
	}

	result, err := ts.listProductsQuery.Execute(ts.ctx, listReq)
	if err != nil {
		t.Fatalf("Failed to list products: %v", err)
	}

	if len(result.Products) != 2 {
		t.Errorf("Expected 2 Electronics products, got %d", len(result.Products))
	}

	for _, product := range result.Products {
		if product.Category != "Electronics" {
			t.Errorf("Expected category Electronics, got %s", product.Category)
		}
	}

	// Test filter by status
	listReq = &list_products.Request{
		Status: string(domain.ProductStatusActive),
		Limit:  10,
		Offset: 0,
	}

	result, err = ts.listProductsQuery.Execute(ts.ctx, listReq)
	if err != nil {
		t.Fatalf("Failed to list products: %v", err)
	}

	if len(result.Products) != 4 {
		t.Errorf("Expected 4 active products, got %d", len(result.Products))
	}

	// Test pagination
	listReq = &list_products.Request{
		Limit:  2,
		Offset: 0,
	}

	result, err = ts.listProductsQuery.Execute(ts.ctx, listReq)
	if err != nil {
		t.Fatalf("Failed to list products: %v", err)
	}

	if len(result.Products) != 2 {
		t.Errorf("Expected 2 products with limit 2, got %d", len(result.Products))
	}

	if result.Total != 4 {
		t.Errorf("Expected total 4 products, got %d", result.Total)
	}
}

func TestGetProductWithEffectivePrice(t *testing.T) {
	ts := setupTest(t)
	defer ts.teardownTest(t)
	defer ts.cleanupDatabase(t)

	// Create and activate product
	createReq := &create_product.Request{
		Name:        "Product with Price",
		Description: "A product",
		Category:    "Electronics",
		BasePrice:   moneyFromRat(big.NewRat(10000, 100)), // $100.00
	}

	createResp, err := ts.createProduct.Execute(ts.ctx, createReq)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}

	productID := createResp.ProductID

	_, err = ts.activateProduct.Execute(ts.ctx, &activate_product.Request{ProductID: productID})
	if err != nil {
		t.Fatalf("Failed to activate product: %v", err)
	}

	// Get product without discount (effective price = base price)
	result, err := ts.getProductQuery.Execute(ts.ctx, productID)
	if err != nil {
		t.Fatalf("Failed to get product: %v", err)
	}

	// Verify effective price equals base price
	if result.EffectivePrice == nil {
		t.Fatal("Expected effective price, got nil")
	}

	expectedPrice := big.NewRat(10000, 100)
	if result.EffectivePrice.Cmp(expectedPrice) != 0 {
		t.Errorf("Expected effective price %s, got %s", expectedPrice.String(), result.EffectivePrice.String())
	}

	// Apply discount
	now := getDiscountTime()
	discountAmount := domain.NewMoney(20) // 20% discount (0.20 as decimal)
	discountReq := &apply_discount.Request{
		ProductID: productID,
		Discount: &domain.Discount{
			ID:        "discount-1",
			Amount:    &discountAmount,
			StartDate: now.Add(-1 * time.Hour),
			EndDate:   now.Add(24 * time.Hour),
		},
	}

	_, err = ts.applyDiscount.Execute(ts.ctx, discountReq)
	if err != nil {
		t.Fatalf("Failed to apply discount: %v", err)
	}

	// Get product again and verify effective price
	result, err = ts.getProductQuery.Execute(ts.ctx, productID)
	if err != nil {
		t.Fatalf("Failed to get product: %v", err)
	}

	// Effective price should be $80.00 (20% off $100.00)
	expectedPrice = big.NewRat(8000, 100)
	if result.EffectivePrice.Cmp(expectedPrice) != 0 {
		t.Errorf("Expected effective price %s, got %s", expectedPrice.String(), result.EffectivePrice.String())
	}
}
