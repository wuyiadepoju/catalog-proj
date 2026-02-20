package repo

import (
	"context"
	"fmt"
	"math/big"

	"catalog-proj/internal/app/product/queries/get_product"
	"catalog-proj/internal/app/product/queries/list_products"
	"catalog-proj/internal/models/m_product"
	"cloud.google.com/go/spanner"
	"google.golang.org/api/iterator"
)

// SpannerReadModel implements ReadModel using direct Spanner queries
// This bypasses the domain layer and returns DTOs directly
type SpannerReadModel struct {
	client *spanner.Client
}

// NewSpannerReadModel creates a new Spanner read model
func NewSpannerReadModel(client *spanner.Client) *SpannerReadModel {
	return &SpannerReadModel{
		client: client,
	}
}

// GetProduct retrieves a single product by ID
func (r *SpannerReadModel) GetProduct(ctx context.Context, id string) (*get_product.DTO, error) {
	row, err := r.client.Single().ReadRow(ctx, m_product.TableName, spanner.Key{id}, m_product.AllColumns())
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	model := &m_product.Product{}
	if err := row.ToStruct(model); err != nil {
		return nil, fmt.Errorf("failed to parse product row: %w", err)
	}

	return r.modelToDTO(model), nil
}

// ListProducts retrieves a list of products with optional filters
func (r *SpannerReadModel) ListProducts(ctx context.Context, req *list_products.Request) (*list_products.DTO, error) {
	// Build base WHERE clause for both count and data queries
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argIndex := 1

	if req.Category != "" {
		whereClause += fmt.Sprintf(" AND category = @p%d", argIndex)
		args = append(args, req.Category)
		argIndex++
	}

	if req.Status != "" {
		whereClause += fmt.Sprintf(" AND status = @p%d", argIndex)
		args = append(args, req.Status)
		argIndex++
	}

	// Get total count (separate query without limit/offset)
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) as total
		FROM %s
		%s
	`, m_product.TableName, whereClause)

	countStmt := spanner.Statement{
		SQL:    countQuery,
		Params: buildParams(args),
	}

	countIter := r.client.Single().Query(ctx, countStmt)
	defer countIter.Stop()

	var total int
	countRow, err := countIter.Next()
	if err != nil && err != iterator.Done {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}
	if err == nil {
		var countValue int64
		if err := countRow.ColumnByName("total", &countValue); err != nil {
			return nil, fmt.Errorf("failed to read count: %w", err)
		}
		total = int(countValue)
	}

	// Build data query with limit/offset
	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		%s
		ORDER BY created_at DESC
	`, buildColumnList(m_product.AllColumns()), m_product.TableName, whereClause)

	dataArgs := make([]interface{}, len(args))
	copy(dataArgs, args)
	dataArgIndex := argIndex

	if req.Limit > 0 {
		query += fmt.Sprintf(" LIMIT @p%d", dataArgIndex)
		dataArgs = append(dataArgs, req.Limit)
		dataArgIndex++
	}

	if req.Offset > 0 {
		query += fmt.Sprintf(" OFFSET @p%d", dataArgIndex)
		dataArgs = append(dataArgs, req.Offset)
	}

	stmt := spanner.Statement{
		SQL:    query,
		Params: buildParams(dataArgs),
	}

	iter := r.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	var products []list_products.ProductItem
	for {
		row, err := iter.Next()
		if err != nil {
			// iterator.Done() indicates end of iteration
			if err == iterator.Done {
				break
			}
			return nil, fmt.Errorf("failed to iterate products: %w", err)
		}

		model := &m_product.Product{}
		if err := row.ToStruct(model); err != nil {
			return nil, fmt.Errorf("failed to parse product row: %w", err)
		}

		products = append(products, r.modelToProductItem(model))
	}

	return &list_products.DTO{
		Products: products,
		Total:    total,
	}, nil
}

// modelToDTO converts a database model to a GetProduct DTO
func (r *SpannerReadModel) modelToDTO(model *m_product.Product) *get_product.DTO {
	// Convert numerator/denominator to *big.Rat
	var basePrice *big.Rat
	if model.BasePriceDenominator != 0 {
		basePrice = big.NewRat(model.BasePriceNumerator, model.BasePriceDenominator)
	}

	return &get_product.DTO{
		ID:                model.ProductID,
		Name:              model.Name,
		Description:       model.Description,
		Category:          model.Category,
		BasePrice:         basePrice,
		DiscountID:        model.DiscountID,
		DiscountAmount:    model.DiscountAmount,
		DiscountStartDate: model.DiscountStartDate,
		DiscountEndDate:   model.DiscountEndDate,
		Status:            model.Status,
		ArchivedAt:        model.ArchivedAt,
		CreatedAt:         model.CreatedAt,
		UpdatedAt:         model.UpdatedAt,
	}
}

// modelToProductItem converts a database model to a ListProducts ProductItem
func (r *SpannerReadModel) modelToProductItem(model *m_product.Product) list_products.ProductItem {
	// Convert numerator/denominator to *big.Rat
	var basePrice *big.Rat
	if model.BasePriceDenominator != 0 {
		basePrice = big.NewRat(model.BasePriceNumerator, model.BasePriceDenominator)
	}

	return list_products.ProductItem{
		ID:                model.ProductID,
		Name:              model.Name,
		Description:       model.Description,
		Category:          model.Category,
		BasePrice:         basePrice,
		DiscountID:        model.DiscountID,
		DiscountAmount:    model.DiscountAmount,
		DiscountStartDate: model.DiscountStartDate,
		DiscountEndDate:   model.DiscountEndDate,
		Status:            model.Status,
		ArchivedAt:        model.ArchivedAt,
		CreatedAt:         model.CreatedAt,
		UpdatedAt:         model.UpdatedAt,
	}
}

// buildColumnList converts a slice of column names to a comma-separated string
func buildColumnList(columns []string) string {
	result := ""
	for i, col := range columns {
		if i > 0 {
			result += ", "
		}
		result += col
	}
	return result
}

// buildParams converts a slice of values to a map for Spanner parameters
func buildParams(args []interface{}) map[string]interface{} {
	params := make(map[string]interface{})
	for i, arg := range args {
		params[fmt.Sprintf("p%d", i+1)] = arg
	}
	return params
}
