# Product Catalog Service

A production-ready Product Catalog Service implementing Domain-Driven Design (DDD) and Clean Architecture principles. This service manages products and their pricing with precise decimal arithmetic, event-driven architecture, and transactional guarantees.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Setup Instructions](#setup-instructions)
- [Running the Service](#running-the-service)
- [Testing](#testing)
- [Project Structure](#project-structure)
- [Architecture](#architecture)
- [Design Decisions](#design-decisions)
- [Trade-offs](#trade-offs)

## Prerequisites

- **Go 1.21+** - [Download](https://go.dev/dl/)
- **Docker & Docker Compose** - For Spanner emulator
- **Protocol Buffers compiler (`protoc`)** - See setup instructions below
- **Go proto plugins** - See setup instructions below

## Quick Start

```bash
# 1. Start Spanner emulator
docker compose up -d

# 2. Generate proto code
make proto

# 3. Run migrations
make migrate

# 4. Run tests (optional, to verify setup)
make test

# 5. Start the server
make run
```

The gRPC server will start on port `50051` (default).

## Setup Instructions

### 1. Install Protocol Buffers Compiler

**Ubuntu/Debian:**
```bash
sudo apt-get update
sudo apt-get install -y protobuf-compiler
```

**macOS:**
```bash
brew install protobuf
```

**Or download from:** https://grpc.io/docs/protoc-installation/

### 2. Install Go Proto Plugins

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Make sure GOPATH/bin is in your PATH
export PATH=$PATH:$(go env GOPATH)/bin
```

### 3. Generate Proto Code

```bash
# Using Makefile (recommended)
make proto

# Or using the script
./scripts/generate-proto.sh
```

## Running the Service

### Start Spanner Emulator

The service uses Google Cloud Spanner emulator for local development:

```bash
# Start emulator
docker compose up -d

# Verify it's running
docker ps | grep spanner-emulator

# Check logs if needed
docker compose logs spanner-emulator
```

The emulator will be available at:
- **gRPC:** `localhost:9010`
- **HTTP:** `localhost:9020`

**Note:** Make sure Docker is running before starting the emulator.

### Run Migrations

Before starting the server, you need to create the database schema:

```bash
# Using Makefile (recommended)
make migrate

# Or using the script
./scripts/run-migrations.sh

# Or manually
export SPANNER_EMULATOR_HOST=localhost:9010
go run cmd/server/main.go -migrate
```

**Default database:** `projects/test-project/instances/test-instance/databases/test-db`

**Note:** For development, the migration script drops and recreates the database to ensure a clean schema. In production, you would use proper migration versioning.

### Start the Server

```bash
# Using Makefile (recommended)
make run

# Or manually
export SPANNER_EMULATOR_HOST=localhost:9010
go run cmd/server/main.go

# Or with custom port
go run cmd/server/main.go -grpc-port=50051
```

The server will:
- Connect to Spanner emulator (if `SPANNER_EMULATOR_HOST` is set)
- Start gRPC server on port `50051` (default)
- Enable gRPC reflection for tools like `grpcurl`

### Stop the Server

Press `Ctrl+C` for graceful shutdown.

## Testing

### Run All Tests

```bash
# Using Makefile (recommended)
make test

# Or manually
export SPANNER_EMULATOR_HOST=localhost:9010
go test ./...

# Run only E2E tests
go test ./tests/e2e/... -v

# Run specific test
go test ./tests/e2e/... -v -run TestProductCreationFlow
```

### Test Coverage

The E2E tests cover:
- ✅ Product creation flow
- ✅ Product update flow
- ✅ Discount application with price calculation
- ✅ Product activation/deactivation
- ✅ Business rule validation
- ✅ Outbox event creation
- ✅ List products with filters
- ✅ Get product with effective price

**Note:** Tests create isolated databases per test run, so they're safe to run in parallel.

## Project Structure

```
catalog-proj/
├── cmd/
│   └── server/
│       └── main.go                    # Service entry point with migrations
├── internal/
│   ├── app/
│   │   └── product/
│   │       ├── domain/                # Pure domain layer (no external deps)
│   │       │   ├── product.go         # Product aggregate
│   │       │   ├── discount.go        # Discount value object
│   │       │   ├── money.go           # Money value object (*big.Rat)
│   │       │   ├── domain_events.go   # Domain events
│   │       │   ├── domain_errors.go   # Domain errors
│   │       │   └── services/
│   │       │       └── pricing_calculator.go  # Domain service
│   │       ├── usecases/              # Application layer (commands)
│   │       │   ├── create_product/
│   │       │   ├── update_product/
│   │       │   ├── apply_discount/
│   │       │   ├── remove_discount/
│   │       │   ├── activate_product/
│   │       │   ├── deactivate_product/
│   │       │   └── archive_product/
│   │       ├── queries/               # Application layer (queries)
│   │       │   ├── get_product/
│   │       │   └── list_products/
│   │       ├── contracts/             # Repository interfaces
│   │       │   ├── product_repo.go
│   │       │   └── read_model.go
│   │       └── repo/                  # Infrastructure layer
│   │           ├── product_repo.go    # Spanner implementation
│   │           ├── read_model.go      # Read model implementation
│   │           └── outbox_repo.go
│   ├── models/                       # Database models
│   │   ├── m_product/
│   │   │   ├── data.go               # Product DB model
│   │   │   └── fields.go             # Field constants
│   │   └── m_outbox/
│   │       ├── data.go               # Outbox DB model
│   │       └── fields.go
│   ├── transport/
│   │   └── grpc/
│   │       └── product/              # gRPC handlers
│   │           ├── handler.go
│   │           ├── create.go
│   │           ├── update.go
│   │           ├── get.go
│   │           ├── list.go
│   │           ├── mappers.go
│   │           └── errors.go
│   ├── services/
│   │   └── options.go                # Dependency injection
│   └── pkg/
│       ├── committer/
│       │   ├── plan.go               # CommitPlan wrapper
│       │   └── committer.go          # Committer interface
│       └── clock/
│           └── clock.go              # Time abstraction
├── proto/
│   └── product/
│       └── v1/
│           └── product_service.proto  # gRPC API definition
├── migrations/
│   └── 001_initial_schema.sql        # Spanner DDL
├── tests/
│   └── e2e/
│       └── product_test.go           # E2E tests
├── scripts/
│   ├── generate-proto.sh             # Proto generation script
│   └── run-migrations.sh             # Migration script
├── docker-compose.yml                # Spanner emulator setup
├── Makefile                          # Build automation
├── go.mod
├── go.sum
└── README.md
```

## Architecture

This service follows **Domain-Driven Design (DDD)** and **Clean Architecture** principles with strict layer separation.

### Layer Separation

1. **Domain Layer** (`internal/app/product/domain/`)
   - Pure business logic
   - No external dependencies (no `context`, no database, no proto)
   - Uses `*big.Rat` for precise money calculations
   - Encapsulated aggregates with change tracking
   - Domain events as simple structs

2. **Application Layer** (`internal/app/product/usecases/` & `queries/`)
   - Orchestrates domain logic
   - Implements use cases (commands) and queries
   - Uses CommitPlan for atomic transactions
   - Enriches domain events with metadata

3. **Infrastructure Layer** (`internal/app/product/repo/`, `models/`, `transport/`)
   - Database implementations
   - gRPC handlers
   - Model facades

### Key Patterns

#### 1. Golden Mutation Pattern

Every write operation follows this pattern:

```go
// 1. Load or create aggregate
product := domain.NewProduct(...)

// 2. Call domain method
product.ApplyDiscount(discount, now)

// 3. Build commit plan
plan := commitplan.NewPlan()

// 4. Get mutations from repository (repo returns, doesn't apply)
plan.Add(repo.UpdateMut(product))

// 5. Add outbox events
for _, event := range product.DomainEvents() {
    plan.Add(outboxRepo.InsertMut(enrichEvent(event)))
}

// 6. Apply plan atomically (usecase applies, NOT handler!)
committer.Apply(ctx, plan)
```

#### 2. CQRS (Command Query Responsibility Segregation)

- **Commands:** Go through domain aggregates, use CommitPlan
- **Queries:** Bypass domain for optimization, use read models directly

#### 3. Transactional Outbox Pattern

- Domain events captured as simple structs
- Events enriched with metadata in usecases
- Events stored in `outbox_events` table within same transaction
- Ensures reliable event publishing

#### 4. Change Tracking

- Aggregates track dirty fields
- Repositories build targeted updates based on change tracker
- Optimizes database writes (only update changed fields)

## Design Decisions

### 1. Money Storage: Numerator/Denominator vs NUMERIC

**Decision:** Store money as separate `base_price_numerator` and `base_price_denominator` INT64 columns.

**Rationale:**
- Provides maximum precision for financial calculations
- Avoids floating-point rounding errors
- Allows exact representation of any rational number
- Aligns with requirement to use `math/big` for calculations

**Trade-off:** Slightly more complex schema, but ensures precision.

### 2. Discount Storage: `discount_amount` vs `discount_percent`

**Decision:** Store discount as `discount_amount` (NUMERIC) representing percentage as decimal.

**Rationale:**
- Consistent with domain model where `Discount.Amount` is `*Money` (percentage as decimal)
- Allows for future flexibility (could represent fixed amounts)
- Functionally equivalent to `discount_percent`

**Note:** The field name differs from the requirement (`discount_percent`), but the implementation is functionally equivalent.

### 3. Database Recreation for Migrations

**Decision:** Drop and recreate database during development migrations.

**Rationale:**
- Simplifies development workflow
- Ensures clean schema state
- Appropriate for emulator environment

**Production Note:** In production, you would implement proper migration versioning instead of drop/recreate.

### 4. Domain Layer Purity

**Decision:** Strict enforcement of domain layer purity - no `context`, no database, no proto imports.

**Rationale:**
- Ensures domain logic is testable in isolation
- Prevents infrastructure concerns from leaking into business logic
- Makes domain logic portable and framework-agnostic

### 5. Structured Logging

**Decision:** Use `log/slog` for structured logging instead of standard `log` package.

**Rationale:**
- Better observability with key-value pairs
- Easier integration with log aggregation tools
- Industry best practice for production systems

### 6. Single Active Discount Rule

**Decision:** Enforce "only one active discount per product" in domain layer.

**Rationale:**
- Business rule belongs in domain
- Prevents invalid state at the source
- Clear error message when violated

## Trade-offs

### 1. Query Performance vs Domain Purity

**Trade-off:** Queries bypass domain layer for performance.

**Rationale:**
- Read operations don't need business logic validation
- Direct database access is faster
- Still uses domain services (PricingCalculator) for calculations

**Alternative Considered:** Always go through domain - rejected due to performance overhead.

### 2. Change Tracking Complexity

**Trade-off:** Manual change tracking adds complexity.

**Rationale:**
- Enables optimized database updates (only changed fields)
- Reduces database write operations
- Worth the complexity for production systems

**Alternative Considered:** Always update all fields - rejected due to inefficiency.

### 3. Event Enrichment in Usecases

**Trade-off:** Domain events are simple, enrichment happens in usecases.

**Rationale:**
- Keeps domain pure (no infrastructure concerns)
- Allows adding metadata (user_id, company_id) without polluting domain
- Follows separation of concerns

**Alternative Considered:** Rich events in domain - rejected to maintain domain purity.

### 4. No Background Outbox Processor

**Trade-off:** Events are stored but not automatically published.

**Rationale:**
- Out of scope for this task
- Events are reliably stored (transactional guarantee)
- Can be processed by separate service/worker

**Future Enhancement:** Add background worker to process outbox events and publish to Pub/Sub.

### 5. Money Type Alias

**Trade-off:** `type Money *big.Rat` instead of struct wrapper.

**Rationale:**
- Simpler type system
- Direct access to `*big.Rat` methods when needed
- Type conversion is straightforward

**Alternative Considered:** Struct wrapper with methods - rejected for simplicity.

## API Usage Examples

### Using grpcurl

```bash
# Install grpcurl
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# List services
grpcurl -plaintext localhost:50051 list

# Create a product
grpcurl -plaintext -d '{
  "name": "Laptop",
  "description": "High-performance laptop",
  "category": "electronics",
  "base_price": {
    "amount": "99999"
  }
}' localhost:50051 product.v1.ProductService/CreateProduct

# Get a product
grpcurl -plaintext -d '{
  "product_id": "product-uuid"
}' localhost:50051 product.v1.ProductService/GetProduct

# List products
grpcurl -plaintext -d '{
  "limit": 10,
  "offset": 0
}' localhost:50051 product.v1.ProductService/ListProducts

# Apply discount
grpcurl -plaintext -d '{
  "product_id": "product-uuid",
  "discount": {
    "id": "discount-1",
    "amount": {
      "amount": "1000"
    },
    "start_date": "2024-01-01T00:00:00Z",
    "end_date": "2024-12-31T23:59:59Z"
  }
}' localhost:50051 product.v1.ProductService/ApplyDiscount
```

## Troubleshooting

### Issue: "protoc not found"
```bash
# Install protoc
sudo apt-get install protobuf-compiler  # Ubuntu/Debian
brew install protobuf                    # macOS
```

### Issue: "protoc-gen-go not found"
```bash
# Install Go plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
export PATH=$PATH:$(go env GOPATH)/bin
```

### Issue: "Spanner emulator connection failed"
```bash
# Check if emulator is running
docker ps | grep spanner-emulator

# Restart emulator
docker compose down
docker compose up -d
```

### Issue: "Database not found"
```bash
# Run migrations
make migrate
```

### Issue: "commitplan package not found"
The `commitplan` package is in a local directory (`../commitplan`). Make sure it exists and is properly set up.

## Development Workflow

1. **Start emulator:** `docker compose up -d`
2. **Generate proto:** `make proto` (after proto changes)
3. **Run migrations:** `make migrate` (after schema changes)
4. **Run tests:** `make test` (before committing)
5. **Start server:** `make run` (for manual testing)

## Production Considerations

This implementation is designed for the test task. For production use, consider:

1. **Migration Versioning:** Implement proper migration versioning instead of drop/recreate
2. **Outbox Processor:** Add background worker to process and publish outbox events
3. **Authentication/Authorization:** Add security layers
4. **Monitoring:** Add metrics, tracing, and alerting
5. **Configuration:** Externalize configuration (database connection, ports, etc.)
6. **Health Checks:** Add health check endpoints
7. **Graceful Shutdown:** Already implemented
8. **Connection Pooling:** Configure Spanner connection pools appropriately

## License

This is a test task implementation.

## Author

Product Catalog Service - Test Task Implementation
# catalog-proj
