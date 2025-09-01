package main

import (
	"context"
	"log"
	"os"

	"github.com/virtual-staging-ai/api/internal/http"
	"github.com/virtual-staging-ai/api/internal/services"
	"github.com/virtual-staging-ai/api/internal/storage"
)

// main is the entrypoint of the API server.
func main() {
	ctx := context.Background()
	db, err := storage.NewDB(ctx)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create S3 service
	bucketName := os.Getenv("S3_BUCKET_NAME")
	if bucketName == "" {
		bucketName = "virtual-staging-dev"
	}

	s3Service, err := storage.NewS3Service(ctx, bucketName)
	if err != nil {
		log.Fatalf("failed to create S3 service: %v", err)
	}

	// Create repositories
	imageRepo := storage.NewImageRepository(db)
	jobRepo := storage.NewJobRepository(db)

	// Create services
	imageService := services.NewImageService(imageRepo, jobRepo)

	s := http.NewServer(db, s3Service, imageService)
	s.Start(":8080")
}
