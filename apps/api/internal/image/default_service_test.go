package image

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/real-staging-ai/api/internal/config"
	"github.com/real-staging-ai/api/internal/job"
	"github.com/real-staging-ai/api/internal/storage/queries"
)

func TestNewDefaultService(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	t.Run("success: create new default service", func(t *testing.T) {
		imageRepo := &RepositoryMock{}
		jobRepo := &job.RepositoryMock{}
		service := NewDefaultService(cfg, imageRepo, jobRepo)
		assert.NotNil(t, service)
	})
}

func TestDefaultService_CreateImage(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}

	projectID := uuid.New()
	imageID := uuid.New()

	testCases := []struct {
		name          string
		req           *CreateImageRequest
		setupMocks    func(*RepositoryMock, *job.RepositoryMock)
		expectedImage *Image
		expectedErr   error
	}{
		{
			name: "success: create image",
			req: &CreateImageRequest{
				ProjectID:   projectID,
				OriginalURL: "http://example.com/image.jpg",
			},
			setupMocks: func(imageRepo *RepositoryMock, jobRepo *job.RepositoryMock) {
				imageRepo.CreateImageFunc = func(ctx context.Context, projectIDStr, originalURL string, roomType, style *string, seed *int64) (*queries.Image, error) {
					return &queries.Image{
						ID:          pgtype.UUID{Bytes: imageID, Valid: true},
						ProjectID:   pgtype.UUID{Bytes: projectID, Valid: true},
						OriginalUrl: "http://example.com/image.jpg",
						Status:      queries.ImageStatusQueued,
						CreatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
						UpdatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
					}, nil
				}
				jobRepo.CreateJobFunc = func(ctx context.Context, imageID, jobType string, payloadJSON []byte) (*queries.Job, error) {
					return &queries.Job{}, nil
				}
			},
			expectedImage: &Image{
				ID:          imageID,
				ProjectID:   projectID,
				OriginalURL: "http://example.com/image.jpg",
				Status:      StatusQueued,
			},
			expectedErr: nil,
		},
		{
			name:        "fail: nil request",
			req:         nil,
			setupMocks:  func(imageRepo *RepositoryMock, jobRepo *job.RepositoryMock) {},
			expectedErr: errors.New("request cannot be nil"),
		},
		{
			name: "fail: image repo create image error",
			req: &CreateImageRequest{
				ProjectID:   projectID,
				OriginalURL: "http://example.com/image.jpg",
			},
			setupMocks: func(imageRepo *RepositoryMock, jobRepo *job.RepositoryMock) {
				imageRepo.CreateImageFunc = func(ctx context.Context, projectIDStr, originalURL string, roomType, style *string, seed *int64) (*queries.Image, error) {
					return nil, errors.New("db error")
				}
			},
			expectedErr: errors.New("failed to create image: db error"),
		},
		{
			name: "fail: job repo create job error",
			req: &CreateImageRequest{
				ProjectID:   projectID,
				OriginalURL: "http://example.com/image.jpg",
			},
			setupMocks: func(imageRepo *RepositoryMock, jobRepo *job.RepositoryMock) {
				imageRepo.CreateImageFunc = func(ctx context.Context, projectIDStr, originalURL string, roomType, style *string, seed *int64) (*queries.Image, error) {
					return &queries.Image{
						ID:          pgtype.UUID{Bytes: imageID, Valid: true},
						ProjectID:   pgtype.UUID{Bytes: projectID, Valid: true},
						OriginalUrl: "http://example.com/image.jpg",
						Status:      queries.ImageStatusQueued,
						CreatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
						UpdatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
					}, nil
				}
				jobRepo.CreateJobFunc = func(ctx context.Context, imageID, jobType string, payloadJSON []byte) (*queries.Job, error) {
					return nil, errors.New("job error")
				}
			},
			expectedErr: errors.New("failed to create job: job error"),
		},
		{
			name: "fail: json marshal error",
			req: &CreateImageRequest{
				ProjectID:   projectID,
				OriginalURL: "http://example.com/image.jpg",
			},
			setupMocks: func(imageRepo *RepositoryMock, jobRepo *job.RepositoryMock) {
				imageRepo.CreateImageFunc = func(ctx context.Context, projectIDStr, originalURL string, roomType, style *string, seed *int64) (*queries.Image, error) {
					return &queries.Image{
						ID:          pgtype.UUID{Bytes: imageID, Valid: true},
						ProjectID:   pgtype.UUID{Bytes: projectID, Valid: true},
						OriginalUrl: "http://example.com/image.jpg",
						Status:      queries.ImageStatusQueued,
						CreatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
						UpdatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
					}, nil
				}
				jobRepo.CreateJobFunc = func(ctx context.Context, imageID, jobType string, payloadJSON []byte) (*queries.Job, error) {
					return nil, errors.New("json marshal error")
				}
			},
			expectedErr: errors.New("failed to marshal job payload: json marshal error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			imageRepo := &RepositoryMock{}
			jobRepo := &job.RepositoryMock{}
			tc.setupMocks(imageRepo, jobRepo)

			service := NewDefaultService(cfg, imageRepo, jobRepo)

			if tc.name == "fail: json marshal error" {
				jsonMarshal = func(v interface{}) ([]byte, error) {
					return nil, errors.New("json marshal error")
				}
				defer func() {
					jsonMarshal = json.Marshal
				}()
			}

			image, err := service.CreateImage(context.Background(), tc.req)

			if tc.expectedErr != nil {
				assert.EqualError(t, err, tc.expectedErr.Error())
				assert.Nil(t, image)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, image)
				assert.Equal(t, tc.expectedImage.ID, image.ID)
				assert.Equal(t, tc.expectedImage.ProjectID, image.ProjectID)
				assert.Equal(t, tc.expectedImage.OriginalURL, image.OriginalURL)
				assert.Equal(t, tc.expectedImage.Status, image.Status)
			}
		})
	}
}

