package router

import (
	"good-todo-go/internal/pkg"
	"good-todo-go/internal/presentation/public/controller"
	"good-todo-go/internal/presentation/public/router/middleware"

	"github.com/labstack/echo/v4"
)

type Server struct {
	authController *controller.AuthController
	jwtService     *pkg.JWTService
}

func NewServer(
	authController *controller.AuthController,
	jwtService *pkg.JWTService,
) *Server {
	return &Server{
		authController: authController,
		jwtService:     jwtService,
	}
}

func (s *Server) SetupRoutes(e *echo.Echo) {
	// Health check
	e.GET("/health", s.HealthCheck)

	// Auth routes (no authentication required)
	e.POST("/auth/register", s.Register)
	e.POST("/auth/login", s.Login)
	e.POST("/auth/verify-email", s.VerifyEmail)
	e.POST("/auth/refresh", s.RefreshToken)

	// Protected routes
	protected := e.Group("")
	protected.Use(middleware.JWTAuth(s.jwtService))

	// TODO: Add protected routes here
	// protected.GET("/me", s.GetMe)
	// protected.PUT("/me", s.UpdateMe)
	// protected.GET("/todos", s.GetTodos)
	// ...
}
