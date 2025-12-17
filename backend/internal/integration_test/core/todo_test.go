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

func TestTodo_Create(t *testing.T) {
	t.Parallel()

	client := common.SetupTestClient(t)
	dataSet := common.CreateTestDataSet(t, client)
	deps := BuildTestDependencies(client)

	tests := []struct {
		name           string
		userID         string
		tenantID       string
		requestBody    api.CreateTodoRequest
		expectedStatus int
		expectedTitle  string
		wantErr        bool
	}{
		{
			name:     "success - create private todo",
			userID:   dataSet.User1.ID,
			tenantID: dataSet.Tenant1.ID,
			requestBody: api.CreateTodoRequest{
				Title:       "Test Todo",
				Description: strPtr("Test Description"),
				IsPublic:    boolPtr(false),
			},
			expectedStatus: http.StatusCreated,
			expectedTitle:  "Test Todo",
			wantErr:        false,
		},
		{
			name:     "success - create public todo",
			userID:   dataSet.User1.ID,
			tenantID: dataSet.Tenant1.ID,
			requestBody: api.CreateTodoRequest{
				Title:       "Public Todo",
				Description: strPtr("Public Description"),
				IsPublic:    boolPtr(true),
			},
			expectedStatus: http.StatusCreated,
			expectedTitle:  "Public Todo",
			wantErr:        false,
		},
		{
			name:     "fail - empty title",
			userID:   dataSet.User1.ID,
			tenantID: dataSet.Tenant1.ID,
			requestBody: api.CreateTodoRequest{
				Title: "",
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

			req := httptest.NewRequest(http.MethodPost, "/todos", bytes.NewReader(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)
			SetAuthContext(c, tt.userID, tt.tenantID)

			err = deps.TodoController.CreateTodo(c)

			if tt.wantErr {
				if err == nil {
					assert.GreaterOrEqual(t, rec.Code, 400)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			var response api.TodoResponse
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.NotEmpty(t, response.Id)
			assert.Equal(t, tt.expectedTitle, *response.Title)
		})
	}
}

func TestTodo_GetTodos(t *testing.T) {
	t.Parallel()

	client := common.SetupTestClient(t)
	dataSet := common.CreateTestDataSet(t, client)
	deps := BuildTestDependencies(client)

	tests := []struct {
		name           string
		userID         string
		tenantID       string
		limit          *int
		offset         *int
		expectedStatus int
		expectedTotal  int
	}{
		{
			name:           "success - get User1 todos",
			userID:         dataSet.User1.ID,
			tenantID:       dataSet.Tenant1.ID,
			expectedStatus: http.StatusOK,
			expectedTotal:  2, // Todo1 and Todo2
		},
		{
			name:           "success - get User2 todos",
			userID:         dataSet.User2.ID,
			tenantID:       dataSet.Tenant1.ID,
			expectedStatus: http.StatusOK,
			expectedTotal:  1, // Todo3
		},
		{
			name:           "success - get User3 todos (different tenant)",
			userID:         dataSet.User3.ID,
			tenantID:       dataSet.Tenant2.ID,
			expectedStatus: http.StatusOK,
			expectedTotal:  2, // Todo4 and Todo5
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := SetupEcho()

			req := httptest.NewRequest(http.MethodGet, "/todos", nil)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)
			SetAuthContext(c, tt.userID, tt.tenantID)

			params := api.GetTodosParams{
				Limit:  tt.limit,
				Offset: tt.offset,
			}

			err := deps.TodoController.GetTodos(c, params)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			var response api.TodoListResponse
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedTotal, *response.Total)
		})
	}
}

func TestTodo_GetPublicTodos(t *testing.T) {
	t.Parallel()

	client := common.SetupTestClient(t)
	dataSet := common.CreateTestDataSet(t, client)
	deps := BuildTestDependencies(client)

	tests := []struct {
		name           string
		userID         string
		tenantID       string
		expectedStatus int
		expectedTotal  int
		expectedTodoID string
	}{
		{
			name:           "success - get public todos in Tenant1",
			userID:         dataSet.User1.ID,
			tenantID:       dataSet.Tenant1.ID,
			expectedStatus: http.StatusOK,
			expectedTotal:  1,
			expectedTodoID: dataSet.Todo2.ID, // only public todo in Tenant1
		},
		{
			name:           "success - get public todos in Tenant2",
			userID:         dataSet.User3.ID,
			tenantID:       dataSet.Tenant2.ID,
			expectedStatus: http.StatusOK,
			expectedTotal:  1,
			expectedTodoID: dataSet.Todo5.ID, // only public todo in Tenant2
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := SetupEcho()

			req := httptest.NewRequest(http.MethodGet, "/todos/public", nil)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)
			SetAuthContext(c, tt.userID, tt.tenantID)

			params := api.GetPublicTodosParams{}

			err := deps.TodoController.GetPublicTodos(c, params)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			var response api.TodoListResponse
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedTotal, *response.Total)
			if tt.expectedTotal > 0 {
				assert.Equal(t, tt.expectedTodoID, *(*response.Todos)[0].Id)
			}
		})
	}
}

