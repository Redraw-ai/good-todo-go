package router

import (
	"github.com/labstack/echo/v4"
)

func (s *Server) GetMe(c echo.Context) error {
	return s.userController.GetMe(c)
}

func (s *Server) UpdateMe(c echo.Context) error {
	return s.userController.UpdateMe(c)
}
