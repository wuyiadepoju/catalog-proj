# Product Catalog Service - gRPC API Usage Examples

This document provides comprehensive `grpcurl` examples for all API endpoints based on the business requirements.

## Prerequisites

1. Install grpcurl:
   ```bash
   go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
   ```

2. Start the server:
   ```bash
   make run
   ```

3. Default server address: `localhost:50051`

## Understanding the API

### Money Format
- Money amounts are in **cents** (smallest currency unit)
- Example: `$99.99` = `9999` cents
- Example: `$2499.99` = `249999` cents

### Discount Format
- Discount percentage is stored as cents (0-100 range)
- Example: `10%` discount = `10` (represents 0.10 = 10%)
- Example: `25%` discount = `25` (represents 0.25 = 25%)
- Example: `100%` discount = `100` (represents 1.00 = 100%)
- **Range:** 0-100 (representing 0% to 100%)
- **Note:** For fractional percentages like 13.33%, use the closest whole number (13 or 14)

### Timestamp Format
- Use RFC3339 format: `YYYY-MM-DDTHH:MM:SSZ`
- Example: `2026-02-25T00:00:00Z`
- **Note:** Discount dates should be from `2026-02-25T00:00:00Z` onward

---

## 1. Product Management - Create Products

### Create Product with Name, Description, Base Price, and Category

```bash
# Create a laptop product
grpcurl -plaintext -d '{
  "name": "MacBook Pro 16",
  "description": "High-performance laptop with M3 chip, 16GB RAM, 512GB SSD",
  "category": "electronics",
  "base_price": {"amount": "249999"}
}' localhost:50051 product.v1.ProductService/CreateProduct

# Create a book product
grpcurl -plaintext -d '{
  "name": "Domain-Driven Design",
  "description": "Tackling Complexity in the Heart of Software by Eric Evans",
  "category": "books",
  "base_price": {"amount": "5999"}
}' localhost:50051 product.v1.ProductService/CreateProduct

# Create a headphones product
grpcurl -plaintext -d '{
  "name": "Sony WH-1000XM5",
  "description": "Premium noise-cancelling wireless headphones",
  "category": "electronics",
  "base_price": {"amount": "39999"}
}' localhost:50051 product.v1.ProductService/CreateProduct
```

**Response:**
```json
{
  "product_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

---

## 2. Product Management - Activate Products

**Important:** Products are created in `inactive` status by default. You **must** activate them before applying discounts, as discounts can only be applied to active products.

```bash
# Activate a product
grpcurl -plaintext -d '{
  "product_id": "550e8400-e29b-41d4-a716-446655440000"
}' localhost:50051 product.v1.ProductService/ActivateProduct
```

**Response:**
```json
{
  "product_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

---

## 3. Product Management - Update Product Details

Update name, description, or category. All fields are optional.

```bash
# Update product name and description
grpcurl -plaintext -d '{
  "product_id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "MacBook Pro 16 (Updated)",
  "description": "Updated: High-performance laptop with M3 Pro chip, 18GB RAM, 1TB SSD"
}' localhost:50051 product.v1.ProductService/UpdateProduct

# Update category only
grpcurl -plaintext -d '{
  "product_id": "550e8400-e29b-41d4-a716-446655440000",
  "category": "computers"
}' localhost:50051 product.v1.ProductService/UpdateProduct

# Update name only
grpcurl -plaintext -d '{
  "product_id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "New Product Name"
}' localhost:50051 product.v1.ProductService/UpdateProduct
```

**Response:**
```json
{
  "product_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

---

## 4. Product Management - Deactivate Products

```bash
# Deactivate a product
grpcurl -plaintext -d '{
  "product_id": "550e8400-e29b-41d4-a716-446655440000"
}' localhost:50051 product.v1.ProductService/DeactivateProduct
```

**Response:**
```json
{
  "product_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

---

## 5. Product Management - Archive Products (Soft Delete)

```bash
# Archive a product
grpcurl -plaintext -d '{
  "product_id": "550e8400-e29b-41d4-a716-446655440000"
}' localhost:50051 product.v1.ProductService/ArchiveProduct
```

**Response:**
```json
{
  "product_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

---

## 6. Pricing Rules - Apply Discounts

Apply percentage-based discounts with start/end dates. Only one active discount per product at a time.

**Important Requirements:**
- Product must be **active** before applying a discount (use `ActivateProduct` first)
- Discount dates must be from **2026-02-25T00:00:00Z** onward
- Only one active discount per product at a time

```bash
# Apply 15% discount (valid from 2026-02-25 to end of year)
grpcurl -plaintext -d '{
  "product_id": "550e8400-e29b-41d4-a716-446655440000",
  "discount": {
    "id": "discount-laptop-15",
    "amount": {"amount": "15"},
    "start_date": "2026-02-25T00:00:00Z",
    "end_date": "2026-12-31T23:59:59Z"
  }
}' localhost:50051 product.v1.ProductService/ApplyDiscount

