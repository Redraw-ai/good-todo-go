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

func TestUser_GetMe(t *testing.T) {
	t.Parallel()

	client := common.SetupTestClient(t)
	dataSet := common.CreateTestDataSet(t, client)
	deps := BuildTestDependencies(client)

	tests := []struct {
		name           string
		userID         string
		tenantID       string
		expectedStatus int
		expectedEmail  string
		expectedName   string
		wantErr        bool
	}{
		{
			name:           "success - get User1",
			userID:         dataSet.User1.ID,
			tenantID:       dataSet.Tenant1.ID,
			expectedStatus: http.StatusOK,
			expectedEmail:  dataSet.User1.Email,
			expectedName:   dataSet.User1.Name,
			wantErr:        false,
		},
		{
			name:           "success - get User2",
			userID:         dataSet.User2.ID,
			tenantID:       dataSet.Tenant1.ID,
			expectedStatus: http.StatusOK,
			expectedEmail:  dataSet.User2.Email,
			expectedName:   dataSet.User2.Name,
			wantErr:        false,
		},
		{
			name:     "fail - non-existent user",
			userID:   "non-existent-user-id",
			tenantID: dataSet.Tenant1.ID,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := SetupEcho()

			req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)
			SetAuthContext(c, tt.userID, tt.tenantID)

			err := deps.UserController.GetMe(c)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			var response api.UserResponse
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, tt.userID, *response.Id)
			assert.Equal(t, tt.expectedEmail, *response.Email)
			assert.Equal(t, tt.expectedName, *response.Name)
		})
	}
}

func TestUser_UpdateMe(t *testing.T) {
	t.Parallel()

	client := common.SetupTestClient(t)
	dataSet := common.CreateTestDataSet(t, client)
	deps := BuildTestDependencies(client)

	tests := []struct {
		name           string
		userID         string
		tenantID       string
		requestBody    api.UpdateUserRequest
		expectedStatus int
		expectedName   string
		wantErr        bool
	}{
		{
			name:     "success - update name",
			userID:   dataSet.User1.ID,
			tenantID: dataSet.Tenant1.ID,
			requestBody: api.UpdateUserRequest{
				Name: strPtr("Updated Name"),
			},
			expectedStatus: http.StatusOK,
			expectedName:   "Updated Name",
			wantErr:        false,
		},
		{
			name:     "fail - non-existent user",
			userID:   "non-existent-user-id",
			tenantID: dataSet.Tenant1.ID,
			requestBody: api.UpdateUserRequest{
				Name: strPtr("New Name"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := SetupEcho()

			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPatch, "/users/me", bytes.NewReader(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)
			SetAuthContext(c, tt.userID, tt.tenantID)

			err = deps.UserController.UpdateMe(c)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			var response api.UserResponse
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedName, *response.Name)
		})
	}
}

// =============================================================================
// RLS (Row Level Security) Tenant Isolation Tests
// =============================================================================

func TestUser_RLS_TenantIsolation(t *testing.T) {
	t.Parallel()

	adminClient, appClient := common.SetupTestClientWithRLS(t)
	dataSet := common.CreateTestDataSet(t, adminClient)
	deps := BuildTestDependencies(appClient)

	tests := []struct {
		name        string
		userID      string
		tenantID    string
		wantErr     bool
		description string
	}{
		{
			name:        "success - Tenant1 user can access own info",
			userID:      dataSet.User1.ID,
			tenantID:    dataSet.Tenant1.ID,
			wantErr:     false,
			description: "User should access their own info",
		},
		{
			name:        "fail - Tenant1 cannot access Tenant2 user",
			userID:      dataSet.User3.ID,
			tenantID:    dataSet.Tenant1.ID,
			wantErr:     true,
			description: "RLS should block cross-tenant user access",
		},
		{
			name:        "fail - Tenant2 cannot access Tenant1 user",
			userID:      dataSet.User1.ID,
			tenantID:    dataSet.Tenant2.ID,
			wantErr:     true,
			description: "RLS should block cross-tenant user access",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := SetupEcho()

			req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)
			SetAuthContext(c, tt.userID, tt.tenantID)

			err := deps.UserController.GetMe(c)

			if tt.wantErr {
				require.Error(t, err, tt.description)
				return
			}

			require.NoError(t, err, tt.description)
			assert.Equal(t, http.StatusOK, rec.Code)
		})
	}
}

func TestUser_RLS_UpdateIsolation(t *testing.T) {
	t.Parallel()

	adminClient, appClient := common.SetupTestClientWithRLS(t)
	dataSet := common.CreateTestDataSet(t, adminClient)
	deps := BuildTestDependencies(appClient)

	tests := []struct {
		name        string
		userID      string
		tenantID    string
		requestBody api.UpdateUserRequest
		wantErr     bool
		description string
	}{
		{
			name:     "fail - Tenant2 cannot update Tenant1 user",
			userID:   dataSet.User1.ID,
			tenantID: dataSet.Tenant2.ID,
			requestBody: api.UpdateUserRequest{
				Name: strPtr("Hacked Name"),
			},
			wantErr:     true,
			description: "RLS should block cross-tenant user update",
		},
		{
			name:     "fail - Tenant1 cannot update Tenant2 user",
			userID:   dataSet.User3.ID,
			tenantID: dataSet.Tenant1.ID,
			requestBody: api.UpdateUserRequest{
				Name: strPtr("Hacked Name"),
			},
			wantErr:     true,
			description: "RLS should block cross-tenant user update",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := SetupEcho()

			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPatch, "/users/me", bytes.NewReader(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)
			SetAuthContext(c, tt.userID, tt.tenantID)

			err = deps.UserController.UpdateMe(c)

			if tt.wantErr {
				require.Error(t, err, tt.description)
				return
			}

			require.NoError(t, err, tt.description)
		})
	}
}
