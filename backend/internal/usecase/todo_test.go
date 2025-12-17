package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"good-todo-go/internal/domain/model"
	mock_repository "good-todo-go/internal/domain/repository/mock"
	"good-todo-go/internal/pkg/cerror"
	mock_pkg "good-todo-go/internal/pkg/mock"
	"good-todo-go/internal/usecase/input"
	"good-todo-go/internal/usecase/output"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestTodoInteractor_GetTodos(t *testing.T) {
	t.Parallel()

	type want struct {
		output *output.TodoListOutput
		err    error
	}

	now := time.Now()

	tests := []struct {
		name    string
		usecase func(ctx context.Context, ctrl *gomock.Controller) ITodoInteractor
		input   *input.GetTodosInput
		want    want
	}{
		{
			name: "success - get todos",
			usecase: func(ctx context.Context, ctrl *gomock.Controller) ITodoInteractor {
				todoRepo := mock_repository.NewMockITodoRepository(ctrl)
				userRepo := mock_repository.NewMockIUserRepository(ctrl)
				uuidGen := mock_pkg.NewMockIUUIDGenerator(ctrl)

				todos := []*model.Todo{
					{
						ID:          "todo-1",
						UserID:      "user-1",
						TenantID:    "tenant-1",
						Title:       "Test Todo 1",
						Description: "Description 1",
						Completed:   false,
						IsPublic:    false,
						CreatedAt:   now,
						UpdatedAt:   now,
					},
					{
						ID:          "todo-2",
						UserID:      "user-1",
						TenantID:    "tenant-1",
						Title:       "Test Todo 2",
						Description: "Description 2",
						Completed:   true,
						IsPublic:    true,
						CreatedAt:   now,
						UpdatedAt:   now,
					},
				}

				todoRepo.EXPECT().
					FindByUserID(ctx, "user-1", 20, 0).
					Return(todos, nil)

				todoRepo.EXPECT().
					CountByUserID(ctx, "user-1").
					Return(2, nil)

				return &TodoInteractor{
					todoRepo: todoRepo,
					userRepo: userRepo,
					uuidGen:  uuidGen,
				}
			},
			input: &input.GetTodosInput{
				UserID: "user-1",
				Limit:  0, // should default to 20
				Offset: 0,
			},
			want: want{
				output: &output.TodoListOutput{
					Todos: []*output.TodoOutput{
						{
							ID:          "todo-1",
							UserID:      "user-1",
							Title:       "Test Todo 1",
							Description: "Description 1",
							Completed:   false,
							IsPublic:    false,
							DueDate:     nil,
							CreatedAt:   now.Format("2006-01-02T15:04:05Z07:00"),
							UpdatedAt:   now.Format("2006-01-02T15:04:05Z07:00"),
						},
						{
							ID:          "todo-2",
							UserID:      "user-1",
							Title:       "Test Todo 2",
							Description: "Description 2",
							Completed:   true,
							IsPublic:    true,
							DueDate:     nil,
							CreatedAt:   now.Format("2006-01-02T15:04:05Z07:00"),
							UpdatedAt:   now.Format("2006-01-02T15:04:05Z07:00"),
						},
					},
					Total: 2,
				},
				err: nil,
			},
		},
		{
			name: "success - limit capped at 100",
			usecase: func(ctx context.Context, ctrl *gomock.Controller) ITodoInteractor {
				todoRepo := mock_repository.NewMockITodoRepository(ctrl)
				userRepo := mock_repository.NewMockIUserRepository(ctrl)
				uuidGen := mock_pkg.NewMockIUUIDGenerator(ctrl)

				todoRepo.EXPECT().
					FindByUserID(ctx, "user-1", 100, 0). // capped at 100
					Return([]*model.Todo{}, nil)

				todoRepo.EXPECT().
					CountByUserID(ctx, "user-1").
					Return(0, nil)

				return &TodoInteractor{
					todoRepo: todoRepo,
					userRepo: userRepo,
					uuidGen:  uuidGen,
				}
			},
			input: &input.GetTodosInput{
				UserID: "user-1",
				Limit:  500, // should be capped to 100
				Offset: 0,
			},
			want: want{
				output: &output.TodoListOutput{
					Todos: []*output.TodoOutput{},
					Total: 0,
				},
				err: nil,
			},
		},
		{
			name: "error - repository error on find",
			usecase: func(ctx context.Context, ctrl *gomock.Controller) ITodoInteractor {
				todoRepo := mock_repository.NewMockITodoRepository(ctrl)
				userRepo := mock_repository.NewMockIUserRepository(ctrl)
				uuidGen := mock_pkg.NewMockIUUIDGenerator(ctrl)

				todoRepo.EXPECT().
					FindByUserID(ctx, "user-1", 20, 0).
					Return(nil, errors.New("db error"))

				return &TodoInteractor{
					todoRepo: todoRepo,
					userRepo: userRepo,
					uuidGen:  uuidGen,
				}
			},
			input: &input.GetTodosInput{
				UserID: "user-1",
				Limit:  0,
				Offset: 0,
			},
			want: want{
				output: nil,
				err:    cerror.NewInternalServerError("failed to get todos", errors.New("db error")),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := context.Background()
			interactor := tt.usecase(ctx, ctrl)

			gotOutput, gotErr := interactor.GetTodos(ctx, tt.input)

			if tt.want.err != nil {
				assert.Error(t, gotErr)
				assert.Nil(t, gotOutput)
			} else {
				assert.NoError(t, gotErr)
				assert.Equal(t, tt.want.output, gotOutput)
			}
		})
	}
}

