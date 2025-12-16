package input

type RegisterInput struct {
	Email      string
	Password   string
	Name       string
	TenantSlug string
}

type LoginInput struct {
	Email      string
	Password   string
	TenantSlug string
}

type VerifyEmailInput struct {
	Token string
}

type RefreshTokenInput struct {
	RefreshToken string
}
