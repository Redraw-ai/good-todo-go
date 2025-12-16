package usecase

import (
	"context"

	"good-todo-go/internal/domain/repository"
	"good-todo-go/internal/pkg/cerror"
	"good-todo-go/internal/usecase/input"
	"good-todo-go/internal/usecase/output"
)

type IUserInteractor interface {
	GetMe(ctx context.Context, userID string) (*output.UserOutput, error)
	UpdateMe(ctx context.Context, in *input.UpdateUserInput) (*output.UserOutput, error)
}

type UserInteractor struct {
	userRepo repository.IUserRepository
}

func NewUserInteractor(userRepo repository.IUserRepository) IUserInteractor {
	return &UserInteractor{
		userRepo: userRepo,
	}
}

func (i *UserInteractor) GetMe(ctx context.Context, userID string) (*output.UserOutput, error) {
	user, err := i.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, cerror.NewNotFound("user not found", err)
	}
	return output.NewUserOutput(user), nil
}

func (i *UserInteractor) UpdateMe(ctx context.Context, in *input.UpdateUserInput) (*output.UserOutput, error) {
	user, err := i.userRepo.FindByID(ctx, in.UserID)
	if err != nil {
		return nil, cerror.NewNotFound("user not found", err)
	}

	if in.Name != nil {
		user.Name = *in.Name
	}

	updated, err := i.userRepo.Update(ctx, user)
	if err != nil {
		return nil, cerror.NewInternalServerError("failed to update user", err)
	}

	return output.NewUserOutput(updated), nil
}
