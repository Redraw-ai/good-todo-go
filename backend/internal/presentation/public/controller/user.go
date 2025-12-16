package controller

import (
	"net/http"

	"good-todo-go/internal/presentation/public/api"
	"good-todo-go/internal/presentation/public/presenter"
	"good-todo-go/internal/presentation/public/router/context_keys"
	"good-todo-go/internal/usecase"
	"good-todo-go/internal/usecase/input"

	"github.com/labstack/echo/v4"
)

type UserController struct {
	userUsecase   usecase.IUserInteractor
	userPresenter presenter.IUserPresenter
}

func NewUserController(
	userUsecase usecase.IUserInteractor,
	userPresenter presenter.IUserPresenter,
) *UserController {
	return &UserController{
		userUsecase:   userUsecase,
		userPresenter: userPresenter,
	}
}

func (c *UserController) GetMe(ctx echo.Context) error {
	userID, ok := ctx.Get(context_keys.UserIDContextKey).(string)
	if !ok || userID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	out, err := c.userUsecase.GetMe(ctx.Request().Context(), userID)
	if err != nil {
		return handleError(err)
	}

	return c.userPresenter.GetMe(ctx, out)
}

func (c *UserController) UpdateMe(ctx echo.Context) error {
	userID, ok := ctx.Get(context_keys.UserIDContextKey).(string)
	if !ok || userID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	var req api.UpdateUserRequest
	if err := ctx.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	in := &input.UpdateUserInput{
		UserID: userID,
		Name:   req.Name,
	}

	out, err := c.userUsecase.UpdateMe(ctx.Request().Context(), in)
	if err != nil {
		return handleError(err)
	}

	return c.userPresenter.UpdateMe(ctx, out)
}
