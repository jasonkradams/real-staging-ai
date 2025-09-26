package main

import (
	"context"
	"fmt"
	"os"

	"github.com/virtual-staging-ai/api/internal/http"
	"github.com/virtual-staging-ai/api/internal/image"
	"github.com/virtual-staging-ai/api/internal/job"
	"github.com/virtual-staging-ai/api/internal/logging"
	"github.com/virtual-staging-ai/api/internal/storage"
)

// main is the entrypoint of the API server.
func main() {
	log := logging.Default()
	ctx := context.Background()
	db, err := storage.NewDefaultDatabase()
	if err != nil {
		log.Error(ctx, fmt.Sprintf("failed to connect to database: %v", err))
	}
	defer db.Close()

	// Create S3 service
	bucketName := os.Getenv("S3_BUCKET_NAME")
	if bucketName == "" {
		bucketName = "virtual-staging-dev"
	}

	s3Service, err := storage.NewDefaultS3Service(ctx, bucketName)
	if err != nil {
		log.Error(ctx, fmt.Sprintf("failed to create S3 service: %v", err))
	}

	// Create repositories
	imageRepo := image.NewDefaultRepository(db)
	jobRepo := job.NewDefaultRepository(db)

	// Create services (image service sets up queue enqueuer from env, e.g., REDIS_ADDR)
	log.Info(ctx, "Setting up image service with job enqueuer (REDIS_ADDR, JOB_QUEUE_NAME)")
	imageService := image.NewDefaultService(imageRepo, jobRepo)

	s := http.NewServer(db, s3Service, imageService)
	if err := s.Start(":8080"); err != nil {
		log.Error(ctx, fmt.Sprintf("Failed to start server: %v", err))
	}
}
