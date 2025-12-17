package repository

import (
	"context"
	"testing"
	"time"

	"good-todo-go/internal/domain/model"
	"good-todo-go/internal/integration_test/common"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthRepository_CreateTenant(t *testing.T) {
	t.Parallel()

	client := common.SetupTestClient(t)
	repo := NewAuthRepository(client)

	tests := []struct {
		name    string
		tenant  *model.Tenant
		wantErr bool
	}{
		{
			name: "success - create tenant",
			tenant: &model.Tenant{
				ID:   "test-tenant-id-1",
				Name: "Test Tenant",
				Slug: "test-tenant-1",
			},
			wantErr: false,
		},
		{
			name: "success - create another tenant",
			tenant: &model.Tenant{
				ID:   "test-tenant-id-2",
				Name: "Another Tenant",
				Slug: "another-tenant",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			created, err := repo.CreateTenant(context.Background(), tt.tenant)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.tenant.ID, created.ID)
			assert.Equal(t, tt.tenant.Name, created.Name)
			assert.Equal(t, tt.tenant.Slug, created.Slug)
			assert.NotZero(t, created.CreatedAt)
		})
	}
}

func TestAuthRepository_FindTenantBySlug(t *testing.T) {
	t.Parallel()

	client := common.SetupTestClient(t)
	repo := NewAuthRepository(client)

	// Create test tenant
	tenant := common.CreateTenant(t, client, common.DefaultTenantBuilder(client, "find-tenant-slug"))

	tests := []struct {
		name    string
		slug    string
		wantErr bool
	}{
		{
			name:    "success - tenant found",
			slug:    tenant.Slug,
			wantErr: false,
		},
		{
			name:    "fail - tenant not found",
			slug:    "non-existent-slug",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, err := repo.FindTenantBySlug(context.Background(), tt.slug)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tenant.ID, found.ID)
			assert.Equal(t, tenant.Slug, found.Slug)
		})
	}
}

func TestAuthRepository_CreateUser(t *testing.T) {
	t.Parallel()

	client := common.SetupTestClient(t)
	repo := NewAuthRepository(client)

	tenant := common.CreateTenant(t, client, common.DefaultTenantBuilder(client, ""))

	verificationToken := "test-verification-token"
	tokenExpiry := time.Now().Add(24 * time.Hour)

	tests := []struct {
		name    string
		user    *model.User
		wantErr bool
	}{
		{
			name: "success - create user without verification token",
			user: &model.User{
				ID:            "user-id-1",
				TenantID:      tenant.ID,
				Email:         "test1@example.com",
				PasswordHash:  "hashed-password",
				Name:          "Test User 1",
				Role:          "member",
				EmailVerified: false,
			},
			wantErr: false,
		},
		{
			name: "success - create user with verification token",
			user: &model.User{
				ID:                         "user-id-2",
				TenantID:                   tenant.ID,
				Email:                      "test2@example.com",
				PasswordHash:               "hashed-password",
				Name:                       "Test User 2",
				Role:                       "member",
				EmailVerified:              false,
				VerificationToken:          &verificationToken,
				VerificationTokenExpiresAt: &tokenExpiry,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			created, err := repo.CreateUser(context.Background(), tt.user)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.user.ID, created.ID)
			assert.Equal(t, tt.user.TenantID, created.TenantID)
			assert.Equal(t, tt.user.Email, created.Email)
			assert.Equal(t, tt.user.Name, created.Name)
			assert.NotZero(t, created.CreatedAt)
		})
	}
}

func TestAuthRepository_FindUserByEmail(t *testing.T) {
	t.Parallel()

	client := common.SetupTestClient(t)
	repo := NewAuthRepository(client)

	tenant := common.CreateTenant(t, client, common.DefaultTenantBuilder(client, ""))
	user := common.CreateUser(t, client, common.DefaultUserBuilder(client, "", tenant.ID).SetEmail("findbyemail@example.com"))

	tests := []struct {
		name     string
		tenantID string
		email    string
		wantErr  bool
	}{
		{
			name:     "success - user found",
			tenantID: tenant.ID,
			email:    user.Email,
			wantErr:  false,
		},
		{
			name:     "fail - user not found (wrong email)",
			tenantID: tenant.ID,
			email:    "nonexistent@example.com",
			wantErr:  true,
		},
		{
			name:     "fail - user not found (wrong tenant)",
			tenantID: "wrong-tenant-id",
			email:    user.Email,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, err := repo.FindUserByEmail(context.Background(), tt.tenantID, tt.email)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, user.ID, found.ID)
			assert.Equal(t, user.Email, found.Email)
		})
	}
}

