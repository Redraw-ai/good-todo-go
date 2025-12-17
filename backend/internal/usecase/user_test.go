package usecase

import (
	"context"
	"errors"
	"testing"

	"good-todo-go/internal/domain/model"
	mock_repository "good-todo-go/internal/domain/repository/mock"
	"good-todo-go/internal/usecase/input"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestUserInteractor_GetMe(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		userID      string
		setupMocks  func(userRepo *mock_repository.MockIUserRepository)
		wantErr     bool
		errContains string
	}{
		{
			name:   "success - user found",
			userID: "user-id-1",
			setupMocks: func(userRepo *mock_repository.MockIUserRepository) {
				userRepo.EXPECT().
					FindByID(gomock.Any(), "user-id-1").
					Return(&model.User{
						ID:       "user-id-1",
						TenantID: "tenant-id",
						Email:    "test@example.com",
						Name:     "Test User",
						Role:     "member",
					}, nil)
			},
			wantErr: false,
		},
		{
			name:   "fail - user not found",
			userID: "non-existent-id",
			setupMocks: func(userRepo *mock_repository.MockIUserRepository) {
				userRepo.EXPECT().
					FindByID(gomock.Any(), "non-existent-id").
					Return(nil, errors.New("not found"))
			},
			wantErr:     true,
			errContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userRepo := mock_repository.NewMockIUserRepository(ctrl)
			tt.setupMocks(userRepo)

			interactor := NewUserInteractor(userRepo)

			result, err := interactor.GetMe(context.Background(), tt.userID)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.userID, result.ID)
		})
	}
}

func TestUserInteractor_UpdateMe(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		input        *input.UpdateUserInput
		setupMocks   func(userRepo *mock_repository.MockIUserRepository)
		wantErr      bool
		errContains  string
		expectedName string
	}{
		{
			name: "success - update name",
			input: &input.UpdateUserInput{
				UserID: "user-id-1",
				Name:   strPtr("Updated Name"),
			},
			setupMocks: func(userRepo *mock_repository.MockIUserRepository) {
				userRepo.EXPECT().
					FindByID(gomock.Any(), "user-id-1").
					Return(&model.User{
						ID:       "user-id-1",
						TenantID: "tenant-id",
						Email:    "test@example.com",
						Name:     "Original Name",
						Role:     "member",
					}, nil)

				userRepo.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, u *model.User) (*model.User, error) {
						return &model.User{
							ID:       u.ID,
							TenantID: u.TenantID,
							Email:    u.Email,
							Name:     u.Name,
							Role:     u.Role,
						}, nil
					})
			},
			wantErr:      false,
			expectedName: "Updated Name",
		},
		{
			name: "success - no update (nil name)",
			input: &input.UpdateUserInput{
				UserID: "user-id-1",
				Name:   nil,
			},
			setupMocks: func(userRepo *mock_repository.MockIUserRepository) {
				userRepo.EXPECT().
					FindByID(gomock.Any(), "user-id-1").
					Return(&model.User{
						ID:       "user-id-1",
						TenantID: "tenant-id",
						Email:    "test@example.com",
						Name:     "Original Name",
						Role:     "member",
					}, nil)

				userRepo.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, u *model.User) (*model.User, error) {
						return &model.User{
							ID:       u.ID,
							TenantID: u.TenantID,
							Email:    u.Email,
							Name:     u.Name,
							Role:     u.Role,
						}, nil
					})
			},
			wantErr:      false,
			expectedName: "Original Name",
		},
		{
			name: "fail - user not found",
			input: &input.UpdateUserInput{
				UserID: "non-existent-id",
				Name:   strPtr("New Name"),
			},
			setupMocks: func(userRepo *mock_repository.MockIUserRepository) {
				userRepo.EXPECT().
					FindByID(gomock.Any(), "non-existent-id").
					Return(nil, errors.New("not found"))
			},
			wantErr:     true,
			errContains: "not found",
		},
		{
			name: "fail - update error",
			input: &input.UpdateUserInput{
				UserID: "user-id-1",
				Name:   strPtr("New Name"),
			},
			setupMocks: func(userRepo *mock_repository.MockIUserRepository) {
				userRepo.EXPECT().
					FindByID(gomock.Any(), "user-id-1").
					Return(&model.User{
						ID:       "user-id-1",
						TenantID: "tenant-id",
						Email:    "test@example.com",
						Name:     "Original Name",
						Role:     "member",
					}, nil)

				userRepo.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("database error"))
			},
			wantErr:     true,
			errContains: "failed to update",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userRepo := mock_repository.NewMockIUserRepository(ctrl)
			tt.setupMocks(userRepo)

			interactor := NewUserInteractor(userRepo)

			result, err := interactor.UpdateMe(context.Background(), tt.input)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedName, result.Name)
		})
	}
}

func strPtr(s string) *string {
	return &s
}