func TestDefaultService_GetImageByID(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}

	imageID := uuid.New()

	testCases := []struct {
		name          string
		imageID       string
		setupMocks    func(*RepositoryMock)
		expectedImage *Image
		expectedErr   error
	}{
		{
			name:    "success: get image by id",
			imageID: imageID.String(),
			setupMocks: func(imageRepo *RepositoryMock) {
				imageRepo.GetImageByIDFunc = func(ctx context.Context, imageIDStr string) (*queries.Image, error) {
					parsedImageID, _ := uuid.Parse(imageIDStr)
					return &queries.Image{
						ID:          pgtype.UUID{Bytes: parsedImageID, Valid: true},
						OriginalUrl: "http://example.com/image.jpg",
						Status:      queries.ImageStatusQueued,
						CreatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
						UpdatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
						StagedUrl:   pgtype.Text{String: "http://example.com/staged.jpg", Valid: true},
						RoomType:    pgtype.Text{String: "living_room", Valid: true},
						Style:       pgtype.Text{String: "modern", Valid: true},
						Seed:        pgtype.Int8{Int64: 123, Valid: true},
						Error:       pgtype.Text{String: "some error", Valid: true},
					}, nil
				}
			},
			expectedImage: &Image{
				ID:          imageID,
				OriginalURL: "http://example.com/image.jpg",
				Status:      StatusQueued,
			},
			expectedErr: nil,
		},
		{
			name:    "success: get image by id with null fields",
			imageID: imageID.String(),
			setupMocks: func(imageRepo *RepositoryMock) {
				imageRepo.GetImageByIDFunc = func(ctx context.Context, imageIDStr string) (*queries.Image, error) {
					parsedImageID, _ := uuid.Parse(imageIDStr)
					return &queries.Image{
						ID:          pgtype.UUID{Bytes: parsedImageID, Valid: true},
						OriginalUrl: "http://example.com/image.jpg",
						Status:      queries.ImageStatusQueued,
						CreatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
						UpdatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
					}, nil
				}
			},
			expectedImage: &Image{
				ID:          imageID,
				OriginalURL: "http://example.com/image.jpg",
				Status:      StatusQueued,
			},
			expectedErr: nil,
		},
		{
			name:        "fail: empty image id",
			imageID:     "",
			setupMocks:  func(imageRepo *RepositoryMock) {},
			expectedErr: errors.New("image ID cannot be empty"),
		},
		{
			name:    "fail: db error",
			imageID: imageID.String(),
			setupMocks: func(imageRepo *RepositoryMock) {
				imageRepo.GetImageByIDFunc = func(ctx context.Context, imageID string) (*queries.Image, error) {
					return nil, errors.New("db error")
				}
			},
			expectedErr: errors.New("failed to get image: db error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			imageRepo := &RepositoryMock{}
			tc.setupMocks(imageRepo)

			service := NewDefaultService(cfg, imageRepo, nil)
			image, err := service.GetImageByID(context.Background(), tc.imageID)

			if tc.expectedErr != nil {
				assert.EqualError(t, err, tc.expectedErr.Error())
				assert.Nil(t, image)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, image)
				assert.Equal(t, tc.expectedImage.ID, image.ID)
				assert.Equal(t, tc.expectedImage.OriginalURL, image.OriginalURL)
				assert.Equal(t, tc.expectedImage.Status, image.Status)
				if tc.name == "success: get image by id" {
					assert.NotNil(t, image.StagedURL)
					assert.NotNil(t, image.RoomType)
					assert.NotNil(t, image.Style)
					assert.NotNil(t, image.Seed)
					assert.NotNil(t, image.Error)
				} else {
					assert.Nil(t, image.StagedURL)
					assert.Nil(t, image.RoomType)
					assert.Nil(t, image.Style)
					assert.Nil(t, image.Seed)
					assert.Nil(t, image.Error)
				}
			}
		})
	}
}