# Apply 20% discount
grpcurl -plaintext -d '{
  "product_id": "550e8400-e29b-41d4-a716-446655440000",
  "discount": {
    "id": "discount-book-20",
    "amount": {"amount": "20"},
    "start_date": "2026-02-25T00:00:00Z",
    "end_date": "2026-12-31T23:59:59Z"
  }
}' localhost:50051 product.v1.ProductService/ApplyDiscount

# Apply discount with future start date (not yet active)
grpcurl -plaintext -d '{
  "product_id": "550e8400-e29b-41d4-a716-446655440000",
  "discount": {
    "id": "discount-future",
    "amount": {"amount": "30"},
    "start_date": "2026-03-01T00:00:00Z",
    "end_date": "2026-03-31T23:59:59Z"
  }
}' localhost:50051 product.v1.ProductService/ApplyDiscount

# Apply 13% discount (for fractional percentages, use closest whole number)
grpcurl -plaintext -d '{
  "product_id": "550e8400-e29b-41d4-a716-446655440000",
  "discount": {
    "id": "discount-13",
    "amount": {"amount": "13"},
    "start_date": "2026-02-25T00:00:00Z",
    "end_date": "2026-12-31T23:59:59Z"
  }
}' localhost:50051 product.v1.ProductService/ApplyDiscount
```

**Response:**
```json
{
  "product_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Business Rules:**
- Product must be `active` to apply discount
- Only one active discount per product at a time
- Discount must have valid start/end dates

---

## 7. Pricing Rules - Remove Discount

```bash
# Remove discount from a product
grpcurl -plaintext -d '{
  "product_id": "550e8400-e29b-41d4-a716-446655440000"
}' localhost:50051 product.v1.ProductService/RemoveDiscount
```

**Response:**
```json
{
  "product_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

---

## 8. Product Queries - Get Product by ID

Get product with current effective price (base price or discounted price).

```bash
# Get product by ID
grpcurl -plaintext -d '{
  "product_id": "550e8400-e29b-41d4-a716-446655440000"
}' localhost:50051 product.v1.ProductService/GetProduct
```

**Response:**
```json
{
  "product": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "MacBook Pro 16",
    "description": "High-performance laptop with M3 chip, 16GB RAM, 512GB SSD",
    "category": "electronics",
    "base_price": {"amount": 249999},
    "effective_price": {"amount": 212499},
    "discount": {
      "id": "discount-laptop-15",
      "amount": {"amount": 15},
      "start_date": "2026-02-25T00:00:00Z",
      "end_date": "2026-12-31T23:59:59Z"
    },
    "status": "active",
    "created_at": "2026-02-25T10:00:00Z",
    "updated_at": "2026-02-25T10:05:00Z"
  }
}
```

---

## 9. Product Queries - List Products with Pagination

List products with optional filters by category and status.

```bash
# List all active products (pagination)
grpcurl -plaintext -d '{
  "status": "active",
  "limit": 10,
  "offset": 0
}' localhost:50051 product.v1.ProductService/ListProducts

# List products filtered by category
grpcurl -plaintext -d '{
  "category": "electronics",
  "status": "active",
  "limit": 10,
  "offset": 0
}' localhost:50051 product.v1.ProductService/ListProducts

# List products filtered by category: books
grpcurl -plaintext -d '{
  "category": "books",
  "status": "active",
  "limit": 10,
  "offset": 0
}' localhost:50051 product.v1.ProductService/ListProducts

# List products with pagination (page 1)
grpcurl -plaintext -d '{
  "limit": 2,
  "offset": 0
}' localhost:50051 product.v1.ProductService/ListProducts

# List products with pagination (page 2)
grpcurl -plaintext -d '{
  "limit": 2,
  "offset": 2
}' localhost:50051 product.v1.ProductService/ListProducts

