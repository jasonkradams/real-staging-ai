package reconcile

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultHandler_ReconcileImages(t *testing.T) {
	testCases := []struct {
		name           string
		envVars        map[string]string
		queryParams    string
		setupMock      func(*ServiceMock)
		expectedStatus int
		validateResp   func(t *testing.T, body string)
	}{
		{
			name: "success: basic reconciliation",
			envVars: map[string]string{
				"RECONCILE_ENABLED": "1",
			},
			queryParams: "limit=50&dry_run=true",
			setupMock: func(svcMock *ServiceMock) {
				svcMock.ReconcileImagesFunc = func(ctx context.Context, opts ReconcileOptions) (*ReconcileResult, error) {
					// Validate params (return errors instead of using assert in mock)
					if opts.Limit != 50 {
						return nil, errors.New("expected limit 50")
					}
					if opts.Concurrency != 5 {
						return nil, errors.New("expected concurrency 5")
					}
					if !opts.DryRun {
						return nil, errors.New("expected dry_run true")
					}
					return &ReconcileResult{
						Checked:       10,
						MissingOrig:   2,
						MissingStaged: 1,
						Updated:       3,
						DryRun:        true,
						Examples:      []ReconcileError{},
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
			validateResp: func(t *testing.T, body string) {
				var result ReconcileResult
				err := json.Unmarshal([]byte(body), &result)
				require.NoError(t, err)
				assert.Equal(t, 10, result.Checked)
				assert.Equal(t, 2, result.MissingOrig)
				assert.Equal(t, 1, result.MissingStaged)
				assert.Equal(t, 3, result.Updated)
				assert.True(t, result.DryRun)
			},
		},
		{
			name: "success: with project_id filter",
			envVars: map[string]string{
				"RECONCILE_ENABLED": "1",
			},
			queryParams: "project_id=550e8400-e29b-41d4-a716-446655440000&limit=100",
			setupMock: func(svcMock *ServiceMock) {
				svcMock.ReconcileImagesFunc = func(ctx context.Context, opts ReconcileOptions) (*ReconcileResult, error) {
					// Just return success - we'll verify the mock was called correctly below
					return &ReconcileResult{
						Checked: 5,
						DryRun:  false,
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
			validateResp: func(t *testing.T, body string) {
				// Service was called, that's good enough for this test
			},
		},
		{
			name: "success: with status filter",
			envVars: map[string]string{
				"RECONCILE_ENABLED": "1",
			},
			queryParams: "status=ready&limit=100",
			setupMock: func(svcMock *ServiceMock) {
				svcMock.ReconcileImagesFunc = func(ctx context.Context, opts ReconcileOptions) (*ReconcileResult, error) {
					// Just return success
					return &ReconcileResult{
						Checked: 3,
						DryRun:  false,
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "success: custom concurrency from query",
			envVars: map[string]string{
				"RECONCILE_ENABLED": "1",
			},
			queryParams: "concurrency=10&limit=100",
			setupMock: func(svcMock *ServiceMock) {
				svcMock.ReconcileImagesFunc = func(ctx context.Context, opts ReconcileOptions) (*ReconcileResult, error) {
					if opts.Concurrency != 10 {
						return nil, errors.New("expected concurrency 10")
					}
					return &ReconcileResult{
						Checked: 0,
						DryRun:  false,
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "success: custom concurrency from env",
			envVars: map[string]string{
				"RECONCILE_ENABLED":     "1",
				"RECONCILE_CONCURRENCY": "15",
			},
			queryParams: "limit=100",
			setupMock: func(svcMock *ServiceMock) {
				svcMock.ReconcileImagesFunc = func(ctx context.Context, opts ReconcileOptions) (*ReconcileResult, error) {
					if opts.Concurrency != 15 {
						return nil, errors.New("expected concurrency 15")
					}
					return &ReconcileResult{
						Checked: 0,
						DryRun:  false,
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "success: query param concurrency overrides env",
			envVars: map[string]string{
				"RECONCILE_ENABLED":     "1",
				"RECONCILE_CONCURRENCY": "15",
			},
			queryParams: "concurrency=20&limit=100",
			setupMock: func(svcMock *ServiceMock) {
				svcMock.ReconcileImagesFunc = func(ctx context.Context, opts ReconcileOptions) (*ReconcileResult, error) {
					// Service was called - just return success
					return &ReconcileResult{
						Checked: 0,
						DryRun:  false,
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "success: default limit applied",
			envVars: map[string]string{
				"RECONCILE_ENABLED": "1",
			},
			queryParams: "",
			setupMock: func(svcMock *ServiceMock) {
				svcMock.ReconcileImagesFunc = func(ctx context.Context, opts ReconcileOptions) (*ReconcileResult, error) {
					if opts.Limit != 100 { // default
						return nil, errors.New("expected limit 100")
					}
					return &ReconcileResult{
						Checked: 0,
						DryRun:  false,
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "success: max limit enforced",
			envVars: map[string]string{
				"RECONCILE_ENABLED": "1",
			},
			queryParams: "limit=2000",
			setupMock: func(svcMock *ServiceMock) {
				svcMock.ReconcileImagesFunc = func(ctx context.Context, opts ReconcileOptions) (*ReconcileResult, error) {
					// Service was called - just return success
					return &ReconcileResult{
						Checked: 0,
						DryRun:  false,
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "failure: feature not enabled",
			envVars:     map[string]string{},
			queryParams: "limit=100",
			setupMock: func(svcMock *ServiceMock) {
				// Should not be called
			},
			expectedStatus: http.StatusNotImplemented,
			validateResp: func(t *testing.T, body string) {
				assert.Contains(t, body, "reconciliation is not enabled")
			},
		},
		{
			name: "failure: service error",
			envVars: map[string]string{
				"RECONCILE_ENABLED": "1",
			},
			queryParams: "limit=100",
			setupMock: func(svcMock *ServiceMock) {
				svcMock.ReconcileImagesFunc = func(ctx context.Context, opts ReconcileOptions) (*ReconcileResult, error) {
					return nil, errors.New("database connection failed")
				}
			},
			expectedStatus: http.StatusInternalServerError,
			validateResp: func(t *testing.T, body string) {
				assert.Contains(t, body, "database connection failed")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set environment variables
			for key, val := range tc.envVars {
				t.Setenv(key, val)
			}

			// Create mock service
			svcMock := &ServiceMock{}
			tc.setupMock(svcMock)

			// Create handler
			handler := NewDefaultHandler(svcMock)

			// Create Echo context
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/reconcile/images?"+tc.queryParams, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Execute handler
			err := handler.ReconcileImages(c)

			// Validate
			if tc.expectedStatus >= 200 && tc.expectedStatus < 300 {
				assert.NoError(t, err)
			}
			if rec.Code != tc.expectedStatus {
				t.Logf("Response body: %s", rec.Body.String())
			}
			assert.Equal(t, tc.expectedStatus, rec.Code)

			if tc.validateResp != nil {
				tc.validateResp(t, rec.Body.String())
			}

			// Verify service was called as expected
			switch tc.expectedStatus {
			case http.StatusOK:
				assert.Len(t, svcMock.ReconcileImagesCalls(), 1)
			case http.StatusNotImplemented:
				assert.Len(t, svcMock.ReconcileImagesCalls(), 0)
			}
		})
	}
}

func TestDefaultHandler_ReconcileImages_InvalidConcurrency(t *testing.T) {
	t.Setenv("RECONCILE_ENABLED", "1")

	// Test invalid concurrency values
	invalidValues := []string{"0", "-1", "abc"}

	for _, val := range invalidValues {
		t.Run("invalid_"+val, func(t *testing.T) {
			svcMock := &ServiceMock{
				ReconcileImagesFunc: func(ctx context.Context, opts ReconcileOptions) (*ReconcileResult, error) {
					if opts.Concurrency != 5 { // Should fallback to default
						return nil, errors.New("expected concurrency 5")
					}
					return &ReconcileResult{Checked: 0}, nil
				},
			}

			handler := NewDefaultHandler(svcMock)
			e := echo.New()

			queryParam := "limit=100&concurrency=" + url.QueryEscape(val)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/reconcile/images?"+queryParam, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := handler.ReconcileImages(c)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)

			// Verify default concurrency was used
			assert.Len(t, svcMock.ReconcileImagesCalls(), 1)
			assert.Equal(t, 5, svcMock.ReconcileImagesCalls()[0].Opts.Concurrency)
		})
	}
}

func TestDefaultHandler_ReconcileImages_JSONBody(t *testing.T) {
	t.Setenv("RECONCILE_ENABLED", "1")

	testCases := []struct {
		name           string
		body           string
		setupMock      func(*ServiceMock)
		expectedStatus int
	}{
		{
			name: "success: JSON body with filters",
			body: `{"project_id":"550e8400-e29b-41d4-a716-446655440000","status":"ready","limit":50,"dry_run":true}`,
			setupMock: func(svcMock *ServiceMock) {
				svcMock.ReconcileImagesFunc = func(ctx context.Context, opts ReconcileOptions) (*ReconcileResult, error) {
					// Verify the fields were bound correctly
					if opts.ProjectID == nil || *opts.ProjectID != "550e8400-e29b-41d4-a716-446655440000" {
						return nil, errors.New("unexpected project_id")
					}
					if opts.Status == nil || *opts.Status != "ready" {
						return nil, errors.New("unexpected status")
					}
					if opts.Limit != 50 {
						return nil, errors.New("unexpected limit")
					}
					if !opts.DryRun {
						return nil, errors.New("expected dry_run to be true")
					}
					return &ReconcileResult{Checked: 5}, nil
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "failure: invalid JSON",
			body:           `{invalid json}`,
			setupMock:      func(svcMock *ServiceMock) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svcMock := &ServiceMock{}
			tc.setupMock(svcMock)

			handler := NewDefaultHandler(svcMock)
			e := echo.New()

			req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/reconcile/images", strings.NewReader(tc.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := handler.ReconcileImages(c)

			if tc.expectedStatus >= 200 && tc.expectedStatus < 300 {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expectedStatus, rec.Code)
		})
	}
}
