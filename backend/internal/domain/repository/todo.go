package repository

import (
	"context"

	"good-todo-go/internal/domain/model"
)

type ITodoRepository interface {
	// Read operations use View with tenant context (tenantID from context)
	FindByID(ctx context.Context, todoID string) (*model.Todo, error)
	FindByUserID(ctx context.Context, userID string, limit, offset int) ([]*model.Todo, error)
	CountByUserID(ctx context.Context, userID string) (int, error)
	// Write operations use direct table access (RLS protected)
	Create(ctx context.Context, todo *model.Todo) (*model.Todo, error)
	Update(ctx context.Context, todo *model.Todo) (*model.Todo, error)
	Delete(ctx context.Context, todoID string) error
}
