package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"good-todo-go/internal/domain/model"
	mock_repository "good-todo-go/internal/domain/repository/mock"
	"good-todo-go/internal/pkg"
	mock_pkg "good-todo-go/internal/pkg/mock"
	"good-todo-go/internal/usecase/input"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestAuthInteractor_Register(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       *input.RegisterInput
		setupMocks  func(authRepo *mock_repository.MockIAuthRepository, uuidGen *mock_pkg.MockIUUIDGenerator)
		wantErr     bool
		errContains string
	}{
		{
			name: "success - new tenant and user",
			input: &input.RegisterInput{
				Email:      "test@example.com",
				Password:   "password123",
				Name:       "Test User",
				TenantSlug: "test-tenant",
			},
			setupMocks: func(authRepo *mock_repository.MockIAuthRepository, uuidGen *mock_pkg.MockIUUIDGenerator) {
				// Tenant not found, create new
				authRepo.EXPECT().
					FindTenantBySlug(gomock.Any(), "test-tenant").
					Return(nil, errors.New("not found"))

				uuidGen.EXPECT().Generate().Return("tenant-uuid-1")

				authRepo.EXPECT().
					CreateTenant(gomock.Any(), gomock.Any()).
					Return(&model.Tenant{
						ID:   "tenant-uuid-1",
						Name: "test-tenant",
						Slug: "test-tenant",
					}, nil)

				// User not found (new user)
				authRepo.EXPECT().
					FindUserByEmail(gomock.Any(), "tenant-uuid-1", "test@example.com").
					Return(nil, errors.New("not found"))

				uuidGen.EXPECT().Generate().Return("user-uuid-1")

				authRepo.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Return(&model.User{
						ID:            "user-uuid-1",
						TenantID:      "tenant-uuid-1",
						Email:         "test@example.com",
						Name:          "Test User",
						Role:          "member",
						EmailVerified: false,
					}, nil)
			},
			wantErr: false,
		},
		{
			name: "success - existing tenant",
			input: &input.RegisterInput{
				Email:      "test2@example.com",
				Password:   "password123",
				Name:       "Test User 2",
				TenantSlug: "existing-tenant",
			},
			setupMocks: func(authRepo *mock_repository.MockIAuthRepository, uuidGen *mock_pkg.MockIUUIDGenerator) {
				// Tenant found
				authRepo.EXPECT().
					FindTenantBySlug(gomock.Any(), "existing-tenant").
					Return(&model.Tenant{
						ID:   "existing-tenant-id",
						Name: "Existing Tenant",
						Slug: "existing-tenant",
					}, nil)

				// User not found (new user)
				authRepo.EXPECT().
					FindUserByEmail(gomock.Any(), "existing-tenant-id", "test2@example.com").
					Return(nil, errors.New("not found"))

				uuidGen.EXPECT().Generate().Return("user-uuid-2")

				authRepo.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Return(&model.User{
						ID:            "user-uuid-2",
						TenantID:      "existing-tenant-id",
						Email:         "test2@example.com",
						Name:          "Test User 2",
						Role:          "member",
						EmailVerified: false,
					}, nil)
			},
			wantErr: false,
		},
		{
			name: "fail - email already exists",
			input: &input.RegisterInput{
				Email:      "existing@example.com",
				Password:   "password123",
				Name:       "Existing User",
				TenantSlug: "test-tenant",
			},
			setupMocks: func(authRepo *mock_repository.MockIAuthRepository, uuidGen *mock_pkg.MockIUUIDGenerator) {
				authRepo.EXPECT().
					FindTenantBySlug(gomock.Any(), "test-tenant").
					Return(&model.Tenant{
						ID:   "tenant-id",
						Name: "Test Tenant",
						Slug: "test-tenant",
					}, nil)

				// User already exists
				authRepo.EXPECT().
					FindUserByEmail(gomock.Any(), "tenant-id", "existing@example.com").
					Return(&model.User{
						ID:    "existing-user-id",
						Email: "existing@example.com",
					}, nil)
			},
			wantErr:     true,
			errContains: "already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			authRepo := mock_repository.NewMockIAuthRepository(ctrl)
			uuidGen := mock_pkg.NewMockIUUIDGenerator(ctrl)
			jwtService := pkg.NewJWTService("test-secret", 3600, 86400)

			tt.setupMocks(authRepo, uuidGen)

			interactor := NewAuthInteractor(authRepo, jwtService, uuidGen)

			result, err := interactor.Register(context.Background(), tt.input)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, result.AccessToken)
			assert.NotEmpty(t, result.RefreshToken)
			assert.Equal(t, "Bearer", result.TokenType)
			assert.NotNil(t, result.User)
		})
	}
}

