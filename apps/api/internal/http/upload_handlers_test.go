//go:build integration

package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	httpLib "github.com/virtual-staging-ai/api/internal/http"
	"github.com/virtual-staging-ai/api/internal/storage"
	"github.com/virtual-staging-ai/api/internal/testutil"
)

type PresignUploadRequest struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	FileSize    int64  `json:"file_size"`
}

type PresignUploadResponse struct {
	UploadURL string `json:"upload_url"`
	FileKey   string `json:"file_key"`
	ExpiresIn int64  `json:"expires_in"`
}

func TestPresignUpload(t *testing.T) {
	ctx := context.Background()
	db, err := storage.NewDB(ctx)
	require.NoError(t, err)
	defer db.Close()

	testCases := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
		validate       func(t *testing.T, response []byte)
	}{
		{
			name: "success: valid JPEG upload request",
			requestBody: PresignUploadRequest{
				Filename:    "test-image.jpg",
				ContentType: "image/jpeg",
				FileSize:    1024000, // 1MB
			},
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, response []byte) {
				var resp PresignUploadResponse
				err := json.Unmarshal(response, &resp)
				require.NoError(t, err)
				assert.NotEmpty(t, resp.UploadURL)
				assert.Contains(t, resp.UploadURL, "https://")
				assert.NotEmpty(t, resp.FileKey)
				assert.Contains(t, resp.FileKey, "uploads/")
				assert.Contains(t, resp.FileKey, "test-image")
				assert.Contains(t, resp.FileKey, ".jpg")
				assert.Equal(t, int64(900), resp.ExpiresIn) // 15 minutes
			},
		},
		{
			name: "success: valid PNG upload request",
			requestBody: PresignUploadRequest{
				Filename:    "screenshot.png",
				ContentType: "image/png",
				FileSize:    2048000, // 2MB
			},
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, response []byte) {
				var resp PresignUploadResponse
				err := json.Unmarshal(response, &resp)
				require.NoError(t, err)
				assert.Contains(t, resp.FileKey, ".png")
			},
		},
		{
			name: "success: valid WebP upload request",
			requestBody: PresignUploadRequest{
				Filename:    "modern-image.webp",
				ContentType: "image/webp",
				FileSize:    1500000, // 1.5MB
			},
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, response []byte) {
				var resp PresignUploadResponse
				err := json.Unmarshal(response, &resp)
				require.NoError(t, err)
				assert.Contains(t, resp.FileKey, ".webp")
			},
		},
		{
			name: "success: maximum file size",
			requestBody: PresignUploadRequest{
				Filename:    "large-image.jpg",
				ContentType: "image/jpeg",
				FileSize:    10485760, // 10MB (max allowed)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "success: filename with spaces and special chars",
			requestBody: PresignUploadRequest{
				Filename:    "My Living Room Photo (1).jpeg",
				ContentType: "image/jpeg",
				FileSize:    1024000,
			},
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, response []byte) {
				var resp PresignUploadResponse
				err := json.Unmarshal(response, &resp)
				require.NoError(t, err)
				assert.Contains(t, resp.FileKey, "My Living Room Photo (1)")
			},
		},
		{
			name: "fail: missing filename",
			requestBody: PresignUploadRequest{
				ContentType: "image/jpeg",
				FileSize:    1024000,
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedError:  "validation_failed",
		},
		{
			name: "fail: empty filename",
			requestBody: PresignUploadRequest{
				Filename:    "",
				ContentType: "image/jpeg",
				FileSize:    1024000,
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedError:  "validation_failed",
		},
		{
			name: "fail: filename too long",
			requestBody: PresignUploadRequest{
				Filename:    strings.Repeat("a", 256) + ".jpg",
				ContentType: "image/jpeg",
				FileSize:    1024000,
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedError:  "validation_failed",
		},
		{
			name: "fail: invalid file extension",
			requestBody: PresignUploadRequest{
				Filename:    "document.pdf",
				ContentType: "application/pdf",
				FileSize:    1024000,
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedError:  "validation_failed",
		},
		{
			name: "fail: no file extension",
			requestBody: PresignUploadRequest{
				Filename:    "imagefile",
				ContentType: "image/jpeg",
				FileSize:    1024000,
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedError:  "validation_failed",
		},
		{
			name: "fail: missing content type",
			requestBody: PresignUploadRequest{
				Filename: "test.jpg",
				FileSize: 1024000,
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedError:  "validation_failed",
		},
		{
			name: "fail: invalid content type",
			requestBody: PresignUploadRequest{
				Filename:    "test.jpg",
				ContentType: "application/octet-stream",
				FileSize:    1024000,
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedError:  "validation_failed",
		},
		{
			name: "fail: content type doesn't match extension",
			requestBody: PresignUploadRequest{
				Filename:    "test.png",
				ContentType: "image/jpeg",
				FileSize:    1024000,
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedError:  "validation_failed",
		},
		{
			name: "fail: missing file size",
			requestBody: PresignUploadRequest{
				Filename:    "test.jpg",
				ContentType: "image/jpeg",
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedError:  "validation_failed",
		},
		{
			name: "fail: zero file size",
			requestBody: PresignUploadRequest{
				Filename:    "test.jpg",
				ContentType: "image/jpeg",
				FileSize:    0,
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedError:  "validation_failed",
		},
		{
			name: "fail: negative file size",
			requestBody: PresignUploadRequest{
				Filename:    "test.jpg",
				ContentType: "image/jpeg",
				FileSize:    -1000,
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedError:  "validation_failed",
		},
		{
			name: "fail: file size too large",
			requestBody: PresignUploadRequest{
				Filename:    "huge-image.jpg",
				ContentType: "image/jpeg",
				FileSize:    10485761, // 10MB + 1 byte
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedError:  "validation_failed",
		},
		{
			name:           "fail: malformed JSON",
			requestBody:    `{"filename":}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "bad_request",
		},
		{
			name: "fail: invalid field type",
			requestBody: map[string]interface{}{
				"filename":     "test.jpg",
				"content_type": "image/jpeg",
				"file_size":    "not-a-number",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "bad_request",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup clean database state
			testutil.TruncateTables(t, db.GetPool())
			testutil.SeedTables(t, db.GetPool())

			server := httpLib.NewServer(db)

			// Prepare request body
			var body []byte
			if str, ok := tc.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tc.requestBody)
				require.NoError(t, err)
			}

			// Create request
			req := httptest.NewRequest(http.MethodPost, "/api/v1/uploads/presign", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			// TODO: Add Authorization header when auth middleware is implemented
			rec := httptest.NewRecorder()

			// Execute request
			server.ServeHTTP(rec, req)

			// Assert status code
			assert.Equal(t, tc.expectedStatus, rec.Code)

			// Assert error response if expected
			if tc.expectedError != "" {
				var errResp ErrorResponse
				err := json.Unmarshal(rec.Body.Bytes(), &errResp)
				require.NoError(t, err)
				assert.Equal(t, tc.expectedError, errResp.Error)
				assert.NotEmpty(t, errResp.Message)
			}

			// Run custom validation if provided
			if tc.validate != nil {
				tc.validate(t, rec.Body.Bytes())
			}
		})
	}
}

func TestPresignUpload_ValidationErrorDetails(t *testing.T) {
	ctx := context.Background()
	db, err := storage.NewDB(ctx)
	require.NoError(t, err)
	defer db.Close()

	testutil.TruncateTables(t, db.GetPool())
	testutil.SeedTables(t, db.GetPool())

	server := httpLib.NewServer(db)

	testCases := []struct {
		name                 string
		request              PresignUploadRequest
		expectedFields       []string
		expectedFieldMessage map[string]string
	}{
		{
			name: "multiple validation errors",
			request: PresignUploadRequest{
				Filename:    "",
				ContentType: "invalid/type",
				FileSize:    0,
			},
			expectedFields: []string{"filename", "content_type", "file_size"},
			expectedFieldMessage: map[string]string{
				"filename":     "filename is required",
				"content_type": "content_type must be image/jpeg, image/png, or image/webp",
				"file_size":    "file_size must be greater than 0",
			},
		},
		{
			name: "content type mismatch with extension",
			request: PresignUploadRequest{
				Filename:    "image.png",
				ContentType: "image/jpeg",
				FileSize:    1024000,
			},
			expectedFields: []string{"content_type"},
			expectedFieldMessage: map[string]string{
				"content_type": "content_type image/jpeg doesn't match file extension .png",
			},
		},
		{
			name: "invalid extension with valid content type",
			request: PresignUploadRequest{
				Filename:    "file.txt",
				ContentType: "image/jpeg",
				FileSize:    1024000,
			},
			expectedFields: []string{"filename", "content_type"},
			expectedFieldMessage: map[string]string{
				"filename":     "filename must have a valid image extension (.jpg, .jpeg, .png, .webp)",
				"content_type": "content_type image/jpeg doesn't match file extension .txt",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body, err := json.Marshal(tc.request)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/uploads/presign", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			server.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)

			var response ValidationErrorResponse
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, "validation_failed", response.Error)
			assert.NotEmpty(t, response.Message)

			// Check that all expected fields are present
			foundFields := make(map[string]bool)
			for _, validationError := range response.ValidationErrors {
				foundFields[validationError.Field] = true

				// Check specific error messages if provided
				if expectedMsg, exists := tc.expectedFieldMessage[validationError.Field]; exists {
					assert.Equal(t, expectedMsg, validationError.Message)
				}
			}

			// Verify all expected fields are present
			for _, expectedField := range tc.expectedFields {
				assert.True(t, foundFields[expectedField], "Expected validation error for field: %s", expectedField)
			}
		})
	}
}

func TestPresignUpload_Integration(t *testing.T) {
	// Note: This test would require actual AWS credentials and S3 setup
	// For now, we'll skip it in the regular test suite
	t.Skip("Integration test requires AWS S3 setup")

	ctx := context.Background()
	db, err := storage.NewDB(ctx)
	require.NoError(t, err)
	defer db.Close()

	testutil.TruncateTables(t, db.GetPool())
	testutil.SeedTables(t, db.GetPool())

	server := httpLib.NewServer(db)

	// Test with valid request
	requestBody := PresignUploadRequest{
		Filename:    "integration-test.jpg",
		ContentType: "image/jpeg",
		FileSize:    1024000,
	}

	body, err := json.Marshal(requestBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/uploads/presign", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var response PresignUploadResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	// Validate response structure
	assert.NotEmpty(t, response.UploadURL)
	assert.NotEmpty(t, response.FileKey)
	assert.Greater(t, response.ExpiresIn, int64(0))

	// TODO: Test actual upload to the presigned URL
	// This would involve making an HTTP PUT request to response.UploadURL
	// with the file content and verifying it was uploaded successfully
}
