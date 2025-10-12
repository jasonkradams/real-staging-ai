package image

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestDefaultHandler_CreateImage(t *testing.T) {
	testCases := []struct {
		name          string
		requestBody   string
		setupMock     func(*ServiceMock)
		expectedCode  int
		expectedError string
	}{
		{
			name: "success: create image",
			requestBody: `{"project_id": "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11", ` +
				`"original_url": "http://example.com/image.jpg"}`,
			setupMock: func(mock *ServiceMock) {
				mock.CreateImageFunc = func(ctx context.Context, req *CreateImageRequest) (*Image, error) {
					return &Image{ID: uuid.New()}, nil
				}
			},
			expectedCode: http.StatusCreated,
		},
		{
			name:         "fail: bad request - invalid json",
			requestBody:  `{"project_id": "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"`,
			setupMock:    func(mock *ServiceMock) {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "fail: validation error - missing original_url",
			requestBody:  `{"project_id": "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"}`,
			setupMock:    func(mock *ServiceMock) {},
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name:         "fail: validation error",
			requestBody:  `{"original_url": "http://example.com/image.jpg"}`,
			setupMock:    func(mock *ServiceMock) {},
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name: "fail: service error",
			requestBody: `{"project_id": "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11", ` +
				`"original_url": "http://example.com/image.jpg"}`,
			setupMock: func(mock *ServiceMock) {
				mock.CreateImageFunc = func(ctx context.Context, req *CreateImageRequest) (*Image, error) {
					return nil, errors.New("service error")
				}
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(tc.requestBody)))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			serviceMock := &ServiceMock{}
			tc.setupMock(serviceMock)

			h := NewDefaultHandler(serviceMock)

			if assert.NoError(t, h.CreateImage(c)) {
				assert.Equal(t, tc.expectedCode, rec.Code)
			}
		})
	}
}