func TestTodo_GetTodo(t *testing.T) {
	t.Parallel()

	client := common.SetupTestClient(t)
	dataSet := common.CreateTestDataSet(t, client)
	deps := BuildTestDependencies(client)

	tests := []struct {
		name           string
		todoID         string
		userID         string
		tenantID       string
		expectedStatus int
		wantErr        bool
		errContains    string
	}{
		{
			name:           "success - owner can access own private todo",
			todoID:         dataSet.Todo1.ID,
			userID:         dataSet.User1.ID,
			tenantID:       dataSet.Tenant1.ID,
			expectedStatus: http.StatusOK,
			wantErr:        false,
		},
		{
			name:           "success - user can access public todo",
			todoID:         dataSet.Todo2.ID,
			userID:         dataSet.User2.ID,
			tenantID:       dataSet.Tenant1.ID,
			expectedStatus: http.StatusOK,
			wantErr:        false,
		},
		{
			name:        "fail - user cannot access other's private todo",
			todoID:      dataSet.Todo1.ID,
			userID:      dataSet.User2.ID,
			tenantID:    dataSet.Tenant1.ID,
			wantErr:     true,
			errContains: "not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := SetupEcho()

			req := httptest.NewRequest(http.MethodGet, "/todos/"+tt.todoID, nil)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)
			c.SetParamNames("todoId")
			c.SetParamValues(tt.todoID)
			SetAuthContext(c, tt.userID, tt.tenantID)

			err := deps.TodoController.GetTodo(c, tt.todoID)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestTodo_Update(t *testing.T) {
	t.Parallel()

	client := common.SetupTestClient(t)
	dataSet := common.CreateTestDataSet(t, client)
	deps := BuildTestDependencies(client)

	tests := []struct {
		name           string
		todoID         string
		userID         string
		tenantID       string
		requestBody    api.UpdateTodoRequest
		expectedStatus int
		wantErr        bool
		errContains    string
	}{
		{
			name:     "success - owner can update own todo",
			todoID:   dataSet.Todo1.ID,
			userID:   dataSet.User1.ID,
			tenantID: dataSet.Tenant1.ID,
			requestBody: api.UpdateTodoRequest{
				Title:     strPtr("Updated Title"),
				Completed: boolPtr(true),
			},
			expectedStatus: http.StatusOK,
			wantErr:        false,
		},
		{
			name:     "fail - user cannot update other's todo",
			todoID:   dataSet.Todo1.ID,
			userID:   dataSet.User2.ID,
			tenantID: dataSet.Tenant1.ID,
			requestBody: api.UpdateTodoRequest{
				Title: strPtr("Hacked Title"),
			},
			wantErr:     true,
			errContains: "not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := SetupEcho()

			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPatch, "/todos/"+tt.todoID, bytes.NewReader(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)
			c.SetParamNames("todoId")
			c.SetParamValues(tt.todoID)
			SetAuthContext(c, tt.userID, tt.tenantID)

			err = deps.TodoController.UpdateTodo(c, tt.todoID)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			var response api.TodoResponse
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			require.NoError(t, err)

			if tt.requestBody.Title != nil {
				assert.Equal(t, *tt.requestBody.Title, *response.Title)
			}
		})
	}
}

func TestTodo_Delete(t *testing.T) {
	t.Parallel()

	client := common.SetupTestClient(t)
	dataSet := common.CreateTestDataSet(t, client)
	deps := BuildTestDependencies(client)

	tests := []struct {
		name           string
		todoID         string
		userID         string
		tenantID       string
		expectedStatus int
		wantErr        bool
		errContains    string
	}{
		{
			name:           "success - owner can delete own todo",
			todoID:         dataSet.Todo1.ID,
			userID:         dataSet.User1.ID,
			tenantID:       dataSet.Tenant1.ID,
			expectedStatus: http.StatusNoContent,
			wantErr:        false,
		},
		{
			name:        "fail - user cannot delete other's todo",
			todoID:      dataSet.Todo3.ID,
			userID:      dataSet.User1.ID,
			tenantID:    dataSet.Tenant1.ID,
			wantErr:     true,
			errContains: "not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := SetupEcho()

			req := httptest.NewRequest(http.MethodDelete, "/todos/"+tt.todoID, nil)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)
			c.SetParamNames("todoId")
			c.SetParamValues(tt.todoID)
			SetAuthContext(c, tt.userID, tt.tenantID)

			err := deps.TodoController.DeleteTodo(c, tt.todoID)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
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

func TestTodo_RLS_TenantIsolation(t *testing.T) {
	t.Parallel()

	adminClient, appClient := common.SetupTestClientWithRLS(t)
	dataSet := common.CreateTestDataSet(t, adminClient)
	deps := BuildTestDependencies(appClient)

	tests := []struct {
		name          string
		userID        string
		tenantID      string
		expectedTotal int
		description   string
	}{
		{
			name:          "Tenant1 user sees only Tenant1 public todos",
			userID:        dataSet.User1.ID,
			tenantID:      dataSet.Tenant1.ID,
			expectedTotal: 1,
			description:   "RLS should filter out Tenant2's todos",
		},
		{
			name:          "Tenant2 user sees only Tenant2 public todos",
			userID:        dataSet.User3.ID,
			tenantID:      dataSet.Tenant2.ID,
			expectedTotal: 1,
			description:   "RLS should filter out Tenant1's todos",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := SetupEcho()

			req := httptest.NewRequest(http.MethodGet, "/todos/public", nil)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)
			SetAuthContext(c, tt.userID, tt.tenantID)

			params := api.GetPublicTodosParams{}

			err := deps.TodoController.GetPublicTodos(c, params)
			require.NoError(t, err)

			var response api.TodoListResponse
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedTotal, *response.Total, tt.description)
		})
	}
}

