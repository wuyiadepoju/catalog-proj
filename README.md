# Product Catalog Service

A production-ready Product Catalog Service implementing Domain-Driven Design (DDD) and Clean Architecture principles. Manages products and pricing with precise decimal arithmetic, event-driven architecture, and transactional guarantees.

## Prerequisites

- **Go 1.21+** - [Download](https://go.dev/dl/)
- **Docker & Docker Compose** - For Spanner emulator
- **Protocol Buffers compiler (`protoc`)** - `sudo apt-get install protobuf-compiler` (Ubuntu) or `brew install protobuf` (macOS)
- **Go proto plugins:**
  ```bash
  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
  export PATH=$PATH:$(go env GOPATH)/bin
  ```

## Quick Start

```bash
# 1. Start Spanner emulator
docker compose up -d

# 2. Generate proto code
make proto

# 3. Run migrations
make migrate

# 4. Start the server
make run
```

The gRPC server starts on port `50051` (default). Emulator available at `localhost:9010` (gRPC) and `localhost:9020` (HTTP).

## Testing

```bash
# Run all tests
make test

# Run E2E tests only
go test ./tests/e2e/... -v

# Run specific test
go test ./tests/e2e/... -v -run TestProductCreationFlow
```

**Test Coverage:** Product creation/update, discount application, activation/deactivation, business rule validation, outbox events, list/get queries.

## Project Structure

```
catalog-proj/
├── cmd/server/main.go                # Service entry point
├── internal/
│   ├── app/product/
│   │   ├── domain/                   # Pure domain (no external deps)
│   │   │   ├── product.go            # Product aggregate
│   │   │   ├── discount.go           # Discount value object
│   │   │   ├── money.go              # Money value object (*big.Rat)
│   │   │   ├── domain_events.go      # Domain events
│   │   │   ├── domain_errors.go      # Domain errors
│   │   │   └── services/pricing_calculator.go
│   │   ├── usecases/                 # Commands (create, update, activate, etc.)
│   │   ├── queries/                  # Queries (get, list)
│   │   ├── contracts/                # Repository interfaces
│   │   └── repo/                     # Spanner implementations
│   ├── models/                       # Database models (m_product, m_outbox)
│   ├── transport/grpc/product/       # gRPC handlers
│   ├── services/options.go           # Dependency injection
│   └── pkg/committer,clock/          # Shared utilities
├── proto/product/v1/                 # gRPC API definition
├── migrations/                       # Spanner DDL
└── tests/e2e/                        # E2E tests
```

## Architecture

Follows **Domain-Driven Design (DDD)** and **Clean Architecture** with strict layer separation:

1. **Domain Layer** - Pure business logic, no external dependencies, uses `*big.Rat` for money
2. **Application Layer** - Orchestrates domain logic, uses CommitPlan for transactions
3. **Infrastructure Layer** - Database, gRPC handlers, model facades

### Key Patterns

**Golden Mutation Pattern:** Every write operation: Load/Create → Domain method → Build plan → Get mutations → Add outbox events → Apply atomically

**CQRS:** Commands go through domain aggregates; queries bypass domain for performance

**Transactional Outbox:** Domain events stored in same transaction, ensuring reliable publishing

**Change Tracking:** Aggregates track dirty fields, repositories build targeted updates

## Design Decisions

1. **Money Storage:** Numerator/denominator INT64 columns for maximum precision (vs NUMERIC)
2. **Discount Storage:** `discount_amount` (NUMERIC) as percentage decimal - functionally equivalent to `discount_percent`
3. **Domain Purity:** No `context`, database, or proto imports in domain layer
4. **Change Tracking:** Manual tracking enables optimized database updates
5. **CQRS:** Queries bypass domain for performance; commands go through domain
6. **Event Enrichment:** Simple events in domain, enrichment in usecases
7. **Outbox Pattern:** Events stored transactionally; background processor out of scope

**Note:** Development migrations drop/recreate database. Production should use proper migration versioning.

## API Usage (grpcurl)

```bash
# Install: go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# Create product
grpcurl -plaintext -d '{"name":"Laptop","description":"High-performance","category":"electronics","base_price":{"amount":"99999"}}' localhost:50051 product.v1.ProductService/CreateProduct

# Get product
grpcurl -plaintext -d '{"product_id":"uuid"}' localhost:50051 product.v1.ProductService/GetProduct

# List products
grpcurl -plaintext -d '{"limit":10,"offset":0}' localhost:50051 product.v1.ProductService/ListProducts

# Apply discount
grpcurl -plaintext -d '{"product_id":"uuid","discount":{"id":"discount-1","amount":{"amount":"1000"},"start_date":"2024-01-01T00:00:00Z","end_date":"2024-12-31T23:59:59Z"}}' localhost:50051 product.v1.ProductService/ApplyDiscount
```

## Troubleshooting

- **protoc not found:** `sudo apt-get install protobuf-compiler` (Ubuntu) or `brew install protobuf` (macOS)
- **protoc-gen-go not found:** Install Go plugins (see Prerequisites)
- **Spanner connection failed:** Check `docker ps | grep spanner-emulator`, restart with `docker compose down && docker compose up -d`
- **Database not found:** Run `make migrate`

## Production Considerations

For production use, add: migration versioning, outbox processor, authentication/authorization, monitoring/metrics, externalized configuration, health checks, and connection pooling.
