#!/bin/bash
set -e

# Set Spanner emulator host
export SPANNER_EMULATOR_HOST=localhost:9010

# Default database
DATABASE="${1:-projects/test-project/instances/test-instance/databases/test-db}"

echo "Running migrations for database: $DATABASE"
echo "SPANNER_EMULATOR_HOST: $SPANNER_EMULATOR_HOST"

# Run migrations
go run cmd/server/main.go -migrate -spanner-database="$DATABASE"

echo "Migrations completed successfully!"
