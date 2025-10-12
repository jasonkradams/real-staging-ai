package project

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/real-staging-ai/api/internal/storage"
)

// ---------------------- Validation tests ----------------------

// Helper to test project name validation
func testProjectNameValidation(t *testing.T, validateFunc func(string) []ValidationErrorDetail) {
	cases := []struct {
		name        string
		projectName string
		expectError bool
		field       string
		contains    string
	}{
		{
			name:        "success: valid",
			projectName: "My Project",
			expectError: false,
		},
		{
			name:        "fail: empty name",
			projectName: "",
			expectError: true,
			field:       "name",
			contains:    "required",
		},
		{
			name:        "fail: name too long",
			projectName: strings.Repeat("a", 101),
			expectError: true,
			field:       "name",
			contains:    "100",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			errs := validateFunc(tc.projectName)
			if tc.expectError {
				assert.NotEmpty(t, errs)
				assert.Equal(t, tc.field, errs[0].Field)
				if tc.contains != "" {
					assert.Contains(t, errs[0].Message, tc.contains)
				}
			} else {
				assert.Empty(t, errs)
			}
		})
	}
}

func TestValidateCreateProjectRequest(t *testing.T) {
	testProjectNameValidation(t, func(name string) []ValidationErrorDetail {
		return validateCreateProjectRequest(&CreateRequest{Name: name})
	})
}

func TestValidateUpdateProjectRequest(t *testing.T) {
	testProjectNameValidation(t, func(name string) []ValidationErrorDetail {
		return validateUpdateProjectRequest(&UpdateRequest{Name: name})
	})
}

// ---------------------- Handler tests ----------------------

func TestDefaultHandler_Create(t *testing.T) {
	cases := []struct {
		name           string
		body           string
		wantStatusCode int
		contains       string
		setupDB        func() *storage.DatabaseMock
		setHeaders     func(*http.Request)
	}{
		{
			name:           "fail: bad request - invalid json",
			body:           `{"name": "missing-quote}`,
			wantStatusCode: http.StatusBadRequest,
			contains:       "Invalid request format",
		},
		{
			name:           "fail: validation error - empty name",
			body:           `{"name": ""}`,
			wantStatusCode: http.StatusUnprocessableEntity,
			contains:       "validation_failed",
		},
		{
			name:           "success: create project",
			body:           `{"name":"My Project"}`,
			wantStatusCode: http.StatusCreated,
			setupDB:        newDBMockForCreateProjectSuccess,
			setHeaders: func(r *http.Request) {
				r.Header.Set("X-Test-User", "auth0|testuser")
			},
		},
		{
			name:           "success: create project (user auto-created)",
			body:           `{"name":"My Project"}`,
			wantStatusCode: http.StatusCreated,
			setupDB:        newDBMockForCreateProject_UserNotFoundThenCreate,
			setHeaders: func(r *http.Request) {
				r.Header.Set("X-Test-User", "auth0|testuser")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewBufferString(tc.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			if tc.setHeaders != nil {
				tc.setHeaders(req)
			}
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			var h *DefaultHandler
			if tc.setupDB != nil {
				h = NewDefaultHandler(tc.setupDB())
			} else {
				h = NewDefaultHandler(nil)
			}

			err := h.Create(c)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantStatusCode, rec.Code)
			if tc.contains != "" {
				assert.Contains(t, rec.Body.String(), tc.contains)
			}
		})
	}
}

func TestDefaultHandler_GetByID(t *testing.T) {
	cases := []struct {
		name           string
		projectID      string
		wantStatusCode int
		contains       string
		setupDB        func() *storage.DatabaseMock
		setHeaders     func(*http.Request)
	}{
		{
			name:           "fail: bad request - invalid uuid",
			projectID:      "invalid-uuid",
			wantStatusCode: http.StatusBadRequest,
			contains:       "Invalid project ID format",
		},
		{
			name:           "success: get project by id",
			projectID:      uuid.New().String(),
			wantStatusCode: http.StatusOK,
			setupDB:        newDBMockForGetProjectByIDSuccess,
			setHeaders: func(r *http.Request) {
				r.Header.Set("X-Test-User", "auth0|testuser")
			},
		},
		{
			name:           "fail: project not found",
			projectID:      uuid.New().String(),
			wantStatusCode: http.StatusNotFound,
			contains:       "Project not found",
			setupDB:        newDBMockForGetProjectByID_NotFound,
			setHeaders: func(r *http.Request) {
				r.Header.Set("X-Test-User", "auth0|testuser")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/"+tc.projectID, nil)
			if tc.setHeaders != nil {
				tc.setHeaders(req)
			}
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tc.projectID)

			var h *DefaultHandler
			if tc.setupDB != nil {
				h = NewDefaultHandler(tc.setupDB())
			} else {
				h = NewDefaultHandler(nil)
			}

			err := h.GetByID(c)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantStatusCode, rec.Code)
			if tc.contains != "" {
				assert.Contains(t, rec.Body.String(), tc.contains)
			}
		})
	}
}