func TestAuthRepository_FindUserByID(t *testing.T) {
	t.Parallel()

	client := common.SetupTestClient(t)
	repo := NewAuthRepository(client)

	tenant := common.CreateTenant(t, client, common.DefaultTenantBuilder(client, ""))
	user := common.CreateUser(t, client, common.DefaultUserBuilder(client, "findbyid-user", tenant.ID))

	tests := []struct {
		name     string
		tenantID string
		userID   string
		wantErr  bool
	}{
		{
			name:     "success - user found",
			tenantID: tenant.ID,
			userID:   user.ID,
			wantErr:  false,
		},
		{
			name:     "fail - user not found (wrong ID)",
			tenantID: tenant.ID,
			userID:   "non-existent-user-id",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, err := repo.FindUserByID(context.Background(), tt.tenantID, tt.userID)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, user.ID, found.ID)
			assert.Equal(t, user.Email, found.Email)
		})
	}
}

func TestAuthRepository_FindUserByVerificationToken(t *testing.T) {
	t.Parallel()

	client := common.SetupTestClient(t)
	repo := NewAuthRepository(client)

	tenant := common.CreateTenant(t, client, common.DefaultTenantBuilder(client, ""))
	token := "unique-verification-token"
	tokenExpiry := time.Now().Add(24 * time.Hour)
	user := common.CreateUser(t, client, common.DefaultUserBuilder(client, "", tenant.ID).
		SetVerificationToken(token).
		SetVerificationTokenExpiresAt(tokenExpiry))

	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "success - user found by token",
			token:   token,
			wantErr: false,
		},
		{
			name:    "fail - user not found (wrong token)",
			token:   "wrong-token",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, err := repo.FindUserByVerificationToken(context.Background(), tt.token)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, user.ID, found.ID)
			assert.NotNil(t, found.VerificationToken)
			assert.Equal(t, token, *found.VerificationToken)
		})
	}
}

func TestAuthRepository_UpdateUser(t *testing.T) {
	t.Parallel()

	client := common.SetupTestClient(t)
	repo := NewAuthRepository(client)

	tenant := common.CreateTenant(t, client, common.DefaultTenantBuilder(client, ""))
	token := "old-token"
	tokenExpiry := time.Now().Add(24 * time.Hour)
	user := common.CreateUser(t, client, common.DefaultUserBuilder(client, "", tenant.ID).
		SetVerificationToken(token).
		SetVerificationTokenExpiresAt(tokenExpiry).
		SetEmailVerified(false))

	tests := []struct {
		name         string
		updateUser   *model.User
		wantErr      bool
		checkResults func(t *testing.T, updated *model.User)
	}{
		{
			name: "success - update name and verify email",
			updateUser: &model.User{
				ID:            user.ID,
				TenantID:      user.TenantID,
				Email:         user.Email,
				PasswordHash:  user.PasswordHash,
				Name:          "Updated Name",
				Role:          string(user.Role),
				EmailVerified: true,
				// Clear verification token
				VerificationToken:          nil,
				VerificationTokenExpiresAt: nil,
			},
			wantErr: false,
			checkResults: func(t *testing.T, updated *model.User) {
				assert.Equal(t, "Updated Name", updated.Name)
				assert.True(t, updated.EmailVerified)
				assert.Nil(t, updated.VerificationToken)
				assert.Nil(t, updated.VerificationTokenExpiresAt)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updated, err := repo.UpdateUser(context.Background(), tt.updateUser)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			if tt.checkResults != nil {
				tt.checkResults(t, updated)
			}
		})
	}
}
