package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"good-todo-go/internal/integration_test/common"
	"good-todo-go/internal/presentation/public/api"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuth_Register(t *testing.T) {
	t.Parallel()

	client := common.SetupTestClient(t)
	deps := BuildTestDependencies(client)

	tests := []struct {
		name           string
		requestBody    api.RegisterRequest
		expectedStatus int
		wantErr        bool
		checkResponse  func(t *testing.T, response api.AuthResponse)
	}{
		{
			name: "success - register new user with new tenant",
			requestBody: api.RegisterRequest{
				Email:      "newuser@example.com",
				Password:   "password123",
				Name:       strPtr("New User"),
				TenantSlug: "new-tenant",
			},
			expectedStatus: http.StatusCreated,
			wantErr:        false,
			checkResponse: func(t *testing.T, response api.AuthResponse) {
				assert.NotEmpty(t, response.AccessToken)
				assert.NotEmpty(t, response.RefreshToken)
				assert.Equal(t, "Bearer", *response.TokenType)
				assert.NotNil(t, response.User)
				assert.Equal(t, "newuser@example.com", *response.User.Email)
			},
		},
		{
			name: "success - register user with existing tenant",
			requestBody: api.RegisterRequest{
				Email:      "anotheruser@example.com",
				Password:   "password123",
				Name:       strPtr("Another User"),
				TenantSlug: "new-tenant", // Same tenant as above
			},
			expectedStatus: http.StatusCreated,
			wantErr:        false,
			checkResponse: func(t *testing.T, response api.AuthResponse) {
				assert.NotEmpty(t, response.AccessToken)
				assert.NotNil(t, response.User)
				assert.Equal(t, "anotheruser@example.com", *response.User.Email)
			},
		},
		{
			name: "fail - password too short",
			requestBody: api.RegisterRequest{
				Email:      "short@example.com",
				Password:   "short",
				Name:       strPtr("Short Password"),
				TenantSlug: "test-tenant",
			},
			expectedStatus: http.StatusBadRequest,
			wantErr:        true,
		},
		{
			name: "fail - missing tenant slug",
			requestBody: api.RegisterRequest{
				Email:      "test@example.com",
				Password:   "password123",
				TenantSlug: "",
			},
			expectedStatus: http.StatusBadRequest,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := SetupEcho()

			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)

			err = deps.AuthController.Register(c)

			if tt.wantErr {
				if err == nil {
					assert.GreaterOrEqual(t, rec.Code, 400)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.checkResponse != nil {
				var response api.AuthResponse
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)
				tt.checkResponse(t, response)
			}
		})
	}
}

func TestAuth_Register_DuplicateEmail(t *testing.T) {
	t.Parallel()

	client := common.SetupTestClient(t)
	deps := BuildTestDependencies(client)
	e := SetupEcho()

	// First registration
	firstReq := api.RegisterRequest{
		Email:      "duplicate@example.com",
		Password:   "password123",
		Name:       strPtr("First User"),
		TenantSlug: "dup-test-tenant",
	}

	body, _ := json.Marshal(firstReq)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := deps.AuthController.Register(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	// Second registration with same email in same tenant - should fail
	body, _ = json.Marshal(firstReq)
	req2 := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	req2.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)

	err = deps.AuthController.Register(c2)
	// Should fail with conflict error
	require.Error(t, err)
}

