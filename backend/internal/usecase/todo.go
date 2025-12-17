//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock_usecase
package usecase

import (
	"context"

	"good-todo-go/internal/domain/model"
	"good-todo-go/internal/domain/repository"
	"good-todo-go/internal/pkg"
	"good-todo-go/internal/pkg/cerror"
	"good-todo-go/internal/usecase/input"
	"good-todo-go/internal/usecase/output"
)

type ITodoInteractor interface {
	GetTodos(ctx context.Context, in *input.GetTodosInput) (*output.TodoListOutput, error)
	GetPublicTodos(ctx context.Context, in *input.GetPublicTodosInput) (*output.TodoListOutput, error)
	GetTodo(ctx context.Context, todoID, userID string) (*output.TodoOutput, error)
	CreateTodo(ctx context.Context, in *input.CreateTodoInput) (*output.TodoOutput, error)
	UpdateTodo(ctx context.Context, in *input.UpdateTodoInput) (*output.TodoOutput, error)
	DeleteTodo(ctx context.Context, todoID, userID string) error
}

type TodoInteractor struct {
	todoRepo repository.ITodoRepository
	userRepo repository.IUserRepository
	uuidGen  pkg.IUUIDGenerator
}

func NewTodoInteractor(
	todoRepo repository.ITodoRepository,
	userRepo repository.IUserRepository,
	uuidGen pkg.IUUIDGenerator,
) ITodoInteractor {
	return &TodoInteractor{
		todoRepo: todoRepo,
		userRepo: userRepo,
		uuidGen:  uuidGen,
	}
}

func (i *TodoInteractor) GetTodos(ctx context.Context, in *input.GetTodosInput) (*output.TodoListOutput, error) {
	limit := in.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	todos, err := i.todoRepo.FindByUserID(ctx, in.UserID, limit, in.Offset)
	if err != nil {
		return nil, cerror.NewInternalServerError("failed to get todos", err)
	}

	total, err := i.todoRepo.CountByUserID(ctx, in.UserID)
	if err != nil {
		return nil, cerror.NewInternalServerError("failed to count todos", err)
	}

	return output.NewTodoListOutput(todos, total), nil
}

func (i *TodoInteractor) GetPublicTodos(ctx context.Context, in *input.GetPublicTodosInput) (*output.TodoListOutput, error) {
	limit := in.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	todos, err := i.todoRepo.FindPublic(ctx, limit, in.Offset)
	if err != nil {
		return nil, cerror.NewInternalServerError("failed to get public todos", err)
	}

	total, err := i.todoRepo.CountPublic(ctx)
	if err != nil {
		return nil, cerror.NewInternalServerError("failed to count public todos", err)
	}

	// Collect unique user IDs and fetch user info
	userIDs := make([]string, 0, len(todos))
	seen := make(map[string]bool)
	for _, t := range todos {
		if !seen[t.UserID] {
			seen[t.UserID] = true
			userIDs = append(userIDs, t.UserID)
		}
	}

	users, err := i.userRepo.FindByIDs(ctx, userIDs)
	if err != nil {
		return nil, cerror.NewInternalServerError("failed to get users", err)
	}

	userMap := make(map[string]*model.User, len(users))
	for _, u := range users {
		userMap[u.ID] = u
	}

	return output.NewTodoListOutputWithCreators(todos, total, userMap), nil
}

func (i *TodoInteractor) GetTodo(ctx context.Context, todoID, userID string) (*output.TodoOutput, error) {
	todo, err := i.todoRepo.FindByID(ctx, todoID)
	if err != nil {
		return nil, cerror.NewNotFound("todo not found", err)
	}

	// Allow access if user owns the todo OR if it's public
	if todo.UserID != userID && !todo.IsPublic {
		return nil, cerror.NewForbidden("not allowed to access this todo", nil)
	}

	return output.NewTodoOutput(todo), nil
}

func (i *TodoInteractor) CreateTodo(ctx context.Context, in *input.CreateTodoInput) (*output.TodoOutput, error) {
	todo := &model.Todo{
		ID:          i.uuidGen.Generate(),
		UserID:      in.UserID,
		TenantID:    in.TenantID,
		Title:       in.Title,
		Description: in.Description,
		Completed:   false,
		IsPublic:    in.IsPublic,
		DueDate:     in.DueDate,
	}

	created, err := i.todoRepo.Create(ctx, todo)
	if err != nil {
		return nil, cerror.NewInternalServerError("failed to create todo", err)
	}

	return output.NewTodoOutput(created), nil
}

func (i *TodoInteractor) UpdateTodo(ctx context.Context, in *input.UpdateTodoInput) (*output.TodoOutput, error) {
	todo, err := i.todoRepo.FindByID(ctx, in.TodoID)
	if err != nil {
		return nil, cerror.NewNotFound("todo not found", err)
	}

	// Only owner can update
	if todo.UserID != in.UserID {
		return nil, cerror.NewForbidden("not allowed to update this todo", nil)
	}

	if in.Title != nil {
		todo.Title = *in.Title
	}
	if in.Description != nil {
		todo.Description = *in.Description
	}
	if in.Completed != nil {
		todo.Completed = *in.Completed
	}
	if in.IsPublic != nil {
		todo.IsPublic = *in.IsPublic
	}
	if in.DueDate != nil {
		todo.DueDate = in.DueDate
	}

	updated, err := i.todoRepo.Update(ctx, todo)
	if err != nil {
		return nil, cerror.NewInternalServerError("failed to update todo", err)
	}

	return output.NewTodoOutput(updated), nil
}

func (i *TodoInteractor) DeleteTodo(ctx context.Context, todoID, userID string) error {
	todo, err := i.todoRepo.FindByID(ctx, todoID)
	if err != nil {
		return cerror.NewNotFound("todo not found", err)
	}

	// Only owner can delete
	if todo.UserID != userID {
		return cerror.NewForbidden("not allowed to delete this todo", nil)
	}

	if err := i.todoRepo.Delete(ctx, todoID); err != nil {
		return cerror.NewInternalServerError("failed to delete todo", err)
	}

	return nil
}
