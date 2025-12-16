package output

import "good-todo-go/internal/domain/model"

type AuthOutput struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	ExpiresIn    int
	User         *UserOutput
}

type UserOutput struct {
	ID            string
	Email         string
	Name          string
	Role          string
	EmailVerified bool
	TenantID      string
	CreatedAt     string
	UpdatedAt     string
}

func NewUserOutput(user *model.User) *UserOutput {
	return &UserOutput{
		ID:            user.ID,
		Email:         user.Email,
		Name:          user.Name,
		Role:          user.Role,
		EmailVerified: user.EmailVerified,
		TenantID:      user.TenantID,
		CreatedAt:     user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

type VerifyEmailOutput struct {
	Message string
}
