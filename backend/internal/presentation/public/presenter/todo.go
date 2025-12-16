package presenter

import (
	"net/http"
	"time"

	"good-todo-go/internal/presentation/public/api"
	"good-todo-go/internal/usecase/output"

	"github.com/labstack/echo/v4"
)

type ITodoPresenter interface {
	GetTodos(ctx echo.Context, out *output.TodoListOutput) error
	GetTodo(ctx echo.Context, out *output.TodoOutput) error
	CreateTodo(ctx echo.Context, out *output.TodoOutput) error
	UpdateTodo(ctx echo.Context, out *output.TodoOutput) error
	DeleteTodo(ctx echo.Context) error
}

type TodoPresenter struct{}

func NewTodoPresenter() ITodoPresenter {
	return &TodoPresenter{}
}

func (p *TodoPresenter) GetTodos(ctx echo.Context, out *output.TodoListOutput) error {
	todos := make([]api.TodoResponse, len(out.Todos))
	for i, t := range out.Todos {
		todos[i] = *toTodoResponse(t)
	}

	return ctx.JSON(http.StatusOK, api.TodoListResponse{
		Todos: &todos,
		Total: &out.Total,
	})
}

func (p *TodoPresenter) GetTodo(ctx echo.Context, out *output.TodoOutput) error {
	return ctx.JSON(http.StatusOK, toTodoResponse(out))
}

func (p *TodoPresenter) CreateTodo(ctx echo.Context, out *output.TodoOutput) error {
	return ctx.JSON(http.StatusCreated, toTodoResponse(out))
}

func (p *TodoPresenter) UpdateTodo(ctx echo.Context, out *output.TodoOutput) error {
	return ctx.JSON(http.StatusOK, toTodoResponse(out))
}

func (p *TodoPresenter) DeleteTodo(ctx echo.Context) error {
	return ctx.NoContent(http.StatusNoContent)
}

func toTodoResponse(out *output.TodoOutput) *api.TodoResponse {
	createdAt, _ := time.Parse(time.RFC3339, out.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339, out.UpdatedAt)

	resp := &api.TodoResponse{
		Id:          &out.ID,
		Title:       &out.Title,
		Description: &out.Description,
		Completed:   &out.Completed,
		CreatedAt:   &createdAt,
		UpdatedAt:   &updatedAt,
	}

	if out.DueDate != nil {
		dueDate, _ := time.Parse(time.RFC3339, *out.DueDate)
		resp.DueDate = &dueDate
	}

	return resp
}