func TestDefaultHandler_List(t *testing.T) {
	cases := []struct {
		name           string
		wantStatusCode int
		contains       string
	}{
		{
			name:           "fail: unauthorized when user not resolved",
			wantStatusCode: http.StatusUnauthorized,
			contains:       "Invalid or missing JWT token",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			h := NewDefaultHandler(nil)
			err := h.List(c)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantStatusCode, rec.Code)
			if tc.contains != "" {
				assert.Contains(t, rec.Body.String(), tc.contains)
			}
		})
	}
}

func TestDefaultHandler_Update(t *testing.T) {
	cases := []struct {
		name           string
		projectID      string
		body           string
		wantStatusCode int
		contains       string
	}{
		{
			name:           "fail: bad request - invalid uuid",
			projectID:      "invalid-uuid",
			body:           `{"name":"ok"}`,
			wantStatusCode: http.StatusBadRequest,
			contains:       "Invalid project ID format",
		},
		{
			name:           "fail: validation error - empty name",
			projectID:      uuid.New().String(),
			body:           `{"name": ""}`,
			wantStatusCode: http.StatusUnprocessableEntity,
			contains:       "validation_failed",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPut, "/api/v1/projects/"+tc.projectID, bytes.NewBufferString(tc.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tc.projectID)

			h := NewDefaultHandler(nil)
			err := h.Update(c)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantStatusCode, rec.Code)
			if tc.contains != "" {
				assert.Contains(t, rec.Body.String(), tc.contains)
			}
		})
	}
}

func TestDefaultHandler_Delete(t *testing.T) {
	cases := []struct {
		name           string
		projectID      string
		wantStatusCode int
		contains       string
	}{
		{
			name:           "fail: bad request - invalid uuid",
			projectID:      "invalid-uuid",
			wantStatusCode: http.StatusBadRequest,
			contains:       "Invalid project ID format",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodDelete, "/api/v1/projects/"+tc.projectID, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tc.projectID)

			h := NewDefaultHandler(nil)
			err := h.Delete(c)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantStatusCode, rec.Code)
			if tc.contains != "" {
				assert.Contains(t, rec.Body.String(), tc.contains)
			}
		})
	}
}

// ---------------------- DB Mock helpers ----------------------

type fakeRow struct {
	scan func(dest ...any) error
}

func (r fakeRow) Scan(dest ...any) error {
	return r.scan(dest...)
}

// success path for Create: user exists, insert project ok
func newDBMockForCreateProjectSuccess() *storage.DatabaseMock {
	now := time.Now()
	userID := uuid.New()

	return &storage.DatabaseMock{
		QueryRowFunc: func(ctx context.Context, sql string, args ...interface{}) pgx.Row {
			switch {
			// SELECT user by auth0_sub
			case strings.Contains(sql, "FROM users") && strings.Contains(sql, "auth0_sub"):
				return fakeRow{scan: func(dest ...any) error {
					// id, auth0_sub, stripe_customer_id, role, created_at
					if u, ok := dest[0].(*pgtype.UUID); ok {
						u.Bytes = userID
						u.Valid = true
					}
					if s, ok := dest[1].(*string); ok && len(args) > 0 {
						*s = args[0].(string)
					}
					if role, ok := dest[3].(*string); ok {
						*role = "user"
					}
					if ts, ok := dest[4].(*pgtype.Timestamptz); ok {
						ts.Time = now
						ts.Valid = true
					}
					return nil
				}}
			// INSERT project
			case strings.Contains(sql, "INSERT INTO projects"):
				return fakeRow{scan: func(dest ...any) error {
					// id, user_id, created_at
					if id, ok := dest[0].(*string); ok {
						*id = uuid.New().String()
					}
					if uid, ok := dest[1].(*string); ok && len(args) > 1 {
						*uid = args[1].(string)
					}
					if ts, ok := dest[2].(*time.Time); ok {
						*ts = now
					}
					return nil
				}}
			default:
				return fakeRow{scan: func(dest ...any) error { return nil }}
			}
		},
	}
}

