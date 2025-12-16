package presenter

import (
	"net/http"
	"time"

	"good-todo-go/internal/presentation/public/api"
	"good-todo-go/internal/usecase/output"

	"github.com/labstack/echo/v4"
)

type IUserPresenter interface {
	GetMe(ctx echo.Context, out *output.UserOutput) error
	UpdateMe(ctx echo.Context, out *output.UserOutput) error
}

type UserPresenter struct{}

func NewUserPresenter() IUserPresenter {
	return &UserPresenter{}
}

func (p *UserPresenter) GetMe(ctx echo.Context, out *output.UserOutput) error {
	return ctx.JSON(http.StatusOK, toUserResponse(out))
}

func (p *UserPresenter) UpdateMe(ctx echo.Context, out *output.UserOutput) error {
	return ctx.JSON(http.StatusOK, toUserResponse(out))
}

func toUserResponse(out *output.UserOutput) *api.UserResponse {
	role := api.UserResponseRole(out.Role)
	createdAt, _ := time.Parse(time.RFC3339, out.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339, out.UpdatedAt)
	return &api.UserResponse{
		Id:            &out.ID,
		Email:         &out.Email,
		Name:          &out.Name,
		Role:          &role,
		EmailVerified: &out.EmailVerified,
		TenantId:      &out.TenantID,
		CreatedAt:     &createdAt,
		UpdatedAt:     &updatedAt,
	}
}
