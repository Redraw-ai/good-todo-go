package model

import "time"

type User struct {
	ID                         string
	TenantID                   string
	Email                      string
	PasswordHash               string
	Name                       string
	Role                       string
	EmailVerified              bool
	VerificationToken          *string
	VerificationTokenExpiresAt *time.Time
	CreatedAt                  time.Time
	UpdatedAt                  time.Time
}

type Tenant struct {
	ID        string
	Name      string
	Slug      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