func TestDefaultHandler_GetImage(t *testing.T) {
	testCases := []struct {
		name         string
		imageID      string
		setupMock    func(*ServiceMock)
		expectedCode int
	}{
		{
			name:    "success: get image",
			imageID: uuid.New().String(),
			setupMock: func(mock *ServiceMock) {
				mock.GetImageByIDFunc = func(ctx context.Context, imageID string) (*Image, error) {
					return &Image{ID: uuid.New()}, nil
				}
			},
			expectedCode: http.StatusOK,
		},
		{
			name:         "fail: bad request - missing image ID",
			imageID:      "",
			setupMock:    func(mock *ServiceMock) {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "fail: bad request - invalid image ID",
			imageID:      "invalid-uuid",
			setupMock:    func(mock *ServiceMock) {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:    "fail: service error - not found",
			imageID: uuid.New().String(),
			setupMock: func(mock *ServiceMock) {
				mock.GetImageByIDFunc = func(ctx context.Context, imageID string) (*Image, error) {
					return nil, errors.New("no rows in result set")
				}
			},
			expectedCode: http.StatusNotFound,
		},
		{
			name:    "fail: service error - internal server error",
			imageID: uuid.New().String(),
			setupMock: func(mock *ServiceMock) {
				mock.GetImageByIDFunc = func(ctx context.Context, imageID string) (*Image, error) {
					return nil, errors.New("some other error")
				}
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tc.imageID)

			serviceMock := &ServiceMock{}
			tc.setupMock(serviceMock)

			h := NewDefaultHandler(serviceMock)

			if assert.NoError(t, h.GetImage(c)) {
				assert.Equal(t, tc.expectedCode, rec.Code)
			}
		})
	}
}

func TestDefaultHandler_GetProjectImages(t *testing.T) {
	testCases := []struct {
		name         string
		projectID    string
		setupMock    func(*ServiceMock)
		expectedCode int
	}{
		{
			name:      "success: get project images",
			projectID: uuid.New().String(),
			setupMock: func(mock *ServiceMock) {
				mock.GetImagesByProjectIDFunc = func(ctx context.Context, projectID string) ([]*Image, error) {
					return []*Image{}, nil
				}
			},
			expectedCode: http.StatusOK,
		},
		{
			name:         "fail: bad request - missing project ID",
			projectID:    "",
			setupMock:    func(mock *ServiceMock) {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "fail: bad request - invalid project ID",
			projectID:    "invalid-uuid",
			setupMock:    func(mock *ServiceMock) {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:      "fail: service error",
			projectID: uuid.New().String(),
			setupMock: func(mock *ServiceMock) {
				mock.GetImagesByProjectIDFunc = func(ctx context.Context, projectID string) ([]*Image, error) {
					return nil, errors.New("service error")
				}
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("project_id")
			c.SetParamValues(tc.projectID)

			serviceMock := &ServiceMock{}
			tc.setupMock(serviceMock)

			h := NewDefaultHandler(serviceMock)

			if assert.NoError(t, h.GetProjectImages(c)) {
				assert.Equal(t, tc.expectedCode, rec.Code)
			}
		})
	}
}

func TestDefaultHandler_DeleteImage(t *testing.T) {
	testCases := []struct {
		name         string
		imageID      string
		setupMock    func(*ServiceMock)
		expectedCode int
	}{
		{
			name:    "success: delete image",
			imageID: uuid.New().String(),
			setupMock: func(mock *ServiceMock) {
				mock.DeleteImageFunc = func(ctx context.Context, imageID string) error {
					return nil
				}
			},
			expectedCode: http.StatusNoContent,
		},
		{
			name:         "fail: bad request - missing image ID",
			imageID:      "",
			setupMock:    func(mock *ServiceMock) {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "fail: bad request - invalid image ID",
			imageID:      "invalid-uuid",
			setupMock:    func(mock *ServiceMock) {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:    "fail: service error - not found",
			imageID: uuid.New().String(),
			setupMock: func(mock *ServiceMock) {
				mock.DeleteImageFunc = func(ctx context.Context, imageID string) error {
					return errors.New("no rows in result set")
				}
			},
			expectedCode: http.StatusNotFound,
		},
		{
			name:    "fail: service error - internal server error",
			imageID: uuid.New().String(),
			setupMock: func(mock *ServiceMock) {
				mock.DeleteImageFunc = func(ctx context.Context, imageID string) error {
					return errors.New("some other error")
				}
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodDelete, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tc.imageID)

			serviceMock := &ServiceMock{}
			tc.setupMock(serviceMock)

			h := NewDefaultHandler(serviceMock)

			if assert.NoError(t, h.DeleteImage(c)) {
				assert.Equal(t, tc.expectedCode, rec.Code)
			}
		})
	}
}

func TestDefaultHandler_validateCreateImageRequest(t *testing.T) {
	projectID := uuid.New()
	roomType := "living_room"
	style := "modern"
	seed := int64(123)

	testCases := []struct {
		name        string
		req         *CreateImageRequest
		expectError bool
	}{
		{
			name: "success: valid request",
			req: &CreateImageRequest{
				ProjectID:   projectID,
				OriginalURL: "http://example.com/image.jpg",
				RoomType:    &roomType,
				Style:       &style,
				Seed:        &seed,
			},
			expectError: false,
		},
		{
			name: "fail: missing project ID",
			req: &CreateImageRequest{
				OriginalURL: "http://example.com/image.jpg",
			},
			expectError: true,
		},
		{
			name: "fail: missing original URL",
			req: &CreateImageRequest{
				ProjectID: projectID,
			},
			expectError: true,
		},
		{
			name: "fail: invalid room type",
			req: &CreateImageRequest{
				ProjectID:   projectID,
				OriginalURL: "http://example.com/image.jpg",
				RoomType:    func() *string { s := "invalid"; return &s }(),
			},
			expectError: true,
		},
		{
			name: "fail: invalid style",
			req: &CreateImageRequest{
				ProjectID:   projectID,
				OriginalURL: "http://example.com/image.jpg",
				Style:       func() *string { s := "invalid"; return &s }(),
			},
			expectError: true,
		},
		{
			name: "fail: invalid seed (too small)",
			req: &CreateImageRequest{
				ProjectID:   projectID,
				OriginalURL: "http://example.com/image.jpg",
				Seed:        func() *int64 { i := int64(0); return &i }(),
			},
			expectError: true,
		},
		{
			name: "fail: invalid seed (too large)",
			req: &CreateImageRequest{
				ProjectID:   projectID,
				OriginalURL: "http://example.com/image.jpg",
				Seed:        func() *int64 { i := int64(4294967296); return &i }(),
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := NewDefaultHandler(nil)
			errs := h.validateCreateImageRequest(tc.req)
			if tc.expectError {
				assert.NotEmpty(t, errs)
			} else {
				assert.Empty(t, errs)
			}
		})
	}
}
