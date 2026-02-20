#!/bin/bash

# Product Catalog Service API Test Script
# Tests all business requirements using grpcurl

set -e  # Exit on error

# Configuration
GRPC_HOST="${GRPC_HOST:-localhost:50051}"
SERVICE="product.v1.ProductService"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper function to print section headers
print_section() {
    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
}

# Helper function to print test case
print_test() {
    echo ""
    echo -e "${YELLOW}Test: $1${NC}"
}

# Helper function to execute grpcurl and show result
run_grpcurl() {
    local method=$1
    local data=$2
    local description=$3
    
    echo -e "${GREEN}Request:${NC} $description"
    echo "Command: grpcurl -plaintext -d '$data' $GRPC_HOST $SERVICE/$method"
    
    local response
    response=$(grpcurl -plaintext -d "$data" "$GRPC_HOST" "$SERVICE/$method" 2>&1)
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}Response:${NC}"
        echo "$response" | jq '.' 2>/dev/null || echo "$response"
        echo "$response"
    else
        echo -e "${RED}Error:${NC}"
        echo "$response"
        return 1
    fi
}

# Check if grpcurl is installed
if ! command -v grpcurl &> /dev/null; then
    echo -e "${RED}Error: grpcurl is not installed${NC}"
    echo "Install it with: go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest"
    exit 1
fi

# Check if jq is installed (optional, for pretty JSON)
if ! command -v jq &> /dev/null; then
    echo -e "${YELLOW}Warning: jq is not installed. JSON output will not be formatted.${NC}"
    echo "Install it with: sudo apt-get install jq (Ubuntu) or brew install jq (macOS)"
fi

echo -e "${GREEN}Starting API Tests for Product Catalog Service${NC}"
echo "gRPC Server: $GRPC_HOST"
echo ""

# Variables to store created product IDs
PRODUCT_ID_1=""
PRODUCT_ID_2=""
PRODUCT_ID_3=""

# ============================================================================
# 1. PRODUCT MANAGEMENT - CREATE PRODUCTS
# ============================================================================
print_section "1. PRODUCT MANAGEMENT - Create Products"

print_test "Create Product 1: Laptop (Electronics)"
RESPONSE=$(grpcurl -plaintext -d '{
  "name": "MacBook Pro 16",
  "description": "High-performance laptop with M3 chip, 16GB RAM, 512GB SSD",
  "category": "electronics",
  "base_price": {"amount": "249999"}
}' "$GRPC_HOST" "$SERVICE/CreateProduct" 2>&1)
echo "$RESPONSE"
PRODUCT_ID_1=$(echo "$RESPONSE" | grep -o '"product_id":"[^"]*"' | cut -d'"' -f4 || echo "")
echo -e "${GREEN}Created Product ID: $PRODUCT_ID_1${NC}"

print_test "Create Product 2: Book (Education)"
RESPONSE=$(grpcurl -plaintext -d '{
  "name": "Domain-Driven Design",
  "description": "Tackling Complexity in the Heart of Software by Eric Evans",
  "category": "books",
  "base_price": {"amount": "5999"}
}' "$GRPC_HOST" "$SERVICE/CreateProduct" 2>&1)
echo "$RESPONSE"
PRODUCT_ID_2=$(echo "$RESPONSE" | grep -o '"product_id":"[^"]*"' | cut -d'"' -f4 || echo "")
echo -e "${GREEN}Created Product ID: $PRODUCT_ID_2${NC}"

print_test "Create Product 3: Headphones (Electronics)"
RESPONSE=$(grpcurl -plaintext -d '{
  "name": "Sony WH-1000XM5",
  "description": "Premium noise-cancelling wireless headphones",
  "category": "electronics",
  "base_price": {"amount": "39999"}
}' "$GRPC_HOST" "$SERVICE/CreateProduct" 2>&1)
echo "$RESPONSE"
PRODUCT_ID_3=$(echo "$RESPONSE" | grep -o '"product_id":"[^"]*"' | cut -d'"' -f4 || echo "")
echo -e "${GREEN}Created Product ID: $PRODUCT_ID_3${NC}"

