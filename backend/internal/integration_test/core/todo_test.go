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
