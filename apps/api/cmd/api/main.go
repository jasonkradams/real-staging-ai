package main

import (
	"context"
	"log"
	"os"

	"github.com/virtual-staging-ai/api/internal/http"
	"github.com/virtual-staging-ai/api/internal/image"
	"github.com/virtual-staging-ai/api/internal/job"
	"github.com/virtual-staging-ai/api/internal/storage"
)

// main is the entrypoint of the API server.
func main() {
	ctx := context.Background()
	db, err := storage.NewDefaultDatabase()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create S3 service
	bucketName := os.Getenv("S3_BUCKET_NAME")
	if bucketName == "" {
		bucketName = "virtual-staging-dev"
	}

	s3Service, err := storage.NewDefaultS3Service(ctx, bucketName)
	if err != nil {
		log.Fatalf("failed to create S3 service: %v", err)
	}

	// Create repositories
	imageRepo := image.NewDefaultRepository(db)
	jobRepo := job.NewDefaultRepository(db)

	// Create services
	imageService := image.NewDefaultService(imageRepo, jobRepo)

	s := http.NewServer(db, s3Service, imageService)
	if err := s.Start(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