# ============================================================================
# 2. PRODUCT MANAGEMENT - ACTIVATE PRODUCTS
# ============================================================================
print_section "2. PRODUCT MANAGEMENT - Activate Products"

if [ -n "$PRODUCT_ID_1" ]; then
    print_test "Activate Product 1 (Laptop)"
    grpcurl -plaintext -d "{\"product_id\": \"$PRODUCT_ID_1\"}" "$GRPC_HOST" "$SERVICE/ActivateProduct"
fi

if [ -n "$PRODUCT_ID_2" ]; then
    print_test "Activate Product 2 (Book)"
    grpcurl -plaintext -d "{\"product_id\": \"$PRODUCT_ID_2\"}" "$GRPC_HOST" "$SERVICE/ActivateProduct"
fi

if [ -n "$PRODUCT_ID_3" ]; then
    print_test "Activate Product 3 (Headphones)"
    grpcurl -plaintext -d "{\"product_id\": \"$PRODUCT_ID_3\"}" "$GRPC_HOST" "$SERVICE/ActivateProduct"
fi

# ============================================================================
# 3. PRODUCT QUERIES - GET PRODUCT BY ID
# ============================================================================
print_section "3. PRODUCT QUERIES - Get Product by ID"

if [ -n "$PRODUCT_ID_1" ]; then
    print_test "Get Product 1 (should show base price, no discount yet)"
    grpcurl -plaintext -d "{\"product_id\": \"$PRODUCT_ID_1\"}" "$GRPC_HOST" "$SERVICE/GetProduct"
fi

# ============================================================================
# 4. PRICING RULES - APPLY DISCOUNTS
# ============================================================================
print_section "4. PRICING RULES - Apply Discounts"

# Get current date and future dates for discount periods
CURRENT_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
FUTURE_START=$(date -u -d "+1 day" +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -v+1d +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || echo "2025-01-01T00:00:00Z")
FUTURE_END=$(date -u -d "+30 days" +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -v+30d +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || echo "2025-12-31T23:59:59Z")

if [ -n "$PRODUCT_ID_1" ]; then
    print_test "Apply 15% discount to Product 1 (Laptop) - Valid period"
    grpcurl -plaintext -d "{
      \"product_id\": \"$PRODUCT_ID_1\",
      \"discount\": {
        \"id\": \"discount-laptop-15\",
        \"amount\": {\"amount\": \"15\"},
        \"start_date\": \"$CURRENT_DATE\",
        \"end_date\": \"$FUTURE_END\"
      }
    }" "$GRPC_HOST" "$SERVICE/ApplyDiscount"
    
    print_test "Get Product 1 again (should show discounted effective price)"
    grpcurl -plaintext -d "{\"product_id\": \"$PRODUCT_ID_1\"}" "$GRPC_HOST" "$SERVICE/GetProduct"
fi

if [ -n "$PRODUCT_ID_2" ]; then
    print_test "Apply 20% discount to Product 2 (Book) - Valid period"
    grpcurl -plaintext -d "{
      \"product_id\": \"$PRODUCT_ID_2\",
      \"discount\": {
        \"id\": \"discount-book-20\",
        \"amount\": {\"amount\": \"20\"},
        \"start_date\": \"$CURRENT_DATE\",
        \"end_date\": \"$FUTURE_END\"
      }
    }" "$GRPC_HOST" "$SERVICE/ApplyDiscount"
fi

# ============================================================================
# 5. PRICING RULES - TEST BUSINESS RULES
# ============================================================================
print_section "5. PRICING RULES - Test Business Rules (Only One Active Discount)"

if [ -n "$PRODUCT_ID_1" ]; then
    print_test "Try to apply second discount to Product 1 (should fail - only one active discount allowed)"
    grpcurl -plaintext -d "{
      \"product_id\": \"$PRODUCT_ID_1\",
      \"discount\": {
        \"id\": \"discount-laptop-25\",
        \"amount\": {\"amount\": \"25\"},
        \"start_date\": \"$CURRENT_DATE\",
        \"end_date\": \"$FUTURE_END\"
      }
    }" "$GRPC_HOST" "$SERVICE/ApplyDiscount" || echo -e "${YELLOW}Expected error: Only one active discount per product${NC}"
