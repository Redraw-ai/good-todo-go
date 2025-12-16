package router

import "github.com/labstack/echo/v4"

func (s *Server) Register(c echo.Context) error {
	return s.authController.Register(c)
}

func (s *Server) Login(c echo.Context) error {
	return s.authController.Login(c)
}

func (s *Server) VerifyEmail(c echo.Context) error {
	return s.authController.VerifyEmail(c)
}

func (s *Server) RefreshToken(c echo.Context) error {
	return s.authController.RefreshToken(c)
}