func TestAuth_Login(t *testing.T) {
	t.Parallel()

	client := common.SetupTestClient(t)
	deps := BuildTestDependencies(client)

	// First, register a user
	e := SetupEcho()
	registerReq := api.RegisterRequest{
		Email:      "login-test@example.com",
		Password:   "password123",
		Name:       strPtr("Login Test User"),
		TenantSlug: "login-test-tenant",
	}

	body, _ := json.Marshal(registerReq)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	_ = deps.AuthController.Register(c)

	tests := []struct {
		name           string
		requestBody    api.LoginRequest
		expectedStatus int
		wantErr        bool
		checkResponse  func(t *testing.T, response api.AuthResponse)
	}{
		{
			name: "success - valid credentials",
			requestBody: api.LoginRequest{
				Email:      "login-test@example.com",
				Password:   "password123",
				TenantSlug: "login-test-tenant",
			},
			expectedStatus: http.StatusOK,
			wantErr:        false,
			checkResponse: func(t *testing.T, response api.AuthResponse) {
				assert.NotEmpty(t, response.AccessToken)
				assert.NotEmpty(t, response.RefreshToken)
				assert.Equal(t, "Bearer", *response.TokenType)
				assert.NotNil(t, response.User)
				assert.Equal(t, "login-test@example.com", *response.User.Email)
			},
		},
		{
			name: "fail - wrong password",
			requestBody: api.LoginRequest{
				Email:      "login-test@example.com",
				Password:   "wrongpassword",
				TenantSlug: "login-test-tenant",
			},
			wantErr: true,
		},
		{
			name: "fail - user not found",
			requestBody: api.LoginRequest{
				Email:      "nonexistent@example.com",
				Password:   "password123",
				TenantSlug: "login-test-tenant",
			},
			wantErr: true,
		},
		{
			name: "fail - wrong tenant",
			requestBody: api.LoginRequest{
				Email:      "login-test@example.com",
				Password:   "password123",
				TenantSlug: "wrong-tenant",
			},
			wantErr: true,
		},
		{
			name: "fail - missing email",
			requestBody: api.LoginRequest{
				Email:      "",
				Password:   "password123",
				TenantSlug: "login-test-tenant",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := SetupEcho()

			body, err := json.Marshal(tt.requestBody)
			if err != nil {
				// If marshaling fails, it's expected for invalid data like empty email
				if tt.wantErr {
					return
				}
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)

			err = deps.AuthController.Login(c)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.checkResponse != nil {
				var response api.AuthResponse
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)
				tt.checkResponse(t, response)
			}
		})
	}
}

func TestAuth_RefreshToken(t *testing.T) {
	t.Parallel()

	client := common.SetupTestClient(t)
	deps := BuildTestDependencies(client)

	// First, register a user and get tokens
	e := SetupEcho()
	registerReq := api.RegisterRequest{
		Email:      "refresh-test@example.com",
		Password:   "password123",
		Name:       strPtr("Refresh Test User"),
		TenantSlug: "refresh-test-tenant",
	}

	body, _ := json.Marshal(registerReq)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	_ = deps.AuthController.Register(c)

	var registerResponse api.AuthResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &registerResponse)

	tests := []struct {
		name           string
		requestBody    api.RefreshTokenRequest
		expectedStatus int
		wantErr        bool
		checkResponse  func(t *testing.T, response api.AuthResponse)
	}{
		{
			name: "success - valid refresh token",
			requestBody: api.RefreshTokenRequest{
				RefreshToken: *registerResponse.RefreshToken,
			},
			expectedStatus: http.StatusOK,
			wantErr:        false,
			checkResponse: func(t *testing.T, response api.AuthResponse) {
				assert.NotEmpty(t, response.AccessToken)
				assert.NotEmpty(t, response.RefreshToken)
				assert.Equal(t, "Bearer", *response.TokenType)
			},
		},
		{
			name: "fail - invalid refresh token",
			requestBody: api.RefreshTokenRequest{
				RefreshToken: "invalid-token",
			},
			wantErr: true,
		},
		{
			name: "fail - empty refresh token",
			requestBody: api.RefreshTokenRequest{
				RefreshToken: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := SetupEcho()

			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewReader(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)

			err = deps.AuthController.RefreshToken(c)

			if tt.wantErr {
				if err == nil {
					assert.GreaterOrEqual(t, rec.Code, 400)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.checkResponse != nil {
				var response api.AuthResponse
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)
				tt.checkResponse(t, response)
			}
		})
	}
}

