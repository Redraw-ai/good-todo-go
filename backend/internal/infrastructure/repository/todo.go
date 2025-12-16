package repository

import (
	"context"

	"good-todo-go/internal/domain/model"
	"good-todo-go/internal/domain/repository"
	"good-todo-go/internal/ent"
	"good-todo-go/internal/ent/todo"
)

type TodoRepository struct {
	client *ent.Client
}

func NewTodoRepository(client *ent.Client) repository.ITodoRepository {
	return &TodoRepository{client: client}
}

func (r *TodoRepository) FindByID(ctx context.Context, todoID string) (*model.Todo, error) {
	t, err := r.client.Todo.Get(ctx, todoID)
	if err != nil {
		return nil, err
	}
	return toTodoModel(t), nil
}

func (r *TodoRepository) FindByUserID(ctx context.Context, userID string, limit, offset int) ([]*model.Todo, error) {
	todos, err := r.client.Todo.Query().
		Where(todo.UserIDEQ(userID)).
		Order(ent.Desc(todo.FieldCreatedAt)).
		Limit(limit).
		Offset(offset).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*model.Todo, len(todos))
	for i, t := range todos {
		result[i] = toTodoModel(t)
	}
	return result, nil
}

func (r *TodoRepository) CountByUserID(ctx context.Context, userID string) (int, error) {
	return r.client.Todo.Query().
		Where(todo.UserIDEQ(userID)).
		Count(ctx)
}

func (r *TodoRepository) Create(ctx context.Context, t *model.Todo) (*model.Todo, error) {
	builder := r.client.Todo.Create().
		SetID(t.ID).
		SetTenantID(t.TenantID).
		SetUserID(t.UserID).
		SetTitle(t.Title).
		SetDescription(t.Description).
		SetCompleted(t.Completed)

	if t.DueDate != nil {
		builder.SetDueDate(*t.DueDate)
	}

	created, err := builder.Save(ctx)
	if err != nil {
		return nil, err
	}
	return toTodoModel(created), nil
}

func (r *TodoRepository) Update(ctx context.Context, t *model.Todo) (*model.Todo, error) {
	builder := r.client.Todo.UpdateOneID(t.ID).
		SetTitle(t.Title).
		SetDescription(t.Description).
		SetCompleted(t.Completed)

	if t.DueDate != nil {
		builder.SetDueDate(*t.DueDate)
	} else {
		builder.ClearDueDate()
	}

	updated, err := builder.Save(ctx)
	if err != nil {
		return nil, err
	}
	return toTodoModel(updated), nil
}

func (r *TodoRepository) Delete(ctx context.Context, todoID string) error {
	return r.client.Todo.DeleteOneID(todoID).Exec(ctx)
}

func toTodoModel(t *ent.Todo) *model.Todo {
	return &model.Todo{
		ID:          t.ID,
		UserID:      t.UserID,
		TenantID:    t.TenantID,
		Title:       t.Title,
		Description: t.Description,
		Completed:   t.Completed,
		DueDate:     t.DueDate,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}