func TestDefaultService_GetImagesByProjectID(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}

	projectID := uuid.New()

	testCases := []struct {
		name           string
		projectID      string
		setupMocks     func(*RepositoryMock)
		expectedImages []*Image
		expectedErr    error
	}{
		{
			name:      "success: get images by project id",
			projectID: projectID.String(),
			setupMocks: func(imageRepo *RepositoryMock) {
				imageRepo.GetImagesByProjectIDFunc = func(ctx context.Context, projectID string) ([]*queries.Image, error) {
					return []*queries.Image{
							{ID: pgtype.UUID{Bytes: uuid.New(), Valid: true}},
						},
						nil
				}
			},
			expectedImages: []*Image{
				{ID: uuid.New()},
			},
			expectedErr: nil,
		},
		{
			name:        "fail: empty project id",
			projectID:   "",
			setupMocks:  func(imageRepo *RepositoryMock) {},
			expectedErr: errors.New("project ID cannot be empty"),
		},
		{
			name:      "fail: db error",
			projectID: projectID.String(),
			setupMocks: func(imageRepo *RepositoryMock) {
				imageRepo.GetImagesByProjectIDFunc = func(ctx context.Context, projectID string) ([]*queries.Image, error) {
					return nil, errors.New("db error")
				}
			},
			expectedErr: errors.New("failed to get images: db error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			imageRepo := &RepositoryMock{}
			tc.setupMocks(imageRepo)

			service := NewDefaultService(cfg, imageRepo, nil)
			images, err := service.GetImagesByProjectID(context.Background(), tc.projectID)

			if tc.expectedErr != nil {
				assert.EqualError(t, err, tc.expectedErr.Error())
				assert.Nil(t, images)
			} else {
				assert.NoError(t, err)
				assert.Len(t, images, len(tc.expectedImages))
			}
		})
	}
}

func TestDefaultService_UpdateImageStatus(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}

	imageID := uuid.New()

	testCases := []struct {
		name          string
		imageID       string
		status        Status
		setupMocks    func(*RepositoryMock)
		expectedImage *Image
		expectedErr   error
	}{
		{
			name:    "success: update image status",
			imageID: imageID.String(),
			status:  StatusProcessing,
			setupMocks: func(imageRepo *RepositoryMock) {
				imageRepo.UpdateImageStatusFunc = func(ctx context.Context, imageID, status string) (*queries.Image, error) {
					return &queries.Image{},
						nil
				}
			},
			expectedImage: &Image{},
			expectedErr:   nil,
		},
		{
			name:        "fail: empty image id",
			imageID:     "",
			status:      StatusProcessing,
			setupMocks:  func(imageRepo *RepositoryMock) {},
			expectedErr: errors.New("image ID cannot be empty"),
		},
		{
			name:    "fail: db error",
			imageID: imageID.String(),
			status:  StatusProcessing,
			setupMocks: func(imageRepo *RepositoryMock) {
				imageRepo.UpdateImageStatusFunc = func(ctx context.Context, imageID, status string) (*queries.Image, error) {
					return nil, errors.New("db error")
				}
			},
			expectedErr: errors.New("failed to update image status: db error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			imageRepo := &RepositoryMock{}
			tc.setupMocks(imageRepo)

			service := NewDefaultService(cfg, imageRepo, nil)
			image, err := service.UpdateImageStatus(context.Background(), tc.imageID, tc.status)

			if tc.expectedErr != nil {
				assert.EqualError(t, err, tc.expectedErr.Error())
				assert.Nil(t, image)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, image)
			}
		})
	}
}

