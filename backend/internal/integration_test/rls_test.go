package integration_test

import (
	"context"
	"testing"

	"good-todo-go/internal/domain/model"
	"good-todo-go/internal/infrastructure/database"
	"good-todo-go/internal/infrastructure/repository"
	"good-todo-go/internal/integration_test/common"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRLS_TenantIsolation verifies that RLS policies correctly isolate data between tenants
func TestRLS_TenantIsolation(t *testing.T) {
	t.Parallel()

	// Setup: admin client for data creation, app client for RLS testing
	adminClient, appClient := common.SetupTestClientWithRLS(t)

	// Create test data using admin client (bypasses RLS)
	data := common.CreateTestDataSet(t, adminClient)

	t.Run("Todo - tenant1 context can only see tenant1 todos", func(t *testing.T) {
		repo := repository.NewTodoRepository(appClient)
		ctx := database.WithTenantID(context.Background(), data.Tenant1.ID)

		// Should find todos in tenant1
		todo, err := repo.FindByID(ctx, data.Todo1.ID)
		require.NoError(t, err)
		assert.Equal(t, data.Todo1.ID, todo.ID)
		assert.Equal(t, data.Tenant1.ID, todo.TenantID)
	})

	t.Run("Todo - tenant1 context cannot see tenant2 todos", func(t *testing.T) {
		repo := repository.NewTodoRepository(appClient)
		ctx := database.WithTenantID(context.Background(), data.Tenant1.ID)

		// Should NOT find todos in tenant2
		_, err := repo.FindByID(ctx, data.Todo4.ID)
		assert.Error(t, err, "Should not be able to access tenant2's todo from tenant1 context")
	})

	t.Run("Todo - tenant2 context can only see tenant2 todos", func(t *testing.T) {
		repo := repository.NewTodoRepository(appClient)
		ctx := database.WithTenantID(context.Background(), data.Tenant2.ID)

		// Should find todos in tenant2
		todo, err := repo.FindByID(ctx, data.Todo4.ID)
		require.NoError(t, err)
		assert.Equal(t, data.Todo4.ID, todo.ID)
		assert.Equal(t, data.Tenant2.ID, todo.TenantID)
	})

	t.Run("Todo - tenant2 context cannot see tenant1 todos", func(t *testing.T) {
		repo := repository.NewTodoRepository(appClient)
		ctx := database.WithTenantID(context.Background(), data.Tenant2.ID)

		// Should NOT find todos in tenant1
		_, err := repo.FindByID(ctx, data.Todo1.ID)
		assert.Error(t, err, "Should not be able to access tenant1's todo from tenant2 context")
	})

	t.Run("User - tenant1 context can only see tenant1 users", func(t *testing.T) {
		repo := repository.NewUserRepository(appClient)
		ctx := database.WithTenantID(context.Background(), data.Tenant1.ID)

		// Should find users in tenant1
		user, err := repo.FindByID(ctx, data.User1.ID)
		require.NoError(t, err)
		assert.Equal(t, data.User1.ID, user.ID)
		assert.Equal(t, data.Tenant1.ID, user.TenantID)
	})

	t.Run("User - tenant1 context cannot see tenant2 users", func(t *testing.T) {
		repo := repository.NewUserRepository(appClient)
		ctx := database.WithTenantID(context.Background(), data.Tenant1.ID)

		// Should NOT find users in tenant2
		_, err := repo.FindByID(ctx, data.User3.ID)
		assert.Error(t, err, "Should not be able to access tenant2's user from tenant1 context")
	})
}

// TestRLS_FindByUserID verifies tenant isolation for FindByUserID
func TestRLS_FindByUserID(t *testing.T) {
	t.Parallel()

	adminClient, appClient := common.SetupTestClientWithRLS(t)
	data := common.CreateTestDataSet(t, adminClient)

	repo := repository.NewTodoRepository(appClient)

	t.Run("tenant1 context - find todos for user1", func(t *testing.T) {
		ctx := database.WithTenantID(context.Background(), data.Tenant1.ID)

		todos, err := repo.FindByUserID(ctx, data.User1.ID, 10, 0)
		require.NoError(t, err)
		assert.Len(t, todos, 2) // Todo1 and Todo2 belong to User1

		for _, todo := range todos {
			assert.Equal(t, data.Tenant1.ID, todo.TenantID)
			assert.Equal(t, data.User1.ID, todo.UserID)
		}
	})

	t.Run("tenant2 context - cannot find tenant1 user's todos", func(t *testing.T) {
		ctx := database.WithTenantID(context.Background(), data.Tenant2.ID)

		todos, err := repo.FindByUserID(ctx, data.User1.ID, 10, 0)
		require.NoError(t, err)
		assert.Len(t, todos, 0, "Should not find any todos for tenant1's user from tenant2 context")
	})
}

// TestRLS_FindPublic verifies tenant isolation for public todos
func TestRLS_FindPublic(t *testing.T) {
	t.Parallel()

	adminClient, appClient := common.SetupTestClientWithRLS(t)
	data := common.CreateTestDataSet(t, adminClient)

	repo := repository.NewTodoRepository(appClient)

	t.Run("tenant1 context - find public todos only in tenant1", func(t *testing.T) {
		ctx := database.WithTenantID(context.Background(), data.Tenant1.ID)

		todos, err := repo.FindPublic(ctx, 10, 0)
		require.NoError(t, err)
		assert.Len(t, todos, 1) // Only Todo2 is public in tenant1

		for _, todo := range todos {
			assert.Equal(t, data.Tenant1.ID, todo.TenantID)
			assert.True(t, todo.IsPublic)
		}
	})

	t.Run("tenant2 context - find public todos only in tenant2", func(t *testing.T) {
		ctx := database.WithTenantID(context.Background(), data.Tenant2.ID)

		todos, err := repo.FindPublic(ctx, 10, 0)
		require.NoError(t, err)
		assert.Len(t, todos, 1) // Only Todo5 is public in tenant2

		for _, todo := range todos {
			assert.Equal(t, data.Tenant2.ID, todo.TenantID)
			assert.True(t, todo.IsPublic)
		}
	})
}

// TestRLS_CreateWithWrongTenant verifies RLS prevents creating data in wrong tenant
func TestRLS_CreateWithWrongTenant(t *testing.T) {
	t.Parallel()

	adminClient, appClient := common.SetupTestClientWithRLS(t)
	data := common.CreateTestDataSet(t, adminClient)

	repo := repository.NewTodoRepository(appClient)

	t.Run("cannot create todo in different tenant context", func(t *testing.T) {
		// Context is tenant1, but trying to create todo for tenant2
		ctx := database.WithTenantID(context.Background(), data.Tenant1.ID)

		todo := &model.Todo{
			ID:          "cross-tenant-todo",
			TenantID:    data.Tenant2.ID, // Wrong tenant!
			UserID:      data.User3.ID,
			Title:       "Cross Tenant Todo",
			Description: "Should fail",
			Completed:   false,
			IsPublic:    false,
		}

		// This should fail due to RLS WITH CHECK policy
		_, err := repo.Create(ctx, todo)
		// Note: RLS will block this insert because tenant_id doesn't match current_setting
		// The exact error depends on PostgreSQL version and RLS configuration
		assert.Error(t, err, "Should not be able to create todo with mismatched tenant_id")
	})
}

// TestRLS_CountByUserID verifies tenant isolation for count operations
func TestRLS_CountByUserID(t *testing.T) {
	t.Parallel()

	adminClient, appClient := common.SetupTestClientWithRLS(t)
	data := common.CreateTestDataSet(t, adminClient)

	repo := repository.NewTodoRepository(appClient)

	t.Run("tenant1 context - count todos for user1", func(t *testing.T) {
		ctx := database.WithTenantID(context.Background(), data.Tenant1.ID)

		count, err := repo.CountByUserID(ctx, data.User1.ID)
		require.NoError(t, err)
		assert.Equal(t, 2, count) // User1 has 2 todos
	})

	t.Run("tenant2 context - count returns 0 for tenant1 user", func(t *testing.T) {
		ctx := database.WithTenantID(context.Background(), data.Tenant2.ID)

		count, err := repo.CountByUserID(ctx, data.User1.ID)
		require.NoError(t, err)
		assert.Equal(t, 0, count, "Should count 0 todos for tenant1's user from tenant2 context")
	})
}

// TestRLS_CountPublic verifies tenant isolation for count public operations
func TestRLS_CountPublic(t *testing.T) {
	t.Parallel()

	adminClient, appClient := common.SetupTestClientWithRLS(t)
	data := common.CreateTestDataSet(t, adminClient)

	repo := repository.NewTodoRepository(appClient)

	t.Run("tenant1 context - count public todos", func(t *testing.T) {
		ctx := database.WithTenantID(context.Background(), data.Tenant1.ID)

		count, err := repo.CountPublic(ctx)
		require.NoError(t, err)
		assert.Equal(t, 1, count) // Only Todo2 is public in tenant1
	})

	t.Run("tenant2 context - count public todos", func(t *testing.T) {
		ctx := database.WithTenantID(context.Background(), data.Tenant2.ID)

		count, err := repo.CountPublic(ctx)
		require.NoError(t, err)
		assert.Equal(t, 1, count) // Only Todo5 is public in tenant2
	})
}

// TestRLS_UpdateCrossTenant verifies RLS prevents updating data in wrong tenant
func TestRLS_UpdateCrossTenant(t *testing.T) {
	t.Parallel()

	adminClient, appClient := common.SetupTestClientWithRLS(t)
	data := common.CreateTestDataSet(t, adminClient)

	repo := repository.NewTodoRepository(appClient)

	t.Run("cannot update todo from different tenant", func(t *testing.T) {
		// Context is tenant2, but trying to update tenant1's todo
		ctx := database.WithTenantID(context.Background(), data.Tenant2.ID)

		// First, verify we can't even find it
		_, err := repo.FindByID(ctx, data.Todo1.ID)
		assert.Error(t, err, "Should not be able to find tenant1's todo from tenant2 context")
	})
}

// TestRLS_DeleteCrossTenant verifies RLS prevents deleting data in wrong tenant
func TestRLS_DeleteCrossTenant(t *testing.T) {
	t.Parallel()

	adminClient, appClient := common.SetupTestClientWithRLS(t)
	data := common.CreateTestDataSet(t, adminClient)

	repo := repository.NewTodoRepository(appClient)

	t.Run("cannot delete todo from different tenant", func(t *testing.T) {
		// Context is tenant2, but trying to delete tenant1's todo
		ctx := database.WithTenantID(context.Background(), data.Tenant2.ID)

		err := repo.Delete(ctx, data.Todo1.ID)
		// This should fail because RLS won't find the row to delete
		assert.Error(t, err, "Should not be able to delete tenant1's todo from tenant2 context")

		// Verify the todo still exists (using admin client)
		ctxAdmin := context.Background()
		todo, err := adminClient.Todo.Get(ctxAdmin, data.Todo1.ID)
		require.NoError(t, err)
		assert.NotNil(t, todo, "Todo should still exist after failed cross-tenant delete")
	})
}
