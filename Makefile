.PHONY: proto install-proto-tools migrate test test-e2e run emulator clean setup setup-proto check-protoc check-plugins help

# Default target
.DEFAULT_GOAL := help

# Help target
help:
	@echo "Available targets:"
	@echo "  make proto        - Generate Protocol Buffer code"
	@echo "  make migrate      - Run database migrations"
	@echo "  make test         - Run all tests"
	@echo "  make test-e2e     - Run only E2E tests"
	@echo "  make run          - Start the gRPC server"
	@echo "  make emulator     - Start Spanner emulator"
	@echo "  make clean        - Stop emulator and clean up"
	@echo "  make setup        - Full setup (proto tools, proto generation, emulator, migrations)"
	@echo ""
	@echo "Setup targets:"
	@echo "  make setup-proto  - Install and verify proto tools"
	@echo "  make check-protoc - Check if protoc is installed"

# Install required proto tools
install-proto-tools:
	@echo "Installing protoc-gen-go and protoc-gen-go-grpc..."
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "Tools installed. Make sure $(shell go env GOPATH)/bin is in your PATH"

# Generate proto code
proto:
	@echo "Generating proto code..."
	@export PATH=$$PATH:$$(go env GOPATH)/bin && \
	protoc \
		--go_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_out=. \
		--go-grpc_opt=paths=source_relative \
		--proto_path=. \
		proto/product/v1/product_service.proto
	@echo "Proto code generated successfully!"

# Check if protoc is installed
check-protoc:
	@which protoc > /dev/null || (echo "ERROR: protoc is not installed. Install it with:" && echo "  sudo apt-get install protobuf-compiler" && echo "  or visit https://grpc.io/docs/protoc-installation/" && exit 1)
	@echo "protoc found: $$(protoc --version)"

# Check if proto plugins are installed
check-plugins:
	@export PATH=$$PATH:$$(go env GOPATH)/bin && \
	which protoc-gen-go > /dev/null || (echo "ERROR: protoc-gen-go not found. Run: make install-proto-tools" && exit 1) && \
	which protoc-gen-go-grpc > /dev/null || (echo "ERROR: protoc-gen-go-grpc not found. Run: make install-proto-tools" && exit 1) && \
	echo "All proto plugins are installed"

# Full setup: check and install everything needed
setup-proto: check-protoc install-proto-tools check-plugins
	@echo "Proto setup complete!"

# Start Spanner emulator
emulator:
	@echo "Starting Spanner emulator..."
	@docker compose up -d
	@sleep 2
	@echo "Emulator started. Check status with: docker ps | grep spanner-emulator"

# Run database migrations
migrate:
	@echo "Running migrations..."
	@export SPANNER_EMULATOR_HOST=localhost:9010 && \
	go run cmd/server/main.go -migrate
	@echo "Migrations completed!"

# Run all tests
test:
	@echo "Running tests..."
	@export SPANNER_EMULATOR_HOST=localhost:9010 && \
	go test ./... -v
	@echo "Tests completed!"

# Run only E2E tests
test-e2e:
	@echo "Running E2E tests..."
	@export SPANNER_EMULATOR_HOST=localhost:9010 && \
	go test ./tests/e2e/... -v
	@echo "E2E tests completed!"

# Start the gRPC server
run:
	@echo "Starting gRPC server..."
	@echo "Make sure Spanner emulator is running (make emulator)"
	@echo "Make sure migrations are run (make migrate)"
	@export SPANNER_EMULATOR_HOST=localhost:9010 && \
	go run cmd/server/main.go

# Clean up (stop emulator)
clean:
	@echo "Stopping Spanner emulator..."
	@docker compose down
	@echo "Cleanup completed!"

# Full setup: install tools, generate proto, start emulator, run migrations
setup: setup-proto proto emulator
	@sleep 3
	@$(MAKE) migrate
	@echo ""
	@echo "Setup complete! You can now:"
	@echo "  - Run tests: make test"
	@echo "  - Start server: make run"