func TestTodoInteractor_GetTodo(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name    string
		usecase func(ctx context.Context, ctrl *gomock.Controller) ITodoInteractor
		todoID  string
		userID  string
		wantErr bool
	}{
		{
			name: "success - owner can access private todo",
			usecase: func(ctx context.Context, ctrl *gomock.Controller) ITodoInteractor {
				todoRepo := mock_repository.NewMockITodoRepository(ctrl)
				userRepo := mock_repository.NewMockIUserRepository(ctrl)
				uuidGen := mock_pkg.NewMockIUUIDGenerator(ctrl)

				todoRepo.EXPECT().
					FindByID(ctx, "todo-1").
					Return(&model.Todo{
						ID:          "todo-1",
						UserID:      "user-1",
						TenantID:    "tenant-1",
						Title:       "Private Todo",
						Description: "Private",
						IsPublic:    false,
						CreatedAt:   now,
						UpdatedAt:   now,
					}, nil)

				return &TodoInteractor{
					todoRepo: todoRepo,
					userRepo: userRepo,
					uuidGen:  uuidGen,
				}
			},
			todoID:  "todo-1",
			userID:  "user-1", // owner
			wantErr: false,
		},
		{
			name: "success - other user can access public todo",
			usecase: func(ctx context.Context, ctrl *gomock.Controller) ITodoInteractor {
				todoRepo := mock_repository.NewMockITodoRepository(ctrl)
				userRepo := mock_repository.NewMockIUserRepository(ctrl)
				uuidGen := mock_pkg.NewMockIUUIDGenerator(ctrl)

				todoRepo.EXPECT().
					FindByID(ctx, "todo-1").
					Return(&model.Todo{
						ID:          "todo-1",
						UserID:      "user-1",
						TenantID:    "tenant-1",
						Title:       "Public Todo",
						Description: "Public",
						IsPublic:    true,
						CreatedAt:   now,
						UpdatedAt:   now,
					}, nil)

				return &TodoInteractor{
					todoRepo: todoRepo,
					userRepo: userRepo,
					uuidGen:  uuidGen,
				}
			},
			todoID:  "todo-1",
			userID:  "user-2", // not owner but todo is public
			wantErr: false,
		},
		{
			name: "error - other user cannot access private todo",
			usecase: func(ctx context.Context, ctrl *gomock.Controller) ITodoInteractor {
				todoRepo := mock_repository.NewMockITodoRepository(ctrl)
				userRepo := mock_repository.NewMockIUserRepository(ctrl)
				uuidGen := mock_pkg.NewMockIUUIDGenerator(ctrl)

				todoRepo.EXPECT().
					FindByID(ctx, "todo-1").
					Return(&model.Todo{
						ID:          "todo-1",
						UserID:      "user-1",
						TenantID:    "tenant-1",
						Title:       "Private Todo",
						Description: "Private",
						IsPublic:    false,
						CreatedAt:   now,
						UpdatedAt:   now,
					}, nil)

				return &TodoInteractor{
					todoRepo: todoRepo,
					userRepo: userRepo,
					uuidGen:  uuidGen,
				}
			},
			todoID:  "todo-1",
			userID:  "user-2", // not owner and todo is private
			wantErr: true,
		},
		{
			name: "error - todo not found",
			usecase: func(ctx context.Context, ctrl *gomock.Controller) ITodoInteractor {
				todoRepo := mock_repository.NewMockITodoRepository(ctrl)
				userRepo := mock_repository.NewMockIUserRepository(ctrl)
				uuidGen := mock_pkg.NewMockIUUIDGenerator(ctrl)

				todoRepo.EXPECT().
					FindByID(ctx, "todo-not-exist").
					Return(nil, errors.New("not found"))

				return &TodoInteractor{
					todoRepo: todoRepo,
					userRepo: userRepo,
					uuidGen:  uuidGen,
				}
			},
			todoID:  "todo-not-exist",
			userID:  "user-1",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := context.Background()
			interactor := tt.usecase(ctx, ctrl)

			gotOutput, gotErr := interactor.GetTodo(ctx, tt.todoID, tt.userID)

			if tt.wantErr {
				assert.Error(t, gotErr)
				assert.Nil(t, gotOutput)
			} else {
				assert.NoError(t, gotErr)
				assert.NotNil(t, gotOutput)
			}
		})
	}
}

