package router

import (
	"good-todo-go/internal/presentation/public/api"

	"github.com/labstack/echo/v4"
)

func (s *Server) GetTodos(c echo.Context, params api.GetTodosParams) error {
	return s.todoController.GetTodos(c, params)
}

func (s *Server) CreateTodo(c echo.Context) error {
	return s.todoController.CreateTodo(c)
}

func (s *Server) DeleteTodo(c echo.Context, todoId string) error {
	return s.todoController.DeleteTodo(c, todoId)
}

func (s *Server) GetTodo(c echo.Context, todoId string) error {
	return s.todoController.GetTodo(c, todoId)
}

func (s *Server) UpdateTodo(c echo.Context, todoId string) error {
	return s.todoController.UpdateTodo(c, todoId)
}
