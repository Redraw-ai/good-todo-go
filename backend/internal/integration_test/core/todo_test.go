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

func TestTodoInteractor_CreateAndGet(t *testing.T) {
	t.Parallel()

	// Setup
	client := common.SetupTestClient(t)
	ctx := context.Background()

	// Create test data
	dataSet := common.CreateTestDataSet(t, client)

	// Build dependencies
	todoRepo := repository.NewTodoRepository(client)
	userRepo := repository.NewUserRepository(client)
	uuidGen := pkg.NewUUIDGenerator()
	todoInteractor := usecase.NewTodoInteractor(todoRepo, userRepo, uuidGen)

	// Test create todo
	createInput := &input.CreateTodoInput{
		UserID:      dataSet.User1.ID,
		TenantID:    dataSet.Tenant1.ID,
		Title:       "Integration Test Todo",
		Description: "This is a test todo from integration test",
		IsPublic:    false,
	}

	// Set tenant context
	ctxWithTenant := database.WithTenantID(ctx, dataSet.Tenant1.ID)

	// Create todo
	created, err := todoInteractor.CreateTodo(ctxWithTenant, createInput)
	require.NoError(t, err)
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, createInput.Title, created.Title)
	assert.Equal(t, createInput.Description, created.Description)
	assert.False(t, created.Completed)
	assert.False(t, created.IsPublic)

	// Get todo
	got, err := todoInteractor.GetTodo(ctxWithTenant, created.ID, dataSet.User1.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, created.Title, got.Title)
}

func TestTodoInteractor_GetTodos(t *testing.T) {
	t.Parallel()

	// Setup
	client := common.SetupTestClient(t)
	ctx := context.Background()

	// Create test data
	dataSet := common.CreateTestDataSet(t, client)

	// Build dependencies
	todoRepo := repository.NewTodoRepository(client)
	userRepo := repository.NewUserRepository(client)
	uuidGen := pkg.NewUUIDGenerator()
	todoInteractor := usecase.NewTodoInteractor(todoRepo, userRepo, uuidGen)

	// Set tenant context for Tenant1
	ctxWithTenant := database.WithTenantID(ctx, dataSet.Tenant1.ID)

	// Get todos for User1 (should get Todo1 and Todo2)
	result, err := todoInteractor.GetTodos(ctxWithTenant, &input.GetTodosInput{
		UserID: dataSet.User1.ID,
		Limit:  10,
		Offset: 0,
	})
	require.NoError(t, err)
	assert.Equal(t, 2, result.Total)
	assert.Len(t, result.Todos, 2)
}