func TestTodoInteractor_CreateTodo(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name    string
		usecase func(ctx context.Context, ctrl *gomock.Controller) ITodoInteractor
		input   *input.CreateTodoInput
		wantErr bool
	}{
		{
			name: "success - create todo",
			usecase: func(ctx context.Context, ctrl *gomock.Controller) ITodoInteractor {
				todoRepo := mock_repository.NewMockITodoRepository(ctrl)
				userRepo := mock_repository.NewMockIUserRepository(ctrl)
				uuidGen := mock_pkg.NewMockIUUIDGenerator(ctrl)

				uuidGen.EXPECT().
					Generate().
					Return("generated-uuid")

				todoRepo.EXPECT().
					Create(ctx, gomock.Any()).
					DoAndReturn(func(ctx context.Context, todo *model.Todo) (*model.Todo, error) {
						todo.CreatedAt = now
						todo.UpdatedAt = now
						return todo, nil
					})

				return &TodoInteractor{
					todoRepo: todoRepo,
					userRepo: userRepo,
					uuidGen:  uuidGen,
				}
			},
			input: &input.CreateTodoInput{
				UserID:      "user-1",
				TenantID:    "tenant-1",
				Title:       "New Todo",
				Description: "New Description",
				IsPublic:    false,
				DueDate:     nil,
			},
			wantErr: false,
		},
		{
			name: "error - repository error",
			usecase: func(ctx context.Context, ctrl *gomock.Controller) ITodoInteractor {
				todoRepo := mock_repository.NewMockITodoRepository(ctrl)
				userRepo := mock_repository.NewMockIUserRepository(ctrl)
				uuidGen := mock_pkg.NewMockIUUIDGenerator(ctrl)

				uuidGen.EXPECT().
					Generate().
					Return("generated-uuid")

				todoRepo.EXPECT().
					Create(ctx, gomock.Any()).
					Return(nil, errors.New("db error"))

				return &TodoInteractor{
					todoRepo: todoRepo,
					userRepo: userRepo,
					uuidGen:  uuidGen,
				}
			},
			input: &input.CreateTodoInput{
				UserID:      "user-1",
				TenantID:    "tenant-1",
				Title:       "New Todo",
				Description: "New Description",
				IsPublic:    false,
				DueDate:     nil,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := context.Background()
			interactor := tt.usecase(ctx, ctrl)

			gotOutput, gotErr := interactor.CreateTodo(ctx, tt.input)

			if tt.wantErr {
				assert.Error(t, gotErr)
				assert.Nil(t, gotOutput)
			} else {
				assert.NoError(t, gotErr)
				assert.NotNil(t, gotOutput)
				assert.Equal(t, "generated-uuid", gotOutput.ID)
				assert.Equal(t, tt.input.Title, gotOutput.Title)
			}
		})
	}
}

