package controller

import (
	"net/http"
	"time"

	"good-todo-go/internal/presentation/public/api"
	"good-todo-go/internal/presentation/public/presenter"
	"good-todo-go/internal/presentation/public/router/context_keys"
	"good-todo-go/internal/usecase"
	"good-todo-go/internal/usecase/input"

	"github.com/labstack/echo/v4"
)

type TodoController struct {
	todoUsecase   usecase.ITodoInteractor
	todoPresenter presenter.ITodoPresenter
}

func NewTodoController(
	todoUsecase usecase.ITodoInteractor,
	todoPresenter presenter.ITodoPresenter,
) *TodoController {
	return &TodoController{
		todoUsecase:   todoUsecase,
		todoPresenter: todoPresenter,
	}
}

func (c *TodoController) GetTodos(ctx echo.Context, params api.GetTodosParams) error {
	userID, ok := ctx.Get(context_keys.UserIDContextKey).(string)
	if !ok || userID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	limit := 20
	offset := 0
	if params.Limit != nil {
		limit = *params.Limit
	}
	if params.Offset != nil {
		offset = *params.Offset
	}

	in := &input.GetTodosInput{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	}

	out, err := c.todoUsecase.GetTodos(ctx.Request().Context(), in)
	if err != nil {
		return handleError(err)
	}

	return c.todoPresenter.GetTodos(ctx, out)
}

func (c *TodoController) GetPublicTodos(ctx echo.Context, params api.GetPublicTodosParams) error {
	limit := 20
	offset := 0
	if params.Limit != nil {
		limit = *params.Limit
	}
	if params.Offset != nil {
		offset = *params.Offset
	}

	in := &input.GetPublicTodosInput{
		Limit:  limit,
		Offset: offset,
	}

	out, err := c.todoUsecase.GetPublicTodos(ctx.Request().Context(), in)
	if err != nil {
		return handleError(err)
	}

	return c.todoPresenter.GetTodos(ctx, out)
}

func (c *TodoController) GetTodo(ctx echo.Context, todoID string) error {
	userID, ok := ctx.Get(context_keys.UserIDContextKey).(string)
	if !ok || userID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	out, err := c.todoUsecase.GetTodo(ctx.Request().Context(), todoID, userID)
	if err != nil {
		return handleError(err)
	}

	return c.todoPresenter.GetTodo(ctx, out)
}

func (c *TodoController) CreateTodo(ctx echo.Context) error {
	userID, ok := ctx.Get(context_keys.UserIDContextKey).(string)
	if !ok || userID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	tenantID, ok := ctx.Get(context_keys.TenantIDContextKey).(string)
	if !ok || tenantID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	var req api.CreateTodoRequest
	if err := ctx.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if req.Title == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "title is required")
	}

	description := ""
	if req.Description != nil {
		description = *req.Description
	}

	isPublic := false
	if req.IsPublic != nil {
		isPublic = *req.IsPublic
	}

	var dueDate *time.Time
	if req.DueDate != nil {
		dueDate = req.DueDate
	}

	in := &input.CreateTodoInput{
		UserID:      userID,
		TenantID:    tenantID,
		Title:       req.Title,
		Description: description,
		IsPublic:    isPublic,
		DueDate:     dueDate,
	}

	out, err := c.todoUsecase.CreateTodo(ctx.Request().Context(), in)
	if err != nil {
		return handleError(err)
	}

	return c.todoPresenter.CreateTodo(ctx, out)
}

func (c *TodoController) UpdateTodo(ctx echo.Context, todoID string) error {
	userID, ok := ctx.Get(context_keys.UserIDContextKey).(string)
	if !ok || userID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	var req api.UpdateTodoRequest
	if err := ctx.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	in := &input.UpdateTodoInput{
		TodoID:      todoID,
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		Completed:   req.Completed,
		IsPublic:    req.IsPublic,
		DueDate:     req.DueDate,
	}

	out, err := c.todoUsecase.UpdateTodo(ctx.Request().Context(), in)
	if err != nil {
		return handleError(err)
	}

	return c.todoPresenter.UpdateTodo(ctx, out)
}

func (c *TodoController) DeleteTodo(ctx echo.Context, todoID string) error {
	userID, ok := ctx.Get(context_keys.UserIDContextKey).(string)
	if !ok || userID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	if err := c.todoUsecase.DeleteTodo(ctx.Request().Context(), todoID, userID); err != nil {
		return handleError(err)
	}

	return c.todoPresenter.DeleteTodo(ctx)
}