// user not found -> create user -> insert project
func newDBMockForCreateProject_UserNotFoundThenCreate() *storage.DatabaseMock {
	now := time.Now()
	newUserID := uuid.New()

	return &storage.DatabaseMock{
		QueryRowFunc: func(ctx context.Context, sql string, args ...interface{}) pgx.Row {
			switch {
			// Get user by auth0_sub -> not found
			case strings.Contains(sql, "FROM users") && strings.Contains(sql, "auth0_sub") && strings.Contains(sql, "WHERE"):
				return fakeRow{scan: func(dest ...any) error {
					return pgx.ErrNoRows
				}}
			// Create user
			case strings.Contains(sql, "INSERT INTO users"):
				return fakeRow{scan: func(dest ...any) error {
					// id, auth0_sub, stripe_customer_id, role, created_at
					if u, ok := dest[0].(*pgtype.UUID); ok {
						u.Bytes = newUserID
						u.Valid = true
					}
					if s, ok := dest[1].(*string); ok && len(args) > 0 {
						*s = args[0].(string)
					}
					if role, ok := dest[3].(*string); ok {
						*role = "user"
					}
					if ts, ok := dest[4].(*pgtype.Timestamptz); ok {
						ts.Time = now
						ts.Valid = true
					}
					return nil
				}}
			// Insert project
			case strings.Contains(sql, "INSERT INTO projects"):
				return fakeRow{scan: func(dest ...any) error {
					if id, ok := dest[0].(*string); ok {
						*id = uuid.New().String()
					}
					if uid, ok := dest[1].(*string); ok {
						*uid = newUserID.String()
					}
					if ts, ok := dest[2].(*time.Time); ok {
						*ts = now
					}
					return nil
				}}
			default:
				return fakeRow{scan: func(dest ...any) error { return nil }}
			}
		},
	}
}

// success path for GetByID: user exists, project found
func newDBMockForGetProjectByIDSuccess() *storage.DatabaseMock {
	now := time.Now()
	userID := uuid.New()

	return &storage.DatabaseMock{
		QueryRowFunc: func(ctx context.Context, sql string, args ...interface{}) pgx.Row {
			switch {
			// Resolve user
			case strings.Contains(sql, "FROM users") && strings.Contains(sql, "auth0_sub") && strings.Contains(sql, "WHERE"):
				return fakeRow{scan: func(dest ...any) error {
					if u, ok := dest[0].(*pgtype.UUID); ok {
						u.Bytes = userID
						u.Valid = true
					}
					if s, ok := dest[1].(*string); ok && len(args) > 0 {
						*s = args[0].(string)
					}
					if ts, ok := dest[4].(*pgtype.Timestamptz); ok {
						ts.Time = now
						ts.Valid = true
					}
					return nil
				}}
			// Project found
			case strings.Contains(sql, "FROM projects") && strings.Contains(sql, "WHERE"):
				return fakeRow{scan: func(dest ...any) error {
					if id, ok := dest[0].(*string); ok && len(args) > 0 {
						*id = args[0].(string)
					}
					if name, ok := dest[1].(*string); ok {
						*name = "My Project"
					}
					if uid, ok := dest[2].(*string); ok {
						*uid = userID.String()
					}
					if ts, ok := dest[3].(*time.Time); ok {
						*ts = now
					}
					return nil
				}}
			default:
				return fakeRow{scan: func(dest ...any) error { return nil }}
			}
		},
	}
}

// not found path for GetByID: user exists, project missing
func newDBMockForGetProjectByID_NotFound() *storage.DatabaseMock {
	now := time.Now()
	userID := uuid.New()

	return &storage.DatabaseMock{
		QueryRowFunc: func(ctx context.Context, sql string, args ...interface{}) pgx.Row {
			switch {
			// Resolve user
			case strings.Contains(sql, "FROM users") && strings.Contains(sql, "auth0_sub") && strings.Contains(sql, "WHERE"):
				return fakeRow{scan: func(dest ...any) error {
					if u, ok := dest[0].(*pgtype.UUID); ok {
						u.Bytes = userID
						u.Valid = true
					}
					if s, ok := dest[1].(*string); ok && len(args) > 0 {
						*s = args[0].(string)
					}
					if ts, ok := dest[4].(*pgtype.Timestamptz); ok {
						ts.Time = now
						ts.Valid = true
					}
					return nil
				}}
			// Project lookup returns not found
			case strings.Contains(sql, "FROM projects") && strings.Contains(sql, "WHERE"):
				return fakeRow{scan: func(dest ...any) error {
					return pgx.ErrNoRows
				}}
			default:
				return fakeRow{scan: func(dest ...any) error { return nil }}
			}
		},
	}
}
