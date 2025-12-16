package controller

import (
	"net/http"

	"good-todo-go/internal/pkg/cerror"
	"good-todo-go/internal/presentation/public/api"
	"good-todo-go/internal/presentation/public/presenter"
	"good-todo-go/internal/usecase"
	"good-todo-go/internal/usecase/input"

	"github.com/labstack/echo/v4"
)

type AuthController struct {
	authUsecase   usecase.IAuthInteractor
	authPresenter presenter.IAuthPresenter
}

func NewAuthController(
	authUsecase usecase.IAuthInteractor,
	authPresenter presenter.IAuthPresenter,
) *AuthController {
	return &AuthController{
		authUsecase:   authUsecase,
		authPresenter: authPresenter,
	}
}

func (c *AuthController) Register(ctx echo.Context) error {
	var req api.RegisterRequest
	if err := ctx.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	// Validate
	if string(req.Email) == "" || req.Password == "" || req.TenantSlug == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "email, password and tenant_slug are required")
	}
	if len(req.Password) < 8 {
		return echo.NewHTTPError(http.StatusBadRequest, "password must be at least 8 characters")
	}

	name := ""
	if req.Name != nil {
		name = *req.Name
	}

	in := &input.RegisterInput{
		Email:      string(req.Email),
		Password:   req.Password,
		Name:       name,
		TenantSlug: req.TenantSlug,
	}

	out, err := c.authUsecase.Register(ctx.Request().Context(), in)
	if err != nil {
		return handleError(err)
	}

	return c.authPresenter.Register(ctx, out)
}

func (c *AuthController) Login(ctx echo.Context) error {
	var req api.LoginRequest
	if err := ctx.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if string(req.Email) == "" || req.Password == "" || req.TenantSlug == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "email, password and tenant_slug are required")
	}

	in := &input.LoginInput{
		Email:      string(req.Email),
		Password:   req.Password,
		TenantSlug: req.TenantSlug,
	}

	out, err := c.authUsecase.Login(ctx.Request().Context(), in)
	if err != nil {
		return handleError(err)
	}

	return c.authPresenter.Login(ctx, out)
}

func (c *AuthController) VerifyEmail(ctx echo.Context) error {
	var req api.VerifyEmailRequest
	if err := ctx.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if req.Token == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "token is required")
	}

	in := &input.VerifyEmailInput{
		Token: req.Token,
	}

	out, err := c.authUsecase.VerifyEmail(ctx.Request().Context(), in)
	if err != nil {
		return handleError(err)
	}

	return c.authPresenter.VerifyEmail(ctx, out)
}

func (c *AuthController) RefreshToken(ctx echo.Context) error {
	var req api.RefreshTokenRequest
	if err := ctx.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if req.RefreshToken == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "refresh_token is required")
	}

	in := &input.RefreshTokenInput{
		RefreshToken: req.RefreshToken,
	}

	out, err := c.authUsecase.RefreshToken(ctx.Request().Context(), in)
	if err != nil {
		return handleError(err)
	}

	return c.authPresenter.RefreshToken(ctx, out)
}

func handleError(err error) error {
	if appErr, ok := err.(*cerror.AppError); ok {
		return echo.NewHTTPError(appErr.HTTPStatus, appErr.Message)
	}
	return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
}