func TestTodoInteractor_UpdateTodo(t *testing.T) {
	t.Parallel()

	now := time.Now()
	newTitle := "Updated Title"

	tests := []struct {
		name    string
		usecase func(ctx context.Context, ctrl *gomock.Controller) ITodoInteractor
		input   *input.UpdateTodoInput
		wantErr bool
	}{
		{
			name: "success - owner can update",
			usecase: func(ctx context.Context, ctrl *gomock.Controller) ITodoInteractor {
				todoRepo := mock_repository.NewMockITodoRepository(ctrl)
				userRepo := mock_repository.NewMockIUserRepository(ctrl)
				uuidGen := mock_pkg.NewMockIUUIDGenerator(ctrl)

				todoRepo.EXPECT().
					FindByID(ctx, "todo-1").
					Return(&model.Todo{
						ID:          "todo-1",
						UserID:      "user-1",
						TenantID:    "tenant-1",
						Title:       "Original Title",
						Description: "Original",
						IsPublic:    false,
						CreatedAt:   now,
						UpdatedAt:   now,
					}, nil)

				todoRepo.EXPECT().
					Update(ctx, gomock.Any()).
					DoAndReturn(func(ctx context.Context, todo *model.Todo) (*model.Todo, error) {
						return todo, nil
					})

				return &TodoInteractor{
					todoRepo: todoRepo,
					userRepo: userRepo,
					uuidGen:  uuidGen,
				}
			},
			input: &input.UpdateTodoInput{
				TodoID: "todo-1",
				UserID: "user-1", // owner
				Title:  &newTitle,
			},
			wantErr: false,
		},
		{
			name: "error - non-owner cannot update",
			usecase: func(ctx context.Context, ctrl *gomock.Controller) ITodoInteractor {
				todoRepo := mock_repository.NewMockITodoRepository(ctrl)
				userRepo := mock_repository.NewMockIUserRepository(ctrl)
				uuidGen := mock_pkg.NewMockIUUIDGenerator(ctrl)

				todoRepo.EXPECT().
					FindByID(ctx, "todo-1").
					Return(&model.Todo{
						ID:          "todo-1",
						UserID:      "user-1",
						TenantID:    "tenant-1",
						Title:       "Original Title",
						Description: "Original",
						IsPublic:    false,
						CreatedAt:   now,
						UpdatedAt:   now,
					}, nil)

				return &TodoInteractor{
					todoRepo: todoRepo,
					userRepo: userRepo,
					uuidGen:  uuidGen,
				}
			},
			input: &input.UpdateTodoInput{
				TodoID: "todo-1",
				UserID: "user-2", // not owner
				Title:  &newTitle,
			},
			wantErr: true,
		},
		{
			name: "error - todo not found",
			usecase: func(ctx context.Context, ctrl *gomock.Controller) ITodoInteractor {
				todoRepo := mock_repository.NewMockITodoRepository(ctrl)
				userRepo := mock_repository.NewMockIUserRepository(ctrl)
				uuidGen := mock_pkg.NewMockIUUIDGenerator(ctrl)

				todoRepo.EXPECT().
					FindByID(ctx, "todo-not-exist").
					Return(nil, errors.New("not found"))

				return &TodoInteractor{
					todoRepo: todoRepo,
					userRepo: userRepo,
					uuidGen:  uuidGen,
				}
			},
			input: &input.UpdateTodoInput{
				TodoID: "todo-not-exist",
				UserID: "user-1",
				Title:  &newTitle,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := context.Background()
			interactor := tt.usecase(ctx, ctrl)

			gotOutput, gotErr := interactor.UpdateTodo(ctx, tt.input)

			if tt.wantErr {
				assert.Error(t, gotErr)
				assert.Nil(t, gotOutput)
			} else {
				assert.NoError(t, gotErr)
				assert.NotNil(t, gotOutput)
				assert.Equal(t, newTitle, gotOutput.Title)
			}
		})
	}
}

func TestTodoInteractor_DeleteTodo(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name    string
		usecase func(ctx context.Context, ctrl *gomock.Controller) ITodoInteractor
		todoID  string
		userID  string
		wantErr bool
	}{
		{
			name: "success - owner can delete",
			usecase: func(ctx context.Context, ctrl *gomock.Controller) ITodoInteractor {
				todoRepo := mock_repository.NewMockITodoRepository(ctrl)
				userRepo := mock_repository.NewMockIUserRepository(ctrl)
				uuidGen := mock_pkg.NewMockIUUIDGenerator(ctrl)

				todoRepo.EXPECT().
					FindByID(ctx, "todo-1").
					Return(&model.Todo{
						ID:          "todo-1",
						UserID:      "user-1",
						TenantID:    "tenant-1",
						Title:       "To be deleted",
						Description: "Delete me",
						IsPublic:    false,
						CreatedAt:   now,
						UpdatedAt:   now,
					}, nil)

				todoRepo.EXPECT().
					Delete(ctx, "todo-1").
					Return(nil)

				return &TodoInteractor{
					todoRepo: todoRepo,
					userRepo: userRepo,
					uuidGen:  uuidGen,
				}
			},
			todoID:  "todo-1",
			userID:  "user-1", // owner
			wantErr: false,
		},
		{
			name: "error - non-owner cannot delete",
			usecase: func(ctx context.Context, ctrl *gomock.Controller) ITodoInteractor {
				todoRepo := mock_repository.NewMockITodoRepository(ctrl)
				userRepo := mock_repository.NewMockIUserRepository(ctrl)
				uuidGen := mock_pkg.NewMockIUUIDGenerator(ctrl)

				todoRepo.EXPECT().
					FindByID(ctx, "todo-1").
					Return(&model.Todo{
						ID:          "todo-1",
						UserID:      "user-1",
						TenantID:    "tenant-1",
						Title:       "To be deleted",
						Description: "Delete me",
						IsPublic:    false,
						CreatedAt:   now,
						UpdatedAt:   now,
					}, nil)

				return &TodoInteractor{
					todoRepo: todoRepo,
					userRepo: userRepo,
					uuidGen:  uuidGen,
				}
			},
			todoID:  "todo-1",
			userID:  "user-2", // not owner
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := context.Background()
			interactor := tt.usecase(ctx, ctrl)

			gotErr := interactor.DeleteTodo(ctx, tt.todoID, tt.userID)

			if tt.wantErr {
				assert.Error(t, gotErr)
			} else {
				assert.NoError(t, gotErr)
			}
		})
	}
}
