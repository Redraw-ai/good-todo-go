package repository

import (
	"context"

	"good-todo-go/internal/domain/model"
	"good-todo-go/internal/domain/repository"
	"good-todo-go/internal/ent"
	"good-todo-go/internal/ent/user"
)

type UserRepository struct {
	client *ent.Client
}

func NewUserRepository(client *ent.Client) repository.IUserRepository {
	return &UserRepository{client: client}
}

func (r *UserRepository) FindByID(ctx context.Context, userID string) (*model.User, error) {
	u, err := r.client.User.Get(ctx, userID)
	if err != nil {
		return nil, err
	}
	return toUserModel(u), nil
}

func (r *UserRepository) FindByIDs(ctx context.Context, userIDs []string) ([]*model.User, error) {
	if len(userIDs) == 0 {
		return []*model.User{}, nil
	}
	users, err := r.client.User.Query().
		Where(user.IDIn(userIDs...)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*model.User, len(users))
	for i, u := range users {
		result[i] = toUserModel(u)
	}
	return result, nil
}

func (r *UserRepository) Update(ctx context.Context, user *model.User) (*model.User, error) {
	updated, err := r.client.User.UpdateOneID(user.ID).
		SetName(user.Name).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return toUserModel(updated), nil
}
