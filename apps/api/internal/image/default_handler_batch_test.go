package image

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestBatchCreateImages_Success(t *testing.T) {
	e := echo.New()
	projectID := uuid.New()

	reqBody := BatchCreateImagesRequest{
		Images: []CreateImageRequest{
			{
				ProjectID:   projectID,
				OriginalURL: "https://example.com/image1.jpg",
			},
			{
				ProjectID:   projectID,
				OriginalURL: "https://example.com/image2.jpg",
				RoomType:    stringPtr("bedroom"),
				Style:       stringPtr("modern"),
			},
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/images/batch", strings.NewReader(string(body)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Mock service
	serviceMock := &ServiceMock{
		BatchCreateImagesFunc: func(ctx context.Context, reqs []CreateImageRequest) (*BatchCreateImagesResponse, error) {
			images := []*Image{
				{
					ID:          uuid.New(),
					ProjectID:   projectID,
					OriginalURL: reqs[0].OriginalURL,
					Status:      StatusQueued,
				},
				{
					ID:          uuid.New(),
					ProjectID:   projectID,
					OriginalURL: reqs[1].OriginalURL,
					RoomType:    reqs[1].RoomType,
					Style:       reqs[1].Style,
					Status:      StatusQueued,
				},
			}
			return &BatchCreateImagesResponse{
				Images:  images,
				Errors:  []BatchImageError{},
				Success: 2,
				Failed:  0,
			}, nil
		},
	}

	handler := NewDefaultHandler(serviceMock)
	err := handler.BatchCreateImages(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var response BatchCreateImagesResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 2, response.Success)
	assert.Equal(t, 0, response.Failed)
	assert.Len(t, response.Images, 2)
	assert.Len(t, response.Errors, 0)
}

func TestBatchCreateImages_PartialSuccess(t *testing.T) {
	e := echo.New()
	projectID := uuid.New()

	reqBody := BatchCreateImagesRequest{
		Images: []CreateImageRequest{
			{
				ProjectID:   projectID,
				OriginalURL: "https://example.com/image1.jpg",
			},
			{
				ProjectID:   projectID,
				OriginalURL: "https://example.com/image2.jpg",
			},
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/images/batch", strings.NewReader(string(body)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Mock service with partial failure
	serviceMock := &ServiceMock{
		BatchCreateImagesFunc: func(ctx context.Context, reqs []CreateImageRequest) (*BatchCreateImagesResponse, error) {
			return &BatchCreateImagesResponse{
				Images: []*Image{
					{
						ID:          uuid.New(),
						ProjectID:   projectID,
						OriginalURL: reqs[0].OriginalURL,
						Status:      StatusQueued,
					},
				},
				Errors: []BatchImageError{
					{Index: 1, Message: "failed to create image"},
				},
				Success: 1,
				Failed:  1,
			}, nil
		},
	}

	handler := NewDefaultHandler(serviceMock)
	err := handler.BatchCreateImages(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusMultiStatus, rec.Code)

	var response BatchCreateImagesResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 1, response.Success)
	assert.Equal(t, 1, response.Failed)
	assert.Len(t, response.Images, 1)
	assert.Len(t, response.Errors, 1)
}

func TestBatchCreateImages_EmptyRequest(t *testing.T) {
	e := echo.New()

	reqBody := BatchCreateImagesRequest{
		Images: []CreateImageRequest{},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/images/batch", strings.NewReader(string(body)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	serviceMock := &ServiceMock{}
	handler := NewDefaultHandler(serviceMock)
	err := handler.BatchCreateImages(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)

	var response ValidationErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "validation_failed", response.Error)
	assert.Contains(t, response.ValidationErrors[0].Field, "images")
}

func TestBatchCreateImages_TooManyImages(t *testing.T) {
	e := echo.New()
	projectID := uuid.New()

	// Create 51 images
	images := make([]CreateImageRequest, 51)
	for i := 0; i < 51; i++ {
		images[i] = CreateImageRequest{
			ProjectID:   projectID,
			OriginalURL: "https://example.com/image.jpg",
		}
	}

	reqBody := BatchCreateImagesRequest{Images: images}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/images/batch", strings.NewReader(string(body)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	serviceMock := &ServiceMock{}
	handler := NewDefaultHandler(serviceMock)
	err := handler.BatchCreateImages(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)

	var response ValidationErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "validation_failed", response.Error)
	assert.Contains(t, response.ValidationErrors[0].Message, "maximum 50")
}

func TestBatchCreateImages_InvalidImageData(t *testing.T) {
	e := echo.New()

	reqBody := BatchCreateImagesRequest{
		Images: []CreateImageRequest{
			{
				ProjectID:   uuid.Nil, // Invalid
				OriginalURL: "",       // Invalid
			},
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/images/batch", strings.NewReader(string(body)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	serviceMock := &ServiceMock{}
	handler := NewDefaultHandler(serviceMock)
	err := handler.BatchCreateImages(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)

	var response ValidationErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "validation_failed", response.Error)
	assert.Greater(t, len(response.ValidationErrors), 0)
	// Should have validation errors for images[0].project_id and images[0].original_url
	hasProjectIDError := false
	hasOriginalURLError := false
	for _, err := range response.ValidationErrors {
		if strings.Contains(err.Field, "project_id") {
			hasProjectIDError = true
		}
		if strings.Contains(err.Field, "original_url") {
			hasOriginalURLError = true
		}
	}
	assert.True(t, hasProjectIDError)
	assert.True(t, hasOriginalURLError)
}

func stringPtr(s string) *string {
	return &s
}
