package database

import (
	"context"
	"database/sql"
	"fmt"

	"good-todo-go/internal/ent"

	entsql "entgo.io/ent/dialect/sql"
)

// TenantContextKey is the context key for tenant ID
type tenantContextKey struct{}

// WithTenantID adds tenant ID to context
func WithTenantID(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, tenantContextKey{}, tenantID)
}

// GetTenantID gets tenant ID from context
func GetTenantID(ctx context.Context) (string, bool) {
	tenantID, ok := ctx.Value(tenantContextKey{}).(string)
	return tenantID, ok
}

// SetTenantContext sets the PostgreSQL session variable for RLS
// This must be called at the beginning of each request/transaction
func SetTenantContext(ctx context.Context, db *sql.DB, tenantID string) error {
	query := fmt.Sprintf("SET app.current_tenant_id = '%s'", tenantID)
	_, err := db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to set tenant context: %w", err)
	}
	return nil
}

// WithTenantTx executes a function within a transaction with tenant context set
// This is the recommended way to use RLS with connection pooling
// IMPORTANT: SET LOCAL is used to ensure the setting is scoped to the transaction only
func WithTenantTx(ctx context.Context, db *sql.DB, tenantID string, fn func(tx *sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	// Set tenant context within transaction using SET LOCAL
	// SET LOCAL ensures the setting is only for this transaction (connection pool safe)
	query := fmt.Sprintf("SET LOCAL app.current_tenant_id = '%s'", tenantID)
	if _, err := tx.ExecContext(ctx, query); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to set tenant context: %w", err)
	}

	// Execute the function
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

// ClearTenantContext clears the tenant context (for safety)
func ClearTenantContext(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, "RESET app.current_tenant_id")
	if err != nil {
		return fmt.Errorf("failed to clear tenant context: %w", err)
	}
	return nil
}

// NewEntClientWithDB creates an ent client from an existing sql.DB
// This allows us to share the same connection pool
func NewEntClientWithDB(db *sql.DB) *ent.Client {
	drv := entsql.OpenDB("postgres", db)
	return ent.NewClient(ent.Driver(drv))
}

// WithTenantScope starts a transaction with tenant context set and returns ent.Tx
// Usage:
//
//	tx, err := database.WithTenantScope(ctx, client, tenantID)
//	if err != nil { return err }
//	defer tx.Rollback()
//	todos, err := tx.TenantTodoView.Query().All(ctx)
//	return tx.Commit()
func WithTenantScope(ctx context.Context, client *ent.Client, tenantID string) (*ent.Tx, error) {
	tx, err := client.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}

	// Set tenant context within transaction using SET LOCAL
	if _, err := tx.ExecContext(ctx, "SET LOCAL app.current_tenant_id = $1", tenantID); err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	return tx, nil
}

// TenantScopedTx automatically gets tenantID from context and starts a scoped transaction
// Usage:
//
//	tx, err := database.TenantScopedTx(ctx, client)
//	if err != nil { return err }
//	defer tx.Rollback()
//	todos, err := tx.TenantTodoView.Query().All(ctx)
//	return tx.Commit()
func TenantScopedTx(ctx context.Context, client *ent.Client) (*ent.Tx, error) {
	tenantID, ok := GetTenantID(ctx)
	if !ok || tenantID == "" {
		return nil, fmt.Errorf("tenant ID not found in context")
	}
	return WithTenantScope(ctx, client, tenantID)
}