func TestTodoInteractor_GetPublicTodos(t *testing.T) {
	t.Parallel()

	// Setup
	client := common.SetupTestClient(t)
	ctx := context.Background()

	// Create test data
	dataSet := common.CreateTestDataSet(t, client)

	// Build dependencies
	todoRepo := repository.NewTodoRepository(client)
	userRepo := repository.NewUserRepository(client)
	uuidGen := pkg.NewUUIDGenerator()
	todoInteractor := usecase.NewTodoInteractor(todoRepo, userRepo, uuidGen)

	// Set tenant context for Tenant1
	ctxWithTenant := database.WithTenantID(ctx, dataSet.Tenant1.ID)

	// Get public todos (should only get Todo2 - public todo in Tenant1)
	result, err := todoInteractor.GetPublicTodos(ctxWithTenant, &input.GetPublicTodosInput{
		Limit:  10,
		Offset: 0,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, result.Total)
	assert.Len(t, result.Todos, 1)
	assert.Equal(t, dataSet.Todo2.ID, result.Todos[0].ID)
}

func TestTodoInteractor_UpdateTodo(t *testing.T) {
	t.Parallel()

	// Setup
	client := common.SetupTestClient(t)
	ctx := context.Background()

	// Create test data
	dataSet := common.CreateTestDataSet(t, client)

	// Build dependencies
	todoRepo := repository.NewTodoRepository(client)
	userRepo := repository.NewUserRepository(client)
	uuidGen := pkg.NewUUIDGenerator()
	todoInteractor := usecase.NewTodoInteractor(todoRepo, userRepo, uuidGen)

	// Set tenant context
	ctxWithTenant := database.WithTenantID(ctx, dataSet.Tenant1.ID)

	// Update todo
	newTitle := "Updated Title"
	completed := true
	updateInput := &input.UpdateTodoInput{
		TodoID:    dataSet.Todo1.ID,
		UserID:    dataSet.User1.ID,
		Title:     &newTitle,
		Completed: &completed,
	}

	updated, err := todoInteractor.UpdateTodo(ctxWithTenant, updateInput)
	require.NoError(t, err)
	assert.Equal(t, newTitle, updated.Title)
	assert.True(t, updated.Completed)
}

func TestTodoInteractor_DeleteTodo(t *testing.T) {
	t.Parallel()

	// Setup
	client := common.SetupTestClient(t)
	ctx := context.Background()

	// Create test data
	dataSet := common.CreateTestDataSet(t, client)

	// Build dependencies
	todoRepo := repository.NewTodoRepository(client)
	userRepo := repository.NewUserRepository(client)
	uuidGen := pkg.NewUUIDGenerator()
	todoInteractor := usecase.NewTodoInteractor(todoRepo, userRepo, uuidGen)

	// Set tenant context
	ctxWithTenant := database.WithTenantID(ctx, dataSet.Tenant1.ID)

	// Delete todo
	err := todoInteractor.DeleteTodo(ctxWithTenant, dataSet.Todo1.ID, dataSet.User1.ID)
	require.NoError(t, err)

	// Try to get deleted todo - should fail
	_, err = todoInteractor.GetTodo(ctxWithTenant, dataSet.Todo1.ID, dataSet.User1.ID)
	assert.Error(t, err)
}

func TestTodoInteractor_CannotAccessOtherUserTodo(t *testing.T) {
	t.Parallel()

	// Setup
	client := common.SetupTestClient(t)
	ctx := context.Background()

	// Create test data
	dataSet := common.CreateTestDataSet(t, client)

	// Build dependencies
	todoRepo := repository.NewTodoRepository(client)
	userRepo := repository.NewUserRepository(client)
	uuidGen := pkg.NewUUIDGenerator()
	todoInteractor := usecase.NewTodoInteractor(todoRepo, userRepo, uuidGen)

	// Set tenant context
	ctxWithTenant := database.WithTenantID(ctx, dataSet.Tenant1.ID)

	// User2 tries to get User1's private todo - should be forbidden
	_, err := todoInteractor.GetTodo(ctxWithTenant, dataSet.Todo1.ID, dataSet.User2.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not allowed")

	// User2 tries to update User1's todo - should be forbidden
	newTitle := "Hacked Title"
	_, err = todoInteractor.UpdateTodo(ctxWithTenant, &input.UpdateTodoInput{
		TodoID: dataSet.Todo1.ID,
		UserID: dataSet.User2.ID,
		Title:  &newTitle,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not allowed")

	// User2 tries to delete User1's todo - should be forbidden
	err = todoInteractor.DeleteTodo(ctxWithTenant, dataSet.Todo1.ID, dataSet.User2.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not allowed")
}

func TestTodoInteractor_CanAccessPublicTodo(t *testing.T) {
	t.Parallel()

	// Setup
	client := common.SetupTestClient(t)
	ctx := context.Background()

	// Create test data
	dataSet := common.CreateTestDataSet(t, client)

	// Build dependencies
	todoRepo := repository.NewTodoRepository(client)
	userRepo := repository.NewUserRepository(client)
	uuidGen := pkg.NewUUIDGenerator()
	todoInteractor := usecase.NewTodoInteractor(todoRepo, userRepo, uuidGen)

	// Set tenant context
	ctxWithTenant := database.WithTenantID(ctx, dataSet.Tenant1.ID)

	// User2 can access User1's public todo
	todo, err := todoInteractor.GetTodo(ctxWithTenant, dataSet.Todo2.ID, dataSet.User2.ID)
	require.NoError(t, err)
	assert.Equal(t, dataSet.Todo2.ID, todo.ID)
	assert.True(t, todo.IsPublic)
}

// =============================================================================
// RLS (Row Level Security) Tenant Isolation Tests
// =============================================================================

// TestTodoInteractor_RLS_TenantIsolation verifies that RLS properly isolates todos by tenant
func TestTodoInteractor_RLS_TenantIsolation(t *testing.T) {
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

// TestTodoInteractor_RLS_CreateTodo verifies that todos are created with correct tenant
func TestTodoInteractor_RLS_CreateTodo(t *testing.T) {
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

// TestTodoInteractor_RLS_UpdateDeleteTodo verifies that RLS prevents cross-tenant updates/deletes
func TestTodoInteractor_RLS_UpdateDeleteTodo(t *testing.T) {
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
