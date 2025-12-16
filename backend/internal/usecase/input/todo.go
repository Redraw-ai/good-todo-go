package input

import "time"

type CreateTodoInput struct {
	UserID      string
	TenantID    string
	Title       string
	Description string
	DueDate     *time.Time
}

type UpdateTodoInput struct {
	TodoID      string
	UserID      string
	Title       *string
	Description *string
	Completed   *bool
	DueDate     *time.Time
}

type GetTodosInput struct {
	UserID string
	Limit  int
	Offset int
}
