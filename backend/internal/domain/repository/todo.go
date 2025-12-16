package repository

import (
	"context"

	"good-todo-go/internal/domain/model"
)

type ITodoRepository interface {
	FindByID(ctx context.Context, todoID string) (*model.Todo, error)
	FindByUserID(ctx context.Context, userID string, limit, offset int) ([]*model.Todo, error)
	CountByUserID(ctx context.Context, userID string) (int, error)
	Create(ctx context.Context, todo *model.Todo) (*model.Todo, error)
	Update(ctx context.Context, todo *model.Todo) (*model.Todo, error)
	Delete(ctx context.Context, todoID string) error
}
