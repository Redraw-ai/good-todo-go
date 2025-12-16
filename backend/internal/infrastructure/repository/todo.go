package repository

import (
	"context"

	"good-todo-go/internal/domain/model"
	"good-todo-go/internal/domain/repository"
	"good-todo-go/internal/ent"
	"good-todo-go/internal/ent/tenanttodoview"
	"good-todo-go/internal/infrastructure/database"
)

type TodoRepository struct {
	dbClient *database.DBClient
}

func NewTodoRepository(dbClient *database.DBClient) repository.ITodoRepository {
	return &TodoRepository{dbClient: dbClient}
}

// FindByID reads a single todo via TenantTodoView (tenant-scoped from context)
func (r *TodoRepository) FindByID(ctx context.Context, todoID string) (*model.Todo, error) {
	tx, err := database.TenantScopedTx(ctx, r.dbClient.Ent)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	t, err := tx.TenantTodoView.Query().
		Where(tenanttodoview.IDEQ(todoID)).
		Only(ctx)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return toTodoModelFromView(t), nil
}

// FindByUserID reads todos via TenantTodoView (tenant-scoped from context)
func (r *TodoRepository) FindByUserID(ctx context.Context, userID string, limit, offset int) ([]*model.Todo, error) {
	tx, err := database.TenantScopedTx(ctx, r.dbClient.Ent)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	todos, err := tx.TenantTodoView.Query().
		Where(tenanttodoview.UserIDEQ(userID)).
		Order(ent.Desc(tenanttodoview.FieldCreatedAt)).
		Limit(limit).
		Offset(offset).
		All(ctx)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	result := make([]*model.Todo, len(todos))
	for i, t := range todos {
		result[i] = toTodoModelFromView(t)
	}
	return result, nil
}

// CountByUserID counts todos via TenantTodoView (tenant-scoped from context)
func (r *TodoRepository) CountByUserID(ctx context.Context, userID string) (int, error) {
	tx, err := database.TenantScopedTx(ctx, r.dbClient.Ent)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	count, err := tx.TenantTodoView.Query().
		Where(tenanttodoview.UserIDEQ(userID)).
		Count(ctx)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return count, nil
}

// Create writes directly to todos table (RLS protected)
func (r *TodoRepository) Create(ctx context.Context, t *model.Todo) (*model.Todo, error) {
	builder := r.dbClient.Ent.Todo.Create().
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

// Update writes directly to todos table (RLS protected)
func (r *TodoRepository) Update(ctx context.Context, t *model.Todo) (*model.Todo, error) {
	builder := r.dbClient.Ent.Todo.UpdateOneID(t.ID).
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

// Delete writes directly to todos table (RLS protected)
func (r *TodoRepository) Delete(ctx context.Context, todoID string) error {
	return r.dbClient.Ent.Todo.DeleteOneID(todoID).Exec(ctx)
}

// toTodoModel converts ent.Todo to model.Todo
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

// toTodoModelFromView converts ent.TenantTodoView to model.Todo
func toTodoModelFromView(t *ent.TenantTodoView) *model.Todo {
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