func TestTodo_RLS_CrossTenantAccess(t *testing.T) {
	t.Parallel()

	adminClient, appClient := common.SetupTestClientWithRLS(t)
	dataSet := common.CreateTestDataSet(t, adminClient)
	deps := BuildTestDependencies(appClient)

	tests := []struct {
		name        string
		todoID      string
		userID      string
		tenantID    string
		description string
	}{
		{
			name:        "Tenant1 cannot access Tenant2's todo",
			todoID:      dataSet.Todo4.ID,
			userID:      dataSet.User1.ID,
			tenantID:    dataSet.Tenant1.ID,
			description: "RLS should block cross-tenant read",
		},
		{
			name:        "Tenant2 cannot access Tenant1's todo",
			todoID:      dataSet.Todo1.ID,
			userID:      dataSet.User3.ID,
			tenantID:    dataSet.Tenant2.ID,
			description: "RLS should block cross-tenant read",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := SetupEcho()

			req := httptest.NewRequest(http.MethodGet, "/todos/"+tt.todoID, nil)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)
			c.SetParamNames("todoId")
			c.SetParamValues(tt.todoID)
			SetAuthContext(c, tt.userID, tt.tenantID)

			err := deps.TodoController.GetTodo(c, tt.todoID)
			require.Error(t, err, tt.description)
		})
	}
}

func TestTodo_RLS_CrossTenantModification(t *testing.T) {
	t.Parallel()

	adminClient, appClient := common.SetupTestClientWithRLS(t)
	dataSet := common.CreateTestDataSet(t, adminClient)
	deps := BuildTestDependencies(appClient)

	tests := []struct {
		name        string
		operation   string
		todoID      string
		userID      string
		tenantID    string
		description string
	}{
		{
			name:        "Tenant2 cannot update Tenant1's todo",
			operation:   "update",
			todoID:      dataSet.Todo1.ID,
			userID:      dataSet.User3.ID,
			tenantID:    dataSet.Tenant2.ID,
			description: "RLS should block cross-tenant update",
		},
		{
			name:        "Tenant2 cannot delete Tenant1's todo",
			operation:   "delete",
			todoID:      dataSet.Todo1.ID,
			userID:      dataSet.User3.ID,
			tenantID:    dataSet.Tenant2.ID,
			description: "RLS should block cross-tenant delete",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := SetupEcho()

			var req *http.Request
			if tt.operation == "update" {
				body, _ := json.Marshal(api.UpdateTodoRequest{Title: strPtr("Hacked")})
				req = httptest.NewRequest(http.MethodPatch, "/todos/"+tt.todoID, bytes.NewReader(body))
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			} else {
				req = httptest.NewRequest(http.MethodDelete, "/todos/"+tt.todoID, nil)
			}

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("todoId")
			c.SetParamValues(tt.todoID)
			SetAuthContext(c, tt.userID, tt.tenantID)

			var err error
			if tt.operation == "update" {
				err = deps.TodoController.UpdateTodo(c, tt.todoID)
			} else {
				err = deps.TodoController.DeleteTodo(c, tt.todoID)
			}

			require.Error(t, err, tt.description)
		})
	}
}

// Helper functions
func strPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}
