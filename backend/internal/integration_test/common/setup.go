package common

import (
	"context"
	"database/sql"
	"testing"

	"good-todo-go/internal/ent"
	"good-todo-go/internal/infrastructure/repository/test"

	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/lib/pq"
)

// SetupTestClient creates a new test ent client with PostgreSQL testcontainer
// Uses Atlas migrations for schema setup
func SetupTestClient(t *testing.T) *ent.Client {
	t.Helper()

	ctx := context.Background()

	// Create new PostgreSQL container for each test
	pgContainer, err := test.NewPostgresContainer(ctx)
	if err != nil {
		t.Fatalf("failed to create postgres container: %v", err)
	}

	// Cleanup on test end
	t.Cleanup(func() {
		if pgContainer != nil {
			pgContainer.Close(ctx)
		}
	})

	// Run Atlas migrations
	if err := pgContainer.RunMigrations(ctx); err != nil {
		pgContainer.Close(ctx)
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Create ent client with PostgreSQL
	drv := entsql.OpenDB("postgres", pgContainer.DB)
	client := ent.NewClient(ent.Driver(drv))

	return client
}

// SetupTestClientWithRLS creates a new test ent client with RLS enabled
// Returns both admin client (for setup) and app client (RLS enforced)
func SetupTestClientWithRLS(t *testing.T) (adminClient *ent.Client, appClient *ent.Client) {
	t.Helper()

	ctx := context.Background()

	// Create new PostgreSQL container
	pgContainer, err := test.NewPostgresContainer(ctx)
	if err != nil {
		t.Fatalf("failed to create postgres container: %v", err)
	}

	// Run Atlas migrations (includes RLS setup and app user creation)
	if err := pgContainer.RunMigrations(ctx); err != nil {
		pgContainer.Close(ctx)
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Create admin ent client
	adminDrv := entsql.OpenDB("postgres", pgContainer.DB)
	adminClient = ent.NewClient(ent.Driver(adminDrv))

	// Get app user DSN
	appDSN, err := pgContainer.GetAppUserDSN(ctx)
	if err != nil {
		adminClient.Close()
		pgContainer.Close(ctx)
		t.Fatalf("failed to get app user DSN: %v", err)
	}

	// Connect as app user (RLS enforced)
	appDB, err := sql.Open("postgres", appDSN)
	if err != nil {
		adminClient.Close()
		pgContainer.Close(ctx)
		t.Fatalf("failed to open app db: %v", err)
	}

	appDrv := entsql.OpenDB("postgres", appDB)
	appClient = ent.NewClient(ent.Driver(appDrv))

	// Cleanup on test end
	t.Cleanup(func() {
		if appClient != nil {
			appClient.Close()
		}
		if appDB != nil {
			appDB.Close()
		}
		if adminClient != nil {
			adminClient.Close()
		}
		if pgContainer != nil {
			pgContainer.Close(ctx)
		}
	})

	return adminClient, appClient
}