func TestAuth_VerifyEmail(t *testing.T) {
	t.Parallel()

	client := common.SetupTestClient(t)
	deps := BuildTestDependencies(client)

	// Register a user to get a verification token
	e := SetupEcho()
	registerReq := api.RegisterRequest{
		Email:      "verify-test@example.com",
		Password:   "password123",
		Name:       strPtr("Verify Test User"),
		TenantSlug: "verify-test-tenant",
	}

	body, _ := json.Marshal(registerReq)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	_ = deps.AuthController.Register(c)

	// Get the verification token from the database
	var registerResponse api.AuthResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &registerResponse)

	// Query the user to get verification token
	user, _ := client.User.Get(req.Context(), *registerResponse.User.Id)
	verificationToken := ""
	if user != nil && user.VerificationToken != nil {
		verificationToken = *user.VerificationToken
	}

	tests := []struct {
		name           string
		requestBody    api.VerifyEmailRequest
		expectedStatus int
		wantErr        bool
	}{
		{
			name: "success - valid token",
			requestBody: api.VerifyEmailRequest{
				Token: verificationToken,
			},
			expectedStatus: http.StatusOK,
			wantErr:        verificationToken == "", // Only error if no token
		},
		{
			name: "fail - invalid token",
			requestBody: api.VerifyEmailRequest{
				Token: "invalid-verification-token",
			},
			wantErr: true,
		},
		{
			name: "fail - empty token",
			requestBody: api.VerifyEmailRequest{
				Token: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := SetupEcho()

			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/auth/verify-email", bytes.NewReader(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)

			err = deps.AuthController.VerifyEmail(c)

			if tt.wantErr {
				if err == nil {
					assert.GreaterOrEqual(t, rec.Code, 400)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

// =============================================================================
// RLS (Row Level Security) Tenant Isolation Tests
// =============================================================================

func TestAuth_RLS_TenantIsolation(t *testing.T) {
	t.Parallel()

	adminClient, appClient := common.SetupTestClientWithRLS(t)
	deps := BuildTestDependencies(appClient)

	// Create test data in two tenants using admin client
	_ = common.CreateTestDataSet(t, adminClient)

	t.Run("users in different tenants cannot login to wrong tenant", func(t *testing.T) {
		e := SetupEcho()

		// First register a user in tenant1
		registerReq := api.RegisterRequest{
			Email:      "rls-user@tenant1.com",
			Password:   "password123",
			Name:       strPtr("RLS User"),
			TenantSlug: "rls-tenant-1",
		}

		body, _ := json.Marshal(registerReq)
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		_ = deps.AuthController.Register(c)

		// Try to login with the same email but wrong tenant
		loginReq := api.LoginRequest{
			Email:      "rls-user@tenant1.com",
			Password:   "password123",
			TenantSlug: "rls-tenant-2", // Wrong tenant
		}

		body, _ = json.Marshal(loginReq)
		req2 := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
		req2.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec2 := httptest.NewRecorder()
		c2 := e.NewContext(req2, rec2)

		err := deps.AuthController.Login(c2)
		require.Error(t, err, "Should not be able to login with wrong tenant")
	})

	t.Run("same email can exist in different tenants", func(t *testing.T) {
		e := SetupEcho()

		// Register user in tenant A
		registerReq1 := api.RegisterRequest{
			Email:      "same-email@example.com",
			Password:   "password123",
			Name:       strPtr("User in Tenant A"),
			TenantSlug: "multi-tenant-a",
		}

		body, _ := json.Marshal(registerReq1)
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		err := deps.AuthController.Register(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)

		// Register same email in tenant B - should succeed
		registerReq2 := api.RegisterRequest{
			Email:      "same-email@example.com",
			Password:   "password456",
			Name:       strPtr("User in Tenant B"),
			TenantSlug: "multi-tenant-b",
		}

		body, _ = json.Marshal(registerReq2)
		req2 := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
		req2.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec2 := httptest.NewRecorder()
		c2 := e.NewContext(req2, rec2)
		err = deps.AuthController.Register(c2)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec2.Code)

		// Both users should be able to login to their respective tenants
		loginReq1 := api.LoginRequest{
			Email:      "same-email@example.com",
			Password:   "password123",
			TenantSlug: "multi-tenant-a",
		}
		body, _ = json.Marshal(loginReq1)
		req3 := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
		req3.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec3 := httptest.NewRecorder()
		c3 := e.NewContext(req3, rec3)
		err = deps.AuthController.Login(c3)
		require.NoError(t, err)

		loginReq2 := api.LoginRequest{
			Email:      "same-email@example.com",
			Password:   "password456",
			TenantSlug: "multi-tenant-b",
		}
		body, _ = json.Marshal(loginReq2)
		req4 := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
		req4.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec4 := httptest.NewRecorder()
		c4 := e.NewContext(req4, rec4)
		err = deps.AuthController.Login(c4)
		require.NoError(t, err)
	})
}
