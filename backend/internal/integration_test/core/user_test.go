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