func TestDefaultService_UpdateImageWithStagedURL(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}

	imageID := uuid.New()

	testCases := []struct {
		name          string
		imageID       string
		stagedURL     string
		setupMocks    func(*RepositoryMock)
		expectedImage *Image
		expectedErr   error
	}{
		{
			name:      "success: update image with staged url",
			imageID:   imageID.String(),
			stagedURL: "http://example.com/staged.jpg",
			setupMocks: func(imageRepo *RepositoryMock) {
				imageRepo.UpdateImageWithStagedURLFunc = func(ctx context.Context, imageID, stagedURL, status string) (*queries.Image, error) {
					return &queries.Image{},
						nil
				}
			},
			expectedImage: &Image{},
			expectedErr:   nil,
		},
		{
			name:        "fail: empty image id",
			imageID:     "",
			stagedURL:   "http://example.com/staged.jpg",
			setupMocks:  func(imageRepo *RepositoryMock) {},
			expectedErr: errors.New("image ID cannot be empty"),
		},
		{
			name:        "fail: empty staged url",
			imageID:     imageID.String(),
			stagedURL:   "",
			setupMocks:  func(imageRepo *RepositoryMock) {},
			expectedErr: errors.New("staged URL cannot be empty"),
		},
		{
			name:      "fail: db error",
			imageID:   imageID.String(),
			stagedURL: "http://example.com/staged.jpg",
			setupMocks: func(imageRepo *RepositoryMock) {
				imageRepo.UpdateImageWithStagedURLFunc = func(ctx context.Context, imageID, stagedURL, status string) (*queries.Image, error) {
					return nil, errors.New("db error")
				}
			},
			expectedErr: errors.New("failed to update image with staged URL: db error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			imageRepo := &RepositoryMock{}
			tc.setupMocks(imageRepo)

			service := NewDefaultService(cfg, imageRepo, nil)
			image, err := service.UpdateImageWithStagedURL(context.Background(), tc.imageID, tc.stagedURL)

			if tc.expectedErr != nil {
				assert.EqualError(t, err, tc.expectedErr.Error())
				assert.Nil(t, image)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, image)
			}
		})
	}
}

func TestDefaultService_UpdateImageWithError(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}

	imageID := uuid.New()

	testCases := []struct {
		name          string
		imageID       string
		errorMsg      string
		setupMocks    func(*RepositoryMock)
		expectedImage *Image
		expectedErr   error
	}{
		{
			name:     "success: update image with error",
			imageID:  imageID.String(),
			errorMsg: "some error",
			setupMocks: func(imageRepo *RepositoryMock) {
				imageRepo.UpdateImageWithErrorFunc = func(ctx context.Context, imageID, errorMsg string) (*queries.Image, error) {
					return &queries.Image{},
						nil
				}
			},
			expectedImage: &Image{},
			expectedErr:   nil,
		},
		{
			name:        "fail: empty image id",
			imageID:     "",
			errorMsg:    "some error",
			setupMocks:  func(imageRepo *RepositoryMock) {},
			expectedErr: errors.New("image ID cannot be empty"),
		},
		{
			name:        "fail: empty error message",
			imageID:     imageID.String(),
			errorMsg:    "",
			setupMocks:  func(imageRepo *RepositoryMock) {},
			expectedErr: errors.New("error message cannot be empty"),
		},
		{
			name:     "fail: db error",
			imageID:  imageID.String(),
			errorMsg: "some error",
			setupMocks: func(imageRepo *RepositoryMock) {
				imageRepo.UpdateImageWithErrorFunc = func(ctx context.Context, imageID, errorMsg string) (*queries.Image, error) {
					return nil, errors.New("db error")
				}
			},
			expectedErr: errors.New("failed to update image with error: db error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			imageRepo := &RepositoryMock{}
			tc.setupMocks(imageRepo)

			service := NewDefaultService(cfg, imageRepo, nil)
			image, err := service.UpdateImageWithError(context.Background(), tc.imageID, tc.errorMsg)

			if tc.expectedErr != nil {
				assert.EqualError(t, err, tc.expectedErr.Error())
				assert.Nil(t, image)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, image)
			}
		})
	}
}

func TestDefaultService_DeleteImage(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}

	imageID := uuid.New()

	testCases := []struct {
		name        string
		imageID     string
		setupMocks  func(*RepositoryMock)
		expectedErr error
	}{
		{
			name:    "success: delete image",
			imageID: imageID.String(),
			setupMocks: func(imageRepo *RepositoryMock) {
				imageRepo.DeleteImageFunc = func(ctx context.Context, imageID string) error {
					return nil
				}
			},
			expectedErr: nil,
		},
		{
			name:        "fail: empty image id",
			imageID:     "",
			setupMocks:  func(imageRepo *RepositoryMock) {},
			expectedErr: errors.New("image ID cannot be empty"),
		},
		{
			name:    "fail: db error",
			imageID: imageID.String(),
			setupMocks: func(imageRepo *RepositoryMock) {
				imageRepo.DeleteImageFunc = func(ctx context.Context, imageID string) error {
					return errors.New("db error")
				}
			},
			expectedErr: errors.New("failed to delete image: db error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			imageRepo := &RepositoryMock{}
			tc.setupMocks(imageRepo)

			service := NewDefaultService(cfg, imageRepo, nil)
			err := service.DeleteImage(context.Background(), tc.imageID)

			if tc.expectedErr != nil {
				assert.EqualError(t, err, tc.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
