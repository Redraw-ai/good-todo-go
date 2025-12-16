package presenter

import (
	"net/http"

	"good-todo-go/internal/usecase/output"

	"github.com/labstack/echo/v4"
)

type IAuthPresenter interface {
	Register(ctx echo.Context, out *output.AuthOutput) error
	Login(ctx echo.Context, out *output.AuthOutput) error
	VerifyEmail(ctx echo.Context, out *output.VerifyEmailOutput) error
	RefreshToken(ctx echo.Context, out *output.AuthOutput) error
}

type AuthPresenter struct{}

func NewAuthPresenter() IAuthPresenter {
	return &AuthPresenter{}
}

type AuthResponse struct {
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	TokenType    string        `json:"token_type"`
	ExpiresIn    int           `json:"expires_in"`
	User         *UserResponse `json:"user"`
}

type UserResponse struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	Role          string `json:"role"`
	EmailVerified bool   `json:"email_verified"`
	TenantID      string `json:"tenant_id"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

func (p *AuthPresenter) Register(ctx echo.Context, out *output.AuthOutput) error {
	return ctx.JSON(http.StatusCreated, toAuthResponse(out))
}

func (p *AuthPresenter) Login(ctx echo.Context, out *output.AuthOutput) error {
	return ctx.JSON(http.StatusOK, toAuthResponse(out))
}

func (p *AuthPresenter) VerifyEmail(ctx echo.Context, out *output.VerifyEmailOutput) error {
	return ctx.JSON(http.StatusOK, map[string]string{
		"message": out.Message,
	})
}

func (p *AuthPresenter) RefreshToken(ctx echo.Context, out *output.AuthOutput) error {
	return ctx.JSON(http.StatusOK, toAuthResponse(out))
}

func toAuthResponse(out *output.AuthOutput) *AuthResponse {
	return &AuthResponse{
		AccessToken:  out.AccessToken,
		RefreshToken: out.RefreshToken,
		TokenType:    out.TokenType,
		ExpiresIn:    out.ExpiresIn,
		User: &UserResponse{
			ID:            out.User.ID,
			Email:         out.User.Email,
			Name:          out.User.Name,
			Role:          out.User.Role,
			EmailVerified: out.User.EmailVerified,
			TenantID:      out.User.TenantID,
			CreatedAt:     out.User.CreatedAt,
			UpdatedAt:     out.User.UpdatedAt,
		},
	}
}
