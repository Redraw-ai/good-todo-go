package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"time"

	"good-todo-go/internal/domain/model"
	"good-todo-go/internal/domain/repository"
	"good-todo-go/internal/pkg"
	"good-todo-go/internal/pkg/cerror"
	"good-todo-go/internal/usecase/input"
	"good-todo-go/internal/usecase/output"
)

type IAuthInteractor interface {
	Register(ctx context.Context, in *input.RegisterInput) (*output.AuthOutput, error)
	Login(ctx context.Context, in *input.LoginInput) (*output.AuthOutput, error)
	VerifyEmail(ctx context.Context, in *input.VerifyEmailInput) (*output.VerifyEmailOutput, error)
	RefreshToken(ctx context.Context, in *input.RefreshTokenInput) (*output.AuthOutput, error)
}

type AuthInteractor struct {
	authRepo   repository.IAuthRepository
	jwtService *pkg.JWTService
	uuidGen    pkg.IUUIDGenerator
}

func NewAuthInteractor(
	authRepo repository.IAuthRepository,
	jwtService *pkg.JWTService,
	uuidGen pkg.IUUIDGenerator,
) IAuthInteractor {
	return &AuthInteractor{
		authRepo:   authRepo,
		jwtService: jwtService,
		uuidGen:    uuidGen,
	}
}

func (i *AuthInteractor) Register(ctx context.Context, in *input.RegisterInput) (*output.AuthOutput, error) {
	// Find or create tenant
	tenant, err := i.authRepo.FindTenantBySlug(ctx, in.TenantSlug)
	if err != nil {
		// Create new tenant if not exists
		tenant = &model.Tenant{
			ID:   i.uuidGen.Generate(),
			Name: in.TenantSlug, // Use slug as name initially
			Slug: in.TenantSlug,
		}
		tenant, err = i.authRepo.CreateTenant(ctx, tenant)
		if err != nil {
			log.Printf("failed to create tenant: %v", err)
			return nil, cerror.NewInternalServerError("failed to create tenant", err)
		}
	}

	// Check if user already exists
	existingUser, _ := i.authRepo.FindUserByEmail(ctx, tenant.ID, in.Email)
	if existingUser != nil {
		return nil, cerror.NewConflict("email already exists", nil)
	}

	// Hash password
	passwordHash, err := pkg.HashPassword(in.Password)
	if err != nil {
		return nil, cerror.NewInternalServerError("failed to hash password", err)
	}

	// Generate verification token
	verificationToken, err := generateVerificationToken()
	if err != nil {
		return nil, cerror.NewInternalServerError("failed to generate verification token", err)
	}
	tokenExpiry := time.Now().Add(24 * time.Hour)

	// Create user
	user := &model.User{
		ID:                         i.uuidGen.Generate(),
		TenantID:                   tenant.ID,
		Email:                      in.Email,
		PasswordHash:               passwordHash,
		Name:                       in.Name,
		Role:                       "member",
		EmailVerified:              false,
		VerificationToken:          &verificationToken,
		VerificationTokenExpiresAt: &tokenExpiry,
	}

	user, err = i.authRepo.CreateUser(ctx, user)
	if err != nil {
		return nil, cerror.NewInternalServerError("failed to create user", err)
	}

	// Generate JWT tokens
	tokenPair, err := i.jwtService.GenerateTokenPair(user.ID, user.TenantID, user.Email, user.Role)
	if err != nil {
		return nil, cerror.NewInternalServerError("failed to generate tokens", err)
	}

	// TODO: Send verification email

	return &output.AuthOutput{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    tokenPair.ExpiresIn,
		User:         output.NewUserOutput(user),
	}, nil
}

func (i *AuthInteractor) Login(ctx context.Context, in *input.LoginInput) (*output.AuthOutput, error) {
	// Find tenant
	tenant, err := i.authRepo.FindTenantBySlug(ctx, in.TenantSlug)
	if err != nil {
		return nil, cerror.NewUnauthorized("invalid credentials", nil)
	}

	// Find user
	user, err := i.authRepo.FindUserByEmail(ctx, tenant.ID, in.Email)
	if err != nil {
		return nil, cerror.NewUnauthorized("invalid credentials", nil)
	}

	// Verify password
	if !pkg.CheckPasswordHash(in.Password, user.PasswordHash) {
		return nil, cerror.NewUnauthorized("invalid credentials", nil)
	}

	// Generate JWT tokens
	tokenPair, err := i.jwtService.GenerateTokenPair(user.ID, user.TenantID, user.Email, user.Role)
	if err != nil {
		return nil, cerror.NewInternalServerError("failed to generate tokens", err)
	}

	return &output.AuthOutput{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    tokenPair.ExpiresIn,
		User:         output.NewUserOutput(user),
	}, nil
}

func (i *AuthInteractor) VerifyEmail(ctx context.Context, in *input.VerifyEmailInput) (*output.VerifyEmailOutput, error) {
	// Find user by verification token
	user, err := i.authRepo.FindUserByVerificationToken(ctx, in.Token)
	if err != nil {
		return nil, cerror.NewBadRequest("invalid or expired token", nil)
	}

	// Check if token is expired
	if user.VerificationTokenExpiresAt != nil && user.VerificationTokenExpiresAt.Before(time.Now()) {
		return nil, cerror.NewBadRequest("verification token has expired", nil)
	}

	// Update user
	user.EmailVerified = true
	user.VerificationToken = nil
	user.VerificationTokenExpiresAt = nil

	_, err = i.authRepo.UpdateUser(ctx, user)
	if err != nil {
		return nil, cerror.NewInternalServerError("failed to update user", err)
	}

	return &output.VerifyEmailOutput{
		Message: "Email verified successfully",
	}, nil
}

func (i *AuthInteractor) RefreshToken(ctx context.Context, in *input.RefreshTokenInput) (*output.AuthOutput, error) {
	// Validate refresh token
	claims, err := i.jwtService.ValidateRefreshToken(in.RefreshToken)
	if err != nil {
		return nil, cerror.NewUnauthorized("invalid refresh token", err)
	}

	// Find user to ensure they still exist
	user, err := i.authRepo.FindUserByID(ctx, claims.UserID)
	if err != nil {
		return nil, cerror.NewUnauthorized("user not found", nil)
	}

	// Generate new token pair
	tokenPair, err := i.jwtService.GenerateTokenPair(user.ID, user.TenantID, user.Email, user.Role)
	if err != nil {
		return nil, cerror.NewInternalServerError("failed to generate tokens", err)
	}

	return &output.AuthOutput{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    tokenPair.ExpiresIn,
		User:         output.NewUserOutput(user),
	}, nil
}

func generateVerificationToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