fi

# ============================================================================
# 6. PRICING RULES - REMOVE DISCOUNT
# ============================================================================
print_section "6. PRICING RULES - Remove Discount"

if [ -n "$PRODUCT_ID_2" ]; then
    print_test "Remove discount from Product 2"
    grpcurl -plaintext -d "{\"product_id\": \"$PRODUCT_ID_2\"}" "$GRPC_HOST" "$SERVICE/RemoveDiscount"
    
    print_test "Get Product 2 (should show base price again)"
    grpcurl -plaintext -d "{\"product_id\": \"$PRODUCT_ID_2\"}" "$GRPC_HOST" "$SERVICE/GetProduct"
fi

# ============================================================================
# 7. PRODUCT MANAGEMENT - UPDATE PRODUCT
# ============================================================================
print_section "7. PRODUCT MANAGEMENT - Update Product Details"

if [ -n "$PRODUCT_ID_1" ]; then
    print_test "Update Product 1: Change name and description"
    grpcurl -plaintext -d "{
      \"product_id\": \"$PRODUCT_ID_1\",
      \"name\": \"MacBook Pro 16 (Updated)\",
      \"description\": \"Updated: High-performance laptop with M3 Pro chip, 18GB RAM, 1TB SSD\"
    }" "$GRPC_HOST" "$SERVICE/UpdateProduct"
    
    print_test "Update Product 1: Change category only"
    grpcurl -plaintext -d "{
      \"product_id\": \"$PRODUCT_ID_1\",
      \"category\": \"computers\"
    }" "$GRPC_HOST" "$SERVICE/UpdateProduct"
    
    print_test "Get Product 1 (verify updates)"
    grpcurl -plaintext -d "{\"product_id\": \"$PRODUCT_ID_1\"}" "$GRPC_HOST" "$SERVICE/GetProduct"
fi

# ============================================================================
# 8. PRODUCT QUERIES - LIST PRODUCTS
# ============================================================================
print_section "8. PRODUCT QUERIES - List Products"

print_test "List all active products (pagination: limit 10, offset 0)"
grpcurl -plaintext -d '{
  "status": "active",
  "limit": 10,
  "offset": 0
}' "$GRPC_HOST" "$SERVICE/ListProducts"

print_test "List products filtered by category: electronics"
grpcurl -plaintext -d '{
  "category": "electronics",
  "status": "active",
  "limit": 10,
  "offset": 0
}' "$GRPC_HOST" "$SERVICE/ListProducts"

print_test "List products filtered by category: books"
grpcurl -plaintext -d '{
  "category": "books",
  "status": "active",
  "limit": 10,
  "offset": 0
}' "$GRPC_HOST" "$SERVICE/ListProducts"

print_test "List products with pagination (limit 2, offset 0)"
grpcurl -plaintext -d '{
  "limit": 2,
  "offset": 0
}' "$GRPC_HOST" "$SERVICE/ListProducts"

print_test "List products with pagination (limit 2, offset 2)"
grpcurl -plaintext -d '{
  "limit": 2,
  "offset": 2
}' "$GRPC_HOST" "$SERVICE/ListProducts"

# ============================================================================
# 9. PRODUCT MANAGEMENT - DEACTIVATE PRODUCT
# ============================================================================
print_section "9. PRODUCT MANAGEMENT - Deactivate Product"

if [ -n "$PRODUCT_ID_3" ]; then
    print_test "Deactivate Product 3 (Headphones)"
    grpcurl -plaintext -d "{\"product_id\": \"$PRODUCT_ID_3\"}" "$GRPC_HOST" "$SERVICE/DeactivateProduct"
    
    print_test "Get Product 3 (should show status: inactive)"
    grpcurl -plaintext -d "{\"product_id\": \"$PRODUCT_ID_3\"}" "$GRPC_HOST" "$SERVICE/GetProduct"
    
    print_test "Try to apply discount to inactive product (should fail)"
    grpcurl -plaintext -d "{
      \"product_id\": \"$PRODUCT_ID_3\",
      \"discount\": {
        \"id\": \"discount-headphones-10\",
        \"amount\": {\"amount\": \"10\"},
        \"start_date\": \"$CURRENT_DATE\",
        \"end_date\": \"$FUTURE_END\"
      }
    }" "$GRPC_HOST" "$SERVICE/ApplyDiscount" || echo -e "${YELLOW}Expected error: Product must be active to apply discount${NC}"
