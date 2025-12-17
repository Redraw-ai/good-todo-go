package integration_test

import (
	"context"
	"testing"

	"good-todo-go/internal/infrastructure/database"
	"good-todo-go/internal/infrastructure/repository"
	"good-todo-go/internal/integration_test/common"
	"good-todo-go/internal/pkg"
	"good-todo-go/internal/usecase"
	"good-todo-go/internal/usecase/input"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRLS_TenantIsolation_Todos verifies that RLS properly isolates todos by tenant
func TestRLS_TenantIsolation_Todos(t *testing.T) {
	t.Parallel()

	// Setup with RLS enabled
	adminClient, appClient := common.SetupTestClientWithRLS(t)
	ctx := context.Background()

	// Create test data using admin client (bypasses RLS)
	dataSet := common.CreateTestDataSet(t, adminClient)

	// Build dependencies using app client (RLS enforced)
	todoRepo := repository.NewTodoRepository(appClient)
	userRepo := repository.NewUserRepository(appClient)
	uuidGen := pkg.NewUUIDGenerator()
	todoInteractor := usecase.NewTodoInteractor(todoRepo, userRepo, uuidGen)

	t.Run("Tenant1 user can only see Tenant1 todos", func(t *testing.T) {
		// Set context to Tenant1
		ctxTenant1 := database.WithTenantID(ctx, dataSet.Tenant1.ID)

		// Get User1's todos (should get 2 - Todo1 and Todo2)
		result, err := todoInteractor.GetTodos(ctxTenant1, &input.GetTodosInput{
			UserID: dataSet.User1.ID,
			Limit:  10,
			Offset: 0,
		})
		require.NoError(t, err)
		assert.Equal(t, 2, result.Total)

		// Get public todos in Tenant1 (should get 1 - Todo2)
		publicResult, err := todoInteractor.GetPublicTodos(ctxTenant1, &input.GetPublicTodosInput{
			Limit:  10,
			Offset: 0,
		})
		require.NoError(t, err)
		assert.Equal(t, 1, publicResult.Total)
		assert.Equal(t, dataSet.Todo2.ID, publicResult.Todos[0].ID)
	})

	t.Run("Tenant2 user can only see Tenant2 todos", func(t *testing.T) {
		// Set context to Tenant2
		ctxTenant2 := database.WithTenantID(ctx, dataSet.Tenant2.ID)

		// Get User3's todos (should get 2 - Todo4 and Todo5)
		result, err := todoInteractor.GetTodos(ctxTenant2, &input.GetTodosInput{
			UserID: dataSet.User3.ID,
			Limit:  10,
			Offset: 0,
		})
		require.NoError(t, err)
		assert.Equal(t, 2, result.Total)

		// Get public todos in Tenant2 (should get 1 - Todo5)
		publicResult, err := todoInteractor.GetPublicTodos(ctxTenant2, &input.GetPublicTodosInput{
			Limit:  10,
			Offset: 0,
		})
		require.NoError(t, err)
		assert.Equal(t, 1, publicResult.Total)
		assert.Equal(t, dataSet.Todo5.ID, publicResult.Todos[0].ID)
	})

	t.Run("Tenant1 cannot see Tenant2 todos", func(t *testing.T) {
		// Set context to Tenant1
		ctxTenant1 := database.WithTenantID(ctx, dataSet.Tenant1.ID)

		// Try to get Tenant2's todo - should not be found (RLS blocks it)
		_, err := todoInteractor.GetTodo(ctxTenant1, dataSet.Todo4.ID, dataSet.User1.ID)
		require.Error(t, err)
	})

	t.Run("Tenant2 cannot see Tenant1 todos", func(t *testing.T) {
		// Set context to Tenant2
		ctxTenant2 := database.WithTenantID(ctx, dataSet.Tenant2.ID)

		// Try to get Tenant1's todo - should not be found (RLS blocks it)
		_, err := todoInteractor.GetTodo(ctxTenant2, dataSet.Todo1.ID, dataSet.User3.ID)
		require.Error(t, err)
	})
}

// TestRLS_TenantIsolation_CreateTodo verifies that todos are created with correct tenant
func TestRLS_TenantIsolation_CreateTodo(t *testing.T) {
	t.Parallel()

	// Setup with RLS enabled
	adminClient, appClient := common.SetupTestClientWithRLS(t)
	ctx := context.Background()

	// Create test data
	dataSet := common.CreateTestDataSet(t, adminClient)

	// Build dependencies
	todoRepo := repository.NewTodoRepository(appClient)
	userRepo := repository.NewUserRepository(appClient)
	uuidGen := pkg.NewUUIDGenerator()
	todoInteractor := usecase.NewTodoInteractor(todoRepo, userRepo, uuidGen)

	t.Run("Todo created in Tenant1 is not visible from Tenant2", func(t *testing.T) {
		// Create todo in Tenant1
		ctxTenant1 := database.WithTenantID(ctx, dataSet.Tenant1.ID)
		created, err := todoInteractor.CreateTodo(ctxTenant1, &input.CreateTodoInput{
			UserID:      dataSet.User1.ID,
			TenantID:    dataSet.Tenant1.ID,
			Title:       "Tenant1 Only Todo",
			Description: "This should only be visible in Tenant1",
			IsPublic:    true,
		})
		require.NoError(t, err)

		// Verify it's visible in Tenant1
		visible, err := todoInteractor.GetTodo(ctxTenant1, created.ID, dataSet.User1.ID)
		require.NoError(t, err)
		assert.Equal(t, created.ID, visible.ID)

		// Verify it's NOT visible in Tenant2 (RLS blocks it)
		ctxTenant2 := database.WithTenantID(ctx, dataSet.Tenant2.ID)
		_, err = todoInteractor.GetTodo(ctxTenant2, created.ID, dataSet.User3.ID)
		require.Error(t, err)
	})
}

// TestRLS_TenantIsolation_UpdateDeleteTodo verifies that RLS prevents cross-tenant updates/deletes
func TestRLS_TenantIsolation_UpdateDeleteTodo(t *testing.T) {
	t.Parallel()

	// Setup with RLS enabled
	adminClient, appClient := common.SetupTestClientWithRLS(t)
	ctx := context.Background()

	// Create test data
	dataSet := common.CreateTestDataSet(t, adminClient)

	// Build dependencies
	todoRepo := repository.NewTodoRepository(appClient)
	userRepo := repository.NewUserRepository(appClient)
	uuidGen := pkg.NewUUIDGenerator()
	todoInteractor := usecase.NewTodoInteractor(todoRepo, userRepo, uuidGen)

	t.Run("Tenant2 cannot update Tenant1 todo", func(t *testing.T) {
		// Set context to Tenant2
		ctxTenant2 := database.WithTenantID(ctx, dataSet.Tenant2.ID)

		// Try to update Tenant1's todo - should fail (RLS blocks finding it)
		newTitle := "Hacked by Tenant2"
		_, err := todoInteractor.UpdateTodo(ctxTenant2, &input.UpdateTodoInput{
			TodoID: dataSet.Todo1.ID,
			UserID: dataSet.User3.ID,
			Title:  &newTitle,
		})
		require.Error(t, err)
	})

	t.Run("Tenant2 cannot delete Tenant1 todo", func(t *testing.T) {
		// Set context to Tenant2
		ctxTenant2 := database.WithTenantID(ctx, dataSet.Tenant2.ID)

		// Try to delete Tenant1's todo - should fail (RLS blocks finding it)
		err := todoInteractor.DeleteTodo(ctxTenant2, dataSet.Todo1.ID, dataSet.User3.ID)
		require.Error(t, err)

		// Verify the todo still exists in Tenant1
		ctxTenant1 := database.WithTenantID(ctx, dataSet.Tenant1.ID)
		todo, err := todoInteractor.GetTodo(ctxTenant1, dataSet.Todo1.ID, dataSet.User1.ID)
		require.NoError(t, err)
		assert.Equal(t, dataSet.Todo1.ID, todo.ID)
	})
}

// TestRLS_TenantIsolation_Users verifies that RLS properly isolates users by tenant
func TestRLS_TenantIsolation_Users(t *testing.T) {
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
