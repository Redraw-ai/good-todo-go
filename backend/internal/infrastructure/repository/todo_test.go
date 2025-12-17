package repository

import (
	"context"
	"testing"

	"good-todo-go/internal/domain/model"
	"good-todo-go/internal/infrastructure/database"
	"good-todo-go/internal/integration_test/common"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTodoRepository_Create(t *testing.T) {
	t.Parallel()

	client := common.SetupTestClient(t)

	// Create test tenant and user
	tenant := common.CreateTenant(t, client, common.DefaultTenantBuilder(client, ""))
	user := common.CreateUser(t, client, common.DefaultUserBuilder(client, "", tenant.ID))

	repo := NewTodoRepository(client)

	// Set tenant context
	ctx := database.WithTenantID(context.Background(), tenant.ID)

	todo := &model.Todo{
		ID:          "todo-create-1",
		TenantID:    tenant.ID,
		UserID:      user.ID,
		Title:       "Test Todo",
		Description: "Test Description",
		Completed:   false,
		IsPublic:    false,
	}

	created, err := repo.Create(ctx, todo)
	require.NoError(t, err)
	assert.Equal(t, todo.ID, created.ID)
	assert.Equal(t, todo.Title, created.Title)
	assert.Equal(t, todo.TenantID, created.TenantID)
	assert.Equal(t, todo.UserID, created.UserID)
	assert.False(t, created.Completed)
	assert.False(t, created.IsPublic)
}

func TestTodoRepository_FindByID(t *testing.T) {
	t.Parallel()

	client := common.SetupTestClient(t)

	// Create test data
	tenant := common.CreateTenant(t, client, common.DefaultTenantBuilder(client, ""))
	user := common.CreateUser(t, client, common.DefaultUserBuilder(client, "", tenant.ID))
	todo := common.CreateTodo(t, client, common.DefaultTodoBuilder(client, "todo-find-1", tenant.ID, user.ID))

	repo := NewTodoRepository(client)

	// Set tenant context
	ctx := database.WithTenantID(context.Background(), tenant.ID)

	found, err := repo.FindByID(ctx, todo.ID)
	require.NoError(t, err)
	assert.Equal(t, todo.ID, found.ID)
	assert.Equal(t, todo.Title, found.Title)
}

func TestTodoRepository_FindByUserID(t *testing.T) {
	t.Parallel()

	client := common.SetupTestClient(t)

	// Create test data
	tenant := common.CreateTenant(t, client, common.DefaultTenantBuilder(client, ""))
	user := common.CreateUser(t, client, common.DefaultUserBuilder(client, "", tenant.ID))

	// Create multiple todos for the user
	common.CreateTodo(t, client, common.DefaultTodoBuilder(client, "", tenant.ID, user.ID).SetTitle("Todo 1"))
	common.CreateTodo(t, client, common.DefaultTodoBuilder(client, "", tenant.ID, user.ID).SetTitle("Todo 2"))
	common.CreateTodo(t, client, common.DefaultTodoBuilder(client, "", tenant.ID, user.ID).SetTitle("Todo 3"))

	repo := NewTodoRepository(client)

	// Set tenant context
	ctx := database.WithTenantID(context.Background(), tenant.ID)

	todos, err := repo.FindByUserID(ctx, user.ID, 10, 0)
	require.NoError(t, err)
	assert.Len(t, todos, 3)
}

func TestTodoRepository_CountByUserID(t *testing.T) {
	t.Parallel()

	client := common.SetupTestClient(t)

	// Create test data
	tenant := common.CreateTenant(t, client, common.DefaultTenantBuilder(client, ""))
	user := common.CreateUser(t, client, common.DefaultUserBuilder(client, "", tenant.ID))

	// Create multiple todos for the user
	common.CreateTodo(t, client, common.DefaultTodoBuilder(client, "", tenant.ID, user.ID))
	common.CreateTodo(t, client, common.DefaultTodoBuilder(client, "", tenant.ID, user.ID))

	repo := NewTodoRepository(client)

	// Set tenant context
	ctx := database.WithTenantID(context.Background(), tenant.ID)

	count, err := repo.CountByUserID(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestTodoRepository_FindPublic(t *testing.T) {
	t.Parallel()

	client := common.SetupTestClient(t)

	// Create test data
	tenant := common.CreateTenant(t, client, common.DefaultTenantBuilder(client, ""))
	user := common.CreateUser(t, client, common.DefaultUserBuilder(client, "", tenant.ID))

	// Create public and private todos
	common.CreateTodo(t, client, common.DefaultTodoBuilder(client, "", tenant.ID, user.ID).SetIsPublic(true).SetTitle("Public 1"))
	common.CreateTodo(t, client, common.DefaultTodoBuilder(client, "", tenant.ID, user.ID).SetIsPublic(true).SetTitle("Public 2"))
	common.CreateTodo(t, client, common.DefaultTodoBuilder(client, "", tenant.ID, user.ID).SetIsPublic(false).SetTitle("Private"))

	repo := NewTodoRepository(client)

	// Set tenant context
	ctx := database.WithTenantID(context.Background(), tenant.ID)

	todos, err := repo.FindPublic(ctx, 10, 0)
	require.NoError(t, err)
	assert.Len(t, todos, 2)

	for _, todo := range todos {
		assert.True(t, todo.IsPublic)
	}
}

func TestTodoRepository_Update(t *testing.T) {
	t.Parallel()

	client := common.SetupTestClient(t)

	// Create test data
	tenant := common.CreateTenant(t, client, common.DefaultTenantBuilder(client, ""))
	user := common.CreateUser(t, client, common.DefaultUserBuilder(client, "", tenant.ID))
	todo := common.CreateTodo(t, client, common.DefaultTodoBuilder(client, "", tenant.ID, user.ID).SetTitle("Original Title"))

	repo := NewTodoRepository(client)

	// Set tenant context
	ctx := database.WithTenantID(context.Background(), tenant.ID)

	// Update the todo
	todoModel := &model.Todo{
		ID:          todo.ID,
		TenantID:    tenant.ID,
		UserID:      user.ID,
		Title:       "Updated Title",
		Description: "Updated Description",
		Completed:   true,
		IsPublic:    true,
	}

	updated, err := repo.Update(ctx, todoModel)
	require.NoError(t, err)
	assert.Equal(t, "Updated Title", updated.Title)
	assert.Equal(t, "Updated Description", updated.Description)
	assert.True(t, updated.Completed)
	assert.True(t, updated.IsPublic)
}

func TestTodoRepository_Delete(t *testing.T) {
	t.Parallel()

	client := common.SetupTestClient(t)

	// Create test data
	tenant := common.CreateTenant(t, client, common.DefaultTenantBuilder(client, ""))
	user := common.CreateUser(t, client, common.DefaultUserBuilder(client, "", tenant.ID))
	todo := common.CreateTodo(t, client, common.DefaultTodoBuilder(client, "", tenant.ID, user.ID))

	repo := NewTodoRepository(client)

	// Set tenant context
	ctx := database.WithTenantID(context.Background(), tenant.ID)

	// Delete the todo
	err := repo.Delete(ctx, todo.ID)
	require.NoError(t, err)

	// Verify deletion
	_, err = repo.FindByID(ctx, todo.ID)
	assert.Error(t, err) // Should not find deleted todo
}
