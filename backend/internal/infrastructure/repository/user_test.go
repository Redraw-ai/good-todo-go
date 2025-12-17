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

func TestUserRepository_FindByID(t *testing.T) {
	t.Parallel()

	client := common.SetupTestClient(t)

	tenant := common.CreateTenant(t, client, common.DefaultTenantBuilder(client, ""))
	user := common.CreateUser(t, client, common.DefaultUserBuilder(client, "", tenant.ID))

	repo := NewUserRepository(client)

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
			name:     "fail - user not found",
			tenantID: tenant.ID,
			userID:   "non-existent-user-id",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := database.WithTenantID(context.Background(), tt.tenantID)

			found, err := repo.FindByID(ctx, tt.userID)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, user.ID, found.ID)
			assert.Equal(t, user.Email, found.Email)
			assert.Equal(t, user.Name, found.Name)
		})
	}
}

func TestUserRepository_FindByIDs(t *testing.T) {
	t.Parallel()

	client := common.SetupTestClient(t)

	tenant := common.CreateTenant(t, client, common.DefaultTenantBuilder(client, ""))
	user1 := common.CreateUser(t, client, common.DefaultUserBuilder(client, "", tenant.ID).SetEmail("user1@example.com"))
	user2 := common.CreateUser(t, client, common.DefaultUserBuilder(client, "", tenant.ID).SetEmail("user2@example.com"))
	user3 := common.CreateUser(t, client, common.DefaultUserBuilder(client, "", tenant.ID).SetEmail("user3@example.com"))

	repo := NewUserRepository(client)
	ctx := database.WithTenantID(context.Background(), tenant.ID)

	tests := []struct {
		name          string
		userIDs       []string
		expectedCount int
		wantErr       bool
	}{
		{
			name:          "success - find multiple users",
			userIDs:       []string{user1.ID, user2.ID},
			expectedCount: 2,
			wantErr:       false,
		},
		{
			name:          "success - find all users",
			userIDs:       []string{user1.ID, user2.ID, user3.ID},
			expectedCount: 3,
			wantErr:       false,
		},
		{
			name:          "success - find single user",
			userIDs:       []string{user1.ID},
			expectedCount: 1,
			wantErr:       false,
		},
		{
			name:          "success - empty list",
			userIDs:       []string{},
			expectedCount: 0,
			wantErr:       false,
		},
		{
			name:          "success - some users not found",
			userIDs:       []string{user1.ID, "non-existent-id"},
			expectedCount: 1,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, err := repo.FindByIDs(ctx, tt.userIDs)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, found, tt.expectedCount)
		})
	}
}

func TestUserRepository_Update(t *testing.T) {
	t.Parallel()

	client := common.SetupTestClient(t)

	tenant := common.CreateTenant(t, client, common.DefaultTenantBuilder(client, ""))
	user := common.CreateUser(t, client, common.DefaultUserBuilder(client, "", tenant.ID).SetName("Original Name"))

	repo := NewUserRepository(client)
	ctx := database.WithTenantID(context.Background(), tenant.ID)

	tests := []struct {
		name         string
		updateUser   *model.User
		expectedName string
		wantErr      bool
	}{
		{
			name: "success - update name",
			updateUser: &model.User{
				ID:       user.ID,
				TenantID: tenant.ID,
				Email:    user.Email,
				Name:     "Updated Name",
			},
			expectedName: "Updated Name",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updated, err := repo.Update(ctx, tt.updateUser)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedName, updated.Name)
			assert.Equal(t, user.ID, updated.ID)

			// Verify the update persisted
			found, err := repo.FindByID(ctx, user.ID)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedName, found.Name)
		})
	}
}

// =============================================================================
// RLS (Row Level Security) Tenant Isolation Tests
// =============================================================================

func TestUserRepository_RLS_TenantIsolation(t *testing.T) {
	t.Parallel()

	adminClient, appClient := common.SetupTestClientWithRLS(t)

	// Create test data using admin client (bypasses RLS)
	tenant1 := common.CreateTenant(t, adminClient, common.DefaultTenantBuilder(adminClient, "tenant1-rls"))
	tenant2 := common.CreateTenant(t, adminClient, common.DefaultTenantBuilder(adminClient, "tenant2-rls"))

	user1 := common.CreateUser(t, adminClient, common.DefaultUserBuilder(adminClient, "", tenant1.ID).SetEmail("user1-rls@example.com"))
	user2 := common.CreateUser(t, adminClient, common.DefaultUserBuilder(adminClient, "", tenant2.ID).SetEmail("user2-rls@example.com"))

	// Use app client (RLS enforced)
	repo := NewUserRepository(appClient)

	tests := []struct {
		name        string
		tenantID    string
		userID      string
		wantErr     bool
		description string
	}{
		{
			name:        "success - Tenant1 can access Tenant1 user",
			tenantID:    tenant1.ID,
			userID:      user1.ID,
			wantErr:     false,
			description: "User should be accessible within same tenant",
		},
		{
			name:        "fail - Tenant1 cannot access Tenant2 user",
			tenantID:    tenant1.ID,
			userID:      user2.ID,
			wantErr:     true,
			description: "RLS should block cross-tenant access",
		},
		{
			name:        "success - Tenant2 can access Tenant2 user",
			tenantID:    tenant2.ID,
			userID:      user2.ID,
			wantErr:     false,
			description: "User should be accessible within same tenant",
		},
		{
			name:        "fail - Tenant2 cannot access Tenant1 user",
			tenantID:    tenant2.ID,
			userID:      user1.ID,
			wantErr:     true,
			description: "RLS should block cross-tenant access",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := database.WithTenantID(context.Background(), tt.tenantID)

			found, err := repo.FindByID(ctx, tt.userID)

			if tt.wantErr {
				require.Error(t, err, tt.description)
				return
			}

			require.NoError(t, err, tt.description)
			assert.NotNil(t, found)
		})
	}
}

func TestUserRepository_RLS_UpdateIsolation(t *testing.T) {
	t.Parallel()

	adminClient, appClient := common.SetupTestClientWithRLS(t)

	// Create test data
	tenant1 := common.CreateTenant(t, adminClient, common.DefaultTenantBuilder(adminClient, "tenant1-update-rls"))
	tenant2 := common.CreateTenant(t, adminClient, common.DefaultTenantBuilder(adminClient, "tenant2-update-rls"))

	user1 := common.CreateUser(t, adminClient, common.DefaultUserBuilder(adminClient, "", tenant1.ID).SetName("Original Name"))

	repo := NewUserRepository(appClient)

	t.Run("fail - Tenant2 cannot update Tenant1 user", func(t *testing.T) {
		ctx := database.WithTenantID(context.Background(), tenant2.ID)

		_, err := repo.Update(ctx, &model.User{
			ID:       user1.ID,
			TenantID: tenant1.ID,
			Name:     "Hacked Name",
		})

		require.Error(t, err, "RLS should block cross-tenant update")

		// Verify the user wasn't updated using admin client
		adminRepo := NewUserRepository(adminClient)
		adminCtx := database.WithTenantID(context.Background(), tenant1.ID)
		found, err := adminRepo.FindByID(adminCtx, user1.ID)
		require.NoError(t, err)
		assert.Equal(t, "Original Name", found.Name)
	})
}