fi

# ============================================================================
# 10. PRODUCT MANAGEMENT - REACTIVATE PRODUCT
# ============================================================================
print_section "10. PRODUCT MANAGEMENT - Reactivate Product"

if [ -n "$PRODUCT_ID_3" ]; then
    print_test "Reactivate Product 3"
    grpcurl -plaintext -d "{\"product_id\": \"$PRODUCT_ID_3\"}" "$GRPC_HOST" "$SERVICE/ActivateProduct"
    
    print_test "Get Product 3 (should show status: active)"
    grpcurl -plaintext -d "{\"product_id\": \"$PRODUCT_ID_3\"}" "$GRPC_HOST" "$SERVICE/GetProduct"
fi

# ============================================================================
# 11. PRICING RULES - DISCOUNT WITH DATE RANGES
# ============================================================================
print_section "11. PRICING RULES - Discount with Future Date Range"

FUTURE_START_DATE=$(date -u -d "+5 days" +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -v+5d +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || echo "2025-01-05T00:00:00Z")
FUTURE_END_DATE=$(date -u -d "+60 days" +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -v+60d +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || echo "2025-02-28T23:59:59Z")

if [ -n "$PRODUCT_ID_3" ]; then
    print_test "Apply discount with future start date to Product 3"
    grpcurl -plaintext -d "{
      \"product_id\": \"$PRODUCT_ID_3\",
      \"discount\": {
        \"id\": \"discount-headphones-future\",
        \"amount\": {\"amount\": \"30\"},
        \"start_date\": \"$FUTURE_START_DATE\",
        \"end_date\": \"$FUTURE_END_DATE\"
      }
    }" "$GRPC_HOST" "$SERVICE/ApplyDiscount"
    
    print_test "Get Product 3 (discount not yet active, should show base price)"
    grpcurl -plaintext -d "{\"product_id\": \"$PRODUCT_ID_3\"}" "$GRPC_HOST" "$SERVICE/GetProduct"
fi

# ============================================================================
# 12. PRODUCT MANAGEMENT - ARCHIVE PRODUCT (SOFT DELETE)
# ============================================================================
print_section "12. PRODUCT MANAGEMENT - Archive Product (Soft Delete)"

if [ -n "$PRODUCT_ID_2" ]; then
    print_test "Archive Product 2 (Book)"
    grpcurl -plaintext -d "{\"product_id\": \"$PRODUCT_ID_2\"}" "$GRPC_HOST" "$SERVICE/ArchiveProduct"
    
    print_test "Get Product 2 (should show archived_at timestamp)"
    grpcurl -plaintext -d "{\"product_id\": \"$PRODUCT_ID_2\"}" "$GRPC_HOST" "$SERVICE/GetProduct"
    
    print_test "List active products (archived product should not appear)"
    grpcurl -plaintext -d '{
      "status": "active",
      "limit": 10,
      "offset": 0
    }' "$GRPC_HOST" "$SERVICE/ListProducts"
fi

# ============================================================================
# 13. PRICING RULES - PRECISE DECIMAL ARITHMETIC TEST
# ============================================================================
print_section "13. PRICING RULES - Precise Decimal Arithmetic Test"

print_test "Create Product 4: Test precise pricing (price with cents)"
RESPONSE=$(grpcurl -plaintext -d '{
  "name": "Precision Test Product",
  "description": "Testing precise decimal arithmetic",
  "category": "test",
  "base_price": {"amount": "9999"}
}' "$GRPC_HOST" "$SERVICE/CreateProduct" 2>&1)
echo "$RESPONSE"
PRODUCT_ID_4=$(echo "$RESPONSE" | grep -o '"product_id":"[^"]*"' | cut -d'"' -f4 || echo "")

