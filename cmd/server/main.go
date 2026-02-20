package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"catalog-proj/internal/services"
	pb "catalog-proj/proto/product/v1"

	admin "cloud.google.com/go/spanner/admin/database/apiv1"
	"cloud.google.com/go/spanner/admin/database/apiv1/databasepb"
	instanceadmin "cloud.google.com/go/spanner/admin/instance/apiv1"
	"cloud.google.com/go/spanner/admin/instance/apiv1/instancepb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

var (
	spannerDatabase     = flag.String("spanner-database", "", "Spanner database (format: projects/{project}/instances/{instance}/databases/{database})")
	grpcPort            = flag.String("grpc-port", "50051", "gRPC server port")
	shouldRunMigrations = flag.Bool("migrate", false, "Run database migrations")
)

func main() {
	flag.Parse()

	ctx := context.Background()

	// Default database for emulator if not provided
	if *spannerDatabase == "" {
		// Check if using emulator
		if os.Getenv("SPANNER_EMULATOR_HOST") != "" {
			// Default emulator database
			*spannerDatabase = "projects/test-project/instances/test-instance/databases/test-db"
			slog.Info("Using Spanner emulator", "database", *spannerDatabase)
		} else {
			slog.Error("spanner-database flag is required (or set SPANNER_EMULATOR_HOST for emulator)")
			os.Exit(1)
		}
	}

	// Run migrations if requested
	if *shouldRunMigrations {
		if err := runMigrations(ctx, *spannerDatabase); err != nil {
			slog.Error("Failed to run migrations", "error", err)
			os.Exit(1)
		}
		slog.Info("Migrations completed successfully")
		return
	}

	// Create DI container
	opts, err := services.NewOptions(ctx, *spannerDatabase)
	if err != nil {
		slog.Error("Failed to create service options", "error", err)
		os.Exit(1)
	}
	defer opts.Close()

	// Register gRPC service
	pb.RegisterProductServiceServer(opts.GRPCServer, opts.ProductHandler)

	// Enable gRPC reflection for tools like grpcurl
	reflection.Register(opts.GRPCServer)

	// Start gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", *grpcPort))
	if err != nil {
		slog.Error("Failed to listen on port", "port", *grpcPort, "error", err)
		os.Exit(1)
	}

	slog.Info("Starting gRPC server", "port", *grpcPort, "database", *spannerDatabase)

	// Graceful shutdown
	go func() {
		if err := opts.GRPCServer.Serve(lis); err != nil {
			slog.Error("Failed to serve gRPC server", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down server...")
	opts.GRPCServer.GracefulStop()
	slog.Info("Server stopped")
}

// runMigrations runs database migrations
func runMigrations(ctx context.Context, database string) error {
	// Parse database string to extract components
	// Format: projects/{project}/instances/{instance}/databases/{database}
	parts := strings.Split(database, "/")
	if len(parts) != 6 || parts[0] != "projects" || parts[2] != "instances" || parts[4] != "databases" {
		return fmt.Errorf("invalid database format: %s (expected: projects/{project}/instances/{instance}/databases/{database})", database)
	}
	project := parts[1]
	instance := parts[3]
	dbName := parts[5]
	projectName := fmt.Sprintf("projects/%s", project)
	instanceName := fmt.Sprintf("projects/%s/instances/%s", project, instance)

	// Create instance admin client to check/create instance
	instanceAdminClient, err := instanceadmin.NewInstanceAdminClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create instance admin client: %w", err)
	}
	defer instanceAdminClient.Close()

	// Check if instance exists, create if it doesn't
	_, err = instanceAdminClient.GetInstance(ctx, &instancepb.GetInstanceRequest{
		Name: instanceName,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			slog.Info("Instance does not exist, creating", "instance", instanceName)
			// For emulator, create instance with minimal config
			op, err := instanceAdminClient.CreateInstance(ctx, &instancepb.CreateInstanceRequest{
				Parent:     projectName,
				InstanceId: instance,
				Instance: &instancepb.Instance{
					DisplayName: instance,
				},
			})
			if err != nil {
				return fmt.Errorf("failed to create instance: %w", err)
			}

			// Wait for instance creation
			_, err = op.Wait(ctx)
			if err != nil {
				return fmt.Errorf("instance creation failed: %w", err)
			}
			slog.Info("Successfully created instance", "instance", instanceName)
		} else {
			return fmt.Errorf("failed to check instance existence: %w", err)
		}
	}

	// Create database admin client for DDL operations
	adminClient, err := admin.NewDatabaseAdminClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create database admin client: %w", err)
	}
	defer adminClient.Close()

	// Read migration file
	migrationSQL, err := os.ReadFile("migrations/001_initial_schema.sql")
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	// Split SQL into individual statements (split by semicolon, but handle comments)
	statements := parseDDLStatements(string(migrationSQL))

	// Check if database exists
	_, err = adminClient.GetDatabase(ctx, &databasepb.GetDatabaseRequest{
		Name: database,
	})

	if err != nil {
		// Database doesn't exist, create it with DDL statements
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			slog.Info("Database does not exist, creating", "database", database)
			op, err := adminClient.CreateDatabase(ctx, &databasepb.CreateDatabaseRequest{
				Parent:          instanceName,
				CreateStatement: fmt.Sprintf("CREATE DATABASE `%s`", dbName),
				ExtraStatements: statements,
			})
			if err != nil {
				return fmt.Errorf("failed to create database: %w", err)
			}

			// Wait for database creation
			db, err := op.Wait(ctx)
			if err != nil {
				return fmt.Errorf("database creation failed: %w", err)
			}
			slog.Info("Successfully created database", "database", db.Name)
			slog.Info("Successfully applied migrations to database", "database", database)
			return nil
		}
		return fmt.Errorf("failed to check database existence: %w", err)
	}

	// Database exists - for emulator, drop and recreate for clean migrations
	// In production, you'd want to use proper migration versioning
	slog.Info("Database exists, dropping and recreating for clean migration", "database", database)

	// Drop the database
	if err := adminClient.DropDatabase(ctx, &databasepb.DropDatabaseRequest{
		Database: database,
	}); err != nil {
		slog.Warn("Failed to drop database (may not exist or already dropped)", "error", err)
	} else {
		slog.Info("Successfully dropped database")
	}

	// Recreate database with migrations
	slog.Info("Creating database with migrations", "database", database)
	createOp, err := adminClient.CreateDatabase(ctx, &databasepb.CreateDatabaseRequest{
		Parent:          instanceName,
		CreateStatement: fmt.Sprintf("CREATE DATABASE `%s`", dbName),
		ExtraStatements: statements,
	})
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}

	// Wait for database creation
	db, err := createOp.Wait(ctx)
	if err != nil {
		return fmt.Errorf("database creation failed: %w", err)
	}
	slog.Info("Successfully created database", "database", db.Name)
	slog.Info("Successfully applied migrations to database", "database", database)
	return nil
}

// parseDDLStatements parses SQL file into individual DDL statements
func parseDDLStatements(sql string) []string {
	var statements []string
	var currentStatement strings.Builder

	lines := strings.Split(sql, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and full-line comments
		if trimmed == "" || strings.HasPrefix(trimmed, "--") {
			continue
		}

		// Add line to current statement
		if currentStatement.Len() > 0 {
			currentStatement.WriteString(" ")
		}
		currentStatement.WriteString(trimmed)

		// If line ends with semicolon, finalize the statement
		if strings.HasSuffix(trimmed, ";") {
			stmt := strings.TrimSpace(currentStatement.String())
			// Remove trailing semicolon
			stmt = strings.TrimSuffix(stmt, ";")
			if stmt != "" {
				statements = append(statements, stmt)
			}
			currentStatement.Reset()
		}
	}

	// Handle any remaining statement without trailing semicolon
	if currentStatement.Len() > 0 {
		stmt := strings.TrimSpace(currentStatement.String())
		if stmt != "" {
			statements = append(statements, stmt)
		}
	}

	return statements
}