func TestAuthInteractor_Login(t *testing.T) {
	t.Parallel()

	// Create a valid password hash
	passwordHash, _ := pkg.HashPassword("password123")

	tests := []struct {
		name        string
		input       *input.LoginInput
		setupMocks  func(authRepo *mock_repository.MockIAuthRepository)
		wantErr     bool
		errContains string
	}{
		{
			name: "success - valid credentials",
			input: &input.LoginInput{
				Email:      "test@example.com",
				Password:   "password123",
				TenantSlug: "test-tenant",
			},
			setupMocks: func(authRepo *mock_repository.MockIAuthRepository) {
				authRepo.EXPECT().
					FindTenantBySlug(gomock.Any(), "test-tenant").
					Return(&model.Tenant{
						ID:   "tenant-id",
						Slug: "test-tenant",
					}, nil)

				authRepo.EXPECT().
					FindUserByEmail(gomock.Any(), "tenant-id", "test@example.com").
					Return(&model.User{
						ID:           "user-id",
						TenantID:     "tenant-id",
						Email:        "test@example.com",
						PasswordHash: passwordHash,
						Name:         "Test User",
						Role:         "member",
					}, nil)
			},
			wantErr: false,
		},
		{
			name: "fail - tenant not found",
			input: &input.LoginInput{
				Email:      "test@example.com",
				Password:   "password123",
				TenantSlug: "non-existent",
			},
			setupMocks: func(authRepo *mock_repository.MockIAuthRepository) {
				authRepo.EXPECT().
					FindTenantBySlug(gomock.Any(), "non-existent").
					Return(nil, errors.New("not found"))
			},
			wantErr:     true,
			errContains: "invalid credentials",
		},
		{
			name: "fail - user not found",
			input: &input.LoginInput{
				Email:      "nonexistent@example.com",
				Password:   "password123",
				TenantSlug: "test-tenant",
			},
			setupMocks: func(authRepo *mock_repository.MockIAuthRepository) {
				authRepo.EXPECT().
					FindTenantBySlug(gomock.Any(), "test-tenant").
					Return(&model.Tenant{
						ID:   "tenant-id",
						Slug: "test-tenant",
					}, nil)

				authRepo.EXPECT().
					FindUserByEmail(gomock.Any(), "tenant-id", "nonexistent@example.com").
					Return(nil, errors.New("not found"))
			},
			wantErr:     true,
			errContains: "invalid credentials",
		},
		{
			name: "fail - wrong password",
			input: &input.LoginInput{
				Email:      "test@example.com",
				Password:   "wrongpassword",
				TenantSlug: "test-tenant",
			},
			setupMocks: func(authRepo *mock_repository.MockIAuthRepository) {
				authRepo.EXPECT().
					FindTenantBySlug(gomock.Any(), "test-tenant").
					Return(&model.Tenant{
						ID:   "tenant-id",
						Slug: "test-tenant",
					}, nil)

				authRepo.EXPECT().
					FindUserByEmail(gomock.Any(), "tenant-id", "test@example.com").
					Return(&model.User{
						ID:           "user-id",
						TenantID:     "tenant-id",
						Email:        "test@example.com",
						PasswordHash: passwordHash,
					}, nil)
			},
			wantErr:     true,
			errContains: "invalid credentials",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			authRepo := mock_repository.NewMockIAuthRepository(ctrl)
			uuidGen := mock_pkg.NewMockIUUIDGenerator(ctrl)
			jwtService := pkg.NewJWTService("test-secret", 3600, 86400)

			tt.setupMocks(authRepo)

			interactor := NewAuthInteractor(authRepo, jwtService, uuidGen)

			result, err := interactor.Login(context.Background(), tt.input)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, result.AccessToken)
			assert.NotEmpty(t, result.RefreshToken)
			assert.Equal(t, "Bearer", result.TokenType)
		})
	}
}