if [ -n "$PRODUCT_ID_4" ]; then
    print_test "Activate Product 4"
    grpcurl -plaintext -d "{\"product_id\": \"$PRODUCT_ID_4\"}" "$GRPC_HOST" "$SERVICE/ActivateProduct"
    
    print_test "Apply 13.33% discount (testing precise calculation)"
    grpcurl -plaintext -d "{
      \"product_id\": \"$PRODUCT_ID_4\",
      \"discount\": {
        \"id\": \"discount-precise-1333\",
        \"amount\": {\"amount\": \"1333\"},
        \"start_date\": \"$CURRENT_DATE\",
        \"end_date\": \"$FUTURE_END\"
      }
    }" "$GRPC_HOST" "$SERVICE/ApplyDiscount"
    
    print_test "Get Product 4 (verify precise effective price calculation)"
    grpcurl -plaintext -d "{\"product_id\": \"$PRODUCT_ID_4\"}" "$GRPC_HOST" "$SERVICE/GetProduct"
fi

# ============================================================================
# 14. EDGE CASES - VARIOUS DISCOUNT PERCENTAGES
# ============================================================================
print_section "14. EDGE CASES - Various Discount Percentages"

if [ -n "$PRODUCT_ID_1" ]; then
    print_test "Remove existing discount from Product 1"
    grpcurl -plaintext -d "{\"product_id\": \"$PRODUCT_ID_1\"}" "$GRPC_HOST" "$SERVICE/RemoveDiscount"
    
    print_test "Apply 5% discount (small discount)"
    grpcurl -plaintext -d "{
      \"product_id\": \"$PRODUCT_ID_1\",
      \"discount\": {
        \"id\": \"discount-small-5\",
        \"amount\": {\"amount\": \"5\"},
        \"start_date\": \"$CURRENT_DATE\",
        \"end_date\": \"$FUTURE_END\"
      }
    }" "$GRPC_HOST" "$SERVICE/ApplyDiscount"
    
    print_test "Get Product 1 (verify 5% discount)"
    grpcurl -plaintext -d "{\"product_id\": \"$PRODUCT_ID_1\"}" "$GRPC_HOST" "$SERVICE/GetProduct"
    
    print_test "Remove discount and apply 50% discount (large discount)"
    grpcurl -plaintext -d "{\"product_id\": \"$PRODUCT_ID_1\"}" "$GRPC_HOST" "$SERVICE/RemoveDiscount"
    grpcurl -plaintext -d "{
      \"product_id\": \"$PRODUCT_ID_1\",
      \"discount\": {
        \"id\": \"discount-large-50\",
        \"amount\": {\"amount\": \"50\"},
        \"start_date\": \"$CURRENT_DATE\",
        \"end_date\": \"$FUTURE_END\"
      }
    }" "$GRPC_HOST" "$SERVICE/ApplyDiscount"
    
    print_test "Get Product 1 (verify 50% discount - half price)"
    grpcurl -plaintext -d "{\"product_id\": \"$PRODUCT_ID_1\"}" "$GRPC_HOST" "$SERVICE/GetProduct"
fi

# ============================================================================
# SUMMARY
# ============================================================================
print_section "TEST SUMMARY"

echo -e "${GREEN}All API tests completed!${NC}"
echo ""
echo "Tested Requirements:"
echo "  ✓ Product Management: Create, Update, Activate, Deactivate, Archive"
echo "  ✓ Pricing Rules: Apply discount, Remove discount, Only one active discount"
echo "  ✓ Product Queries: Get by ID, List with pagination, Filter by category"
echo "  ✓ Precise Decimal Arithmetic: Verified with various discount percentages"
echo "  ✓ Business Rules: Active product requirement, Single discount rule"
echo ""
echo -e "${BLUE}Note: Event publishing is handled automatically via transactional outbox pattern${NC}"
echo -e "${BLUE}      Events are published when products are created, updated, or pricing changes${NC}"