# List all products (no filters)
grpcurl -plaintext -d '{
  "limit": 10,
  "offset": 0
}' localhost:50051 product.v1.ProductService/ListProducts
```

**Response:**
```json
{
  "products": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "MacBook Pro 16",
      "description": "High-performance laptop with M3 chip, 16GB RAM, 512GB SSD",
      "category": "electronics",
      "base_price": {"amount": 249999},
      "effective_price": {"amount": 212499},
      "discount": {
        "id": "discount-laptop-15",
        "amount": {"amount": 15},
        "start_date": "2026-02-25T00:00:00Z",
        "end_date": "2026-12-31T23:59:59Z"
      },
      "status": "active",
      "created_at": "2026-02-25T10:00:00Z",
      "updated_at": "2026-02-25T10:05:00Z"
    }
  ],
  "total": 1
}
```

---

## 10. Complete Workflow Example

Here's a complete workflow demonstrating all operations:

```bash
# 1. Create a product
RESPONSE=$(grpcurl -plaintext -d '{
  "name": "Test Product",
  "description": "A test product",
  "category": "test",
  "base_price": {"amount": "10000"}
}' localhost:50051 product.v1.ProductService/CreateProduct)

# Extract product_id from response (requires jq or manual parsing)
PRODUCT_ID=$(echo $RESPONSE | jq -r '.product_id')

# 2. Activate the product
grpcurl -plaintext -d "{\"product_id\": \"$PRODUCT_ID\"}" \
  localhost:50051 product.v1.ProductService/ActivateProduct

# 3. Get the product (should show base price)
grpcurl -plaintext -d "{\"product_id\": \"$PRODUCT_ID\"}" \
  localhost:50051 product.v1.ProductService/GetProduct

# 4. Apply a discount
grpcurl -plaintext -d "{
  \"product_id\": \"$PRODUCT_ID\",
  \"discount\": {
    \"id\": \"discount-10\",
    \"amount\": {\"amount\": \"10\"},
    \"start_date\": \"2026-02-25T00:00:00Z\",
    \"end_date\": \"2026-12-31T23:59:59Z\"
  }
}" localhost:50051 product.v1.ProductService/ApplyDiscount

# 5. Get the product again (should show discounted price)
grpcurl -plaintext -d "{\"product_id\": \"$PRODUCT_ID\"}" \
  localhost:50051 product.v1.ProductService/GetProduct

# 6. Update the product
grpcurl -plaintext -d "{
  \"product_id\": \"$PRODUCT_ID\",
  \"name\": \"Updated Test Product\",
  \"description\": \"Updated description\"
}" localhost:50051 product.v1.ProductService/UpdateProduct

# 7. List products
grpcurl -plaintext -d '{
  "status": "active",
  "limit": 10,
  "offset": 0
}' localhost:50051 product.v1.ProductService/ListProducts

# 8. Archive the product
grpcurl -plaintext -d "{\"product_id\": \"$PRODUCT_ID\"}" \
  localhost:50051 product.v1.ProductService/ArchiveProduct
```

---

## 11. Error Cases

### Try to apply discount to inactive product
```bash
# This will fail with error: product must be active
grpcurl -plaintext -d '{
  "product_id": "550e8400-e29b-41d4-a716-446655440000",
  "discount": {
    "id": "discount-1",
    "amount": {"amount": "10"},
    "start_date": "2026-02-25T00:00:00Z",
    "end_date": "2026-12-31T23:59:59Z"
  }
}' localhost:50051 product.v1.ProductService/ApplyDiscount
```

### Try to apply second discount (only one active discount allowed)
```bash
# This will fail if product already has an active discount
grpcurl -plaintext -d '{
  "product_id": "550e8400-e29b-41d4-a716-446655440000",
  "discount": {
    "id": "discount-2",
    "amount": {"amount": "20"},
    "start_date": "2026-02-25T00:00:00Z",
    "end_date": "2026-12-31T23:59:59Z"
  }
}' localhost:50051 product.v1.ProductService/ApplyDiscount
```

---

## 12. Event Publishing

Events are automatically published via the transactional outbox pattern when:
- Products are created (`ProductCreatedEvent`)
- Products are updated (`ProductUpdatedEvent`)
- Discounts are applied (`DiscountAppliedEvent`)
- Discounts are removed (`DiscountRemovedEvent`)
- Products are activated/deactivated (`ProductStatusChangedEvent`)
- Products are archived (`ProductArchivedEvent`)

Events are stored in the `m_outbox` table within the same transaction, ensuring reliable publishing.

---

## Running the Complete Test Suite

Use the provided test script to run all test cases:

```bash
# Make sure server is running first
make run

# In another terminal, run the test script
./test_api.sh

# Or specify a different server address
GRPC_HOST=localhost:50051 ./test_api.sh
```

The test script covers:
- ✅ Product Management: Create, Update, Activate, Deactivate, Archive
- ✅ Pricing Rules: Apply discount, Remove discount, Business rule validation
- ✅ Product Queries: Get by ID, List with pagination, Filter by category
- ✅ Precise Decimal Arithmetic: Various discount percentages
- ✅ Edge Cases: Future discounts, inactive products, multiple discounts