func TestAuthInteractor_VerifyEmail(t *testing.T) {
	t.Parallel()

	validExpiry := time.Now().Add(24 * time.Hour)
	expiredExpiry := time.Now().Add(-1 * time.Hour)

	tests := []struct {
		name        string
		input       *input.VerifyEmailInput
		setupMocks  func(authRepo *mock_repository.MockIAuthRepository)
		wantErr     bool
		errContains string
	}{
		{
			name: "success - valid token",
			input: &input.VerifyEmailInput{
				Token: "valid-token",
			},
			setupMocks: func(authRepo *mock_repository.MockIAuthRepository) {
				token := "valid-token"
				authRepo.EXPECT().
					FindUserByVerificationToken(gomock.Any(), "valid-token").
					Return(&model.User{
						ID:                         "user-id",
						TenantID:                   "tenant-id",
						Email:                      "test@example.com",
						EmailVerified:              false,
						VerificationToken:          &token,
						VerificationTokenExpiresAt: &validExpiry,
					}, nil)

				authRepo.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Return(&model.User{
						ID:            "user-id",
						TenantID:      "tenant-id",
						Email:         "test@example.com",
						EmailVerified: true,
					}, nil)
			},
			wantErr: false,
		},
		{
			name: "fail - invalid token",
			input: &input.VerifyEmailInput{
				Token: "invalid-token",
			},
			setupMocks: func(authRepo *mock_repository.MockIAuthRepository) {
				authRepo.EXPECT().
					FindUserByVerificationToken(gomock.Any(), "invalid-token").
					Return(nil, errors.New("not found"))
			},
			wantErr:     true,
			errContains: "invalid or expired token",
		},
		{
			name: "fail - expired token",
			input: &input.VerifyEmailInput{
				Token: "expired-token",
			},
			setupMocks: func(authRepo *mock_repository.MockIAuthRepository) {
				token := "expired-token"
				authRepo.EXPECT().
					FindUserByVerificationToken(gomock.Any(), "expired-token").
					Return(&model.User{
						ID:                         "user-id",
						TenantID:                   "tenant-id",
						Email:                      "test@example.com",
						EmailVerified:              false,
						VerificationToken:          &token,
						VerificationTokenExpiresAt: &expiredExpiry,
					}, nil)
			},
			wantErr:     true,
			errContains: "expired",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			authRepo := mock_repository.NewMockIAuthRepository(ctrl)
			uuidGen := mock_pkg.NewMockIUUIDGenerator(ctrl)
			jwtService := pkg.NewJWTService("test-secret", 3600, 86400)

			tt.setupMocks(authRepo)

			interactor := NewAuthInteractor(authRepo, jwtService, uuidGen)

			result, err := interactor.VerifyEmail(context.Background(), tt.input)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			assert.Contains(t, result.Message, "verified")
		})
	}
}

func TestAuthInteractor_RefreshToken(t *testing.T) {
	t.Parallel()

	jwtService := pkg.NewJWTService("test-secret", 3600, 86400)

	// Generate a valid refresh token
	tokenPair, _ := jwtService.GenerateTokenPair("user-id", "tenant-id", "test@example.com", "member")

	tests := []struct {
		name        string
		input       *input.RefreshTokenInput
		setupMocks  func(authRepo *mock_repository.MockIAuthRepository)
		wantErr     bool
		errContains string
	}{
		{
			name: "success - valid refresh token",
			input: &input.RefreshTokenInput{
				RefreshToken: tokenPair.RefreshToken,
			},
			setupMocks: func(authRepo *mock_repository.MockIAuthRepository) {
				authRepo.EXPECT().
					FindUserByID(gomock.Any(), "tenant-id", "user-id").
					Return(&model.User{
						ID:       "user-id",
						TenantID: "tenant-id",
						Email:    "test@example.com",
						Role:     "member",
					}, nil)
			},
			wantErr: false,
		},
		{
			name: "fail - invalid refresh token",
			input: &input.RefreshTokenInput{
				RefreshToken: "invalid-token",
			},
			setupMocks:  func(authRepo *mock_repository.MockIAuthRepository) {},
			wantErr:     true,
			errContains: "invalid refresh token",
		},
		{
			name: "fail - user not found",
			input: &input.RefreshTokenInput{
				RefreshToken: tokenPair.RefreshToken,
			},
			setupMocks: func(authRepo *mock_repository.MockIAuthRepository) {
				authRepo.EXPECT().
					FindUserByID(gomock.Any(), "tenant-id", "user-id").
					Return(nil, errors.New("not found"))
			},
			wantErr:     true,
			errContains: "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			authRepo := mock_repository.NewMockIAuthRepository(ctrl)
			uuidGen := mock_pkg.NewMockIUUIDGenerator(ctrl)

			tt.setupMocks(authRepo)

			interactor := NewAuthInteractor(authRepo, jwtService, uuidGen)

			result, err := interactor.RefreshToken(context.Background(), tt.input)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, result.AccessToken)
			assert.NotEmpty(t, result.RefreshToken)
		})
	}
}
