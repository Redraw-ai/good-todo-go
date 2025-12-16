package repository

import (
	"context"

	"good-todo-go/internal/domain/model"
)

type IAuthRepository interface {
	// Tenant operations
	FindTenantBySlug(ctx context.Context, slug string) (*model.Tenant, error)
	CreateTenant(ctx context.Context, tenant *model.Tenant) (*model.Tenant, error)

	// User operations
	FindUserByEmail(ctx context.Context, tenantID, email string) (*model.User, error)
	FindUserByID(ctx context.Context, userID string) (*model.User, error)
	FindUserByVerificationToken(ctx context.Context, token string) (*model.User, error)
	CreateUser(ctx context.Context, user *model.User) (*model.User, error)
	UpdateUser(ctx context.Context, user *model.User) (*model.User, error)
}
