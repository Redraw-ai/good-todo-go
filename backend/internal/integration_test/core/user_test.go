package integration_test

import (
	"context"
	"testing"

	"good-todo-go/internal/infrastructure/database"
	"good-todo-go/internal/infrastructure/repository"
	"good-todo-go/internal/integration_test/common"
	"good-todo-go/internal/usecase"
	"good-todo-go/internal/usecase/input"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserInteractor_GetMe(t *testing.T) {
	t.Parallel()

	// Setup
	client := common.SetupTestClient(t)
	ctx := context.Background()

	// Create test data
	dataSet := common.CreateTestDataSet(t, client)

	// Build dependencies
	userRepo := repository.NewUserRepository(client)
	userInteractor := usecase.NewUserInteractor(userRepo)

	// Set tenant context
	ctxWithTenant := database.WithTenantID(ctx, dataSet.Tenant1.ID)

	// Get user
	user, err := userInteractor.GetMe(ctxWithTenant, dataSet.User1.ID)
	require.NoError(t, err)
	assert.Equal(t, dataSet.User1.ID, user.ID)
	assert.Equal(t, dataSet.User1.Email, user.Email)
	assert.Equal(t, dataSet.User1.Name, user.Name)
}

func TestUserInteractor_UpdateMe(t *testing.T) {
	t.Parallel()

	// Setup
	client := common.SetupTestClient(t)
	ctx := context.Background()

	// Create test data
	dataSet := common.CreateTestDataSet(t, client)

	// Build dependencies
	userRepo := repository.NewUserRepository(client)
	userInteractor := usecase.NewUserInteractor(userRepo)

	// Set tenant context
	ctxWithTenant := database.WithTenantID(ctx, dataSet.Tenant1.ID)

	// Update user
	newName := "Updated Name"
	updated, err := userInteractor.UpdateMe(ctxWithTenant, &input.UpdateUserInput{
		UserID: dataSet.User1.ID,
		Name:   &newName,
	})
	require.NoError(t, err)
	assert.Equal(t, newName, updated.Name)

	// Verify the update persisted
	user, err := userInteractor.GetMe(ctxWithTenant, dataSet.User1.ID)
	require.NoError(t, err)
	assert.Equal(t, newName, user.Name)
}

func TestUserInteractor_GetMe_NotFound(t *testing.T) {
	t.Parallel()

	// Setup
	client := common.SetupTestClient(t)
	ctx := context.Background()

	// Create test data (need tenant for context)
	dataSet := common.CreateTestDataSet(t, client)

	// Build dependencies
	userRepo := repository.NewUserRepository(client)
	userInteractor := usecase.NewUserInteractor(userRepo)

	// Set tenant context
	ctxWithTenant := database.WithTenantID(ctx, dataSet.Tenant1.ID)

	// Try to get non-existent user
	_, err := userInteractor.GetMe(ctxWithTenant, "non-existent-user-id")
	require.Error(t, err)
}

// =============================================================================
// RLS (Row Level Security) Tenant Isolation Tests
// =============================================================================

// TestUserInteractor_RLS_TenantIsolation verifies that RLS properly isolates users by tenant
func TestUserInteractor_RLS_TenantIsolation(t *testing.T) {
	t.Parallel()

	// Setup with RLS enabled
	adminClient, appClient := common.SetupTestClientWithRLS(t)
	ctx := context.Background()

	// Create test data
	dataSet := common.CreateTestDataSet(t, adminClient)

	// Build dependencies
	userRepo := repository.NewUserRepository(appClient)
	userInteractor := usecase.NewUserInteractor(userRepo)

	t.Run("Tenant1 user can get their own info", func(t *testing.T) {
		ctxTenant1 := database.WithTenantID(ctx, dataSet.Tenant1.ID)

		user, err := userInteractor.GetMe(ctxTenant1, dataSet.User1.ID)
		require.NoError(t, err)
		assert.Equal(t, dataSet.User1.ID, user.ID)
		assert.Equal(t, dataSet.User1.Email, user.Email)
	})

	t.Run("Tenant1 cannot access Tenant2 user", func(t *testing.T) {
		ctxTenant1 := database.WithTenantID(ctx, dataSet.Tenant1.ID)

		// Try to get Tenant2's user - should fail (RLS blocks it)
		_, err := userInteractor.GetMe(ctxTenant1, dataSet.User3.ID)
		require.Error(t, err)
	})

	t.Run("Tenant2 cannot access Tenant1 user", func(t *testing.T) {
		ctxTenant2 := database.WithTenantID(ctx, dataSet.Tenant2.ID)

		// Try to get Tenant1's user - should fail (RLS blocks it)
		_, err := userInteractor.GetMe(ctxTenant2, dataSet.User1.ID)
		require.Error(t, err)
	})
}

// TestUserInteractor_RLS_UpdateIsolation verifies that RLS prevents cross-tenant user updates
func TestUserInteractor_RLS_UpdateIsolation(t *testing.T) {
	t.Parallel()

	// Setup with RLS enabled
	adminClient, appClient := common.SetupTestClientWithRLS(t)
	ctx := context.Background()

	// Create test data
	dataSet := common.CreateTestDataSet(t, adminClient)

	// Build dependencies
	userRepo := repository.NewUserRepository(appClient)
	userInteractor := usecase.NewUserInteractor(userRepo)

	t.Run("Tenant2 cannot update Tenant1 user", func(t *testing.T) {
		ctxTenant2 := database.WithTenantID(ctx, dataSet.Tenant2.ID)

		// Try to update Tenant1's user - should fail (RLS blocks finding it)
		newName := "Hacked Name"
		_, err := userInteractor.UpdateMe(ctxTenant2, &input.UpdateUserInput{
			UserID: dataSet.User1.ID,
			Name:   &newName,
		})
		require.Error(t, err)

		// Verify the user wasn't updated
		ctxTenant1 := database.WithTenantID(ctx, dataSet.Tenant1.ID)
		user, err := userInteractor.GetMe(ctxTenant1, dataSet.User1.ID)
		require.NoError(t, err)
		assert.Equal(t, dataSet.User1.Name, user.Name)
		assert.NotEqual(t, newName, user.Name)
	})
}
