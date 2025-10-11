package main

import (
	"context"
	"fmt"

	"github.com/real-staging-ai/api/internal/config"
	"github.com/real-staging-ai/api/internal/http"
	"github.com/real-staging-ai/api/internal/image"
	"github.com/real-staging-ai/api/internal/job"
	"github.com/real-staging-ai/api/internal/logging"
	"github.com/real-staging-ai/api/internal/storage"
)

// main is the entrypoint of the API server.
func main() {
	log := logging.Default()
	ctx := context.Background()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Error(ctx, fmt.Sprintf("failed to load configuration: %v", err))
		return
	}
	log.Info(ctx, fmt.Sprintf("Loaded configuration for environment: %s", cfg.App.Env))

	db, err := storage.NewDefaultDatabase(&cfg.DB)
	if err != nil {
		log.Error(ctx, fmt.Sprintf("failed to connect to database: %v", err))
		return
	}
	defer db.Close()

	s3Service, err := storage.NewDefaultS3Service(ctx, &cfg.S3)
	if err != nil {
		log.Error(ctx, fmt.Sprintf("failed to create S3 service: %v", err))
	} else {
		// Ensure bucket exists in dev/local (MinIO) to avoid presign/upload failures
		if err := s3Service.CreateBucket(ctx); err != nil {
			log.Error(ctx, fmt.Sprintf("failed to ensure S3 bucket exists: %v", err))
		}
	}

	// Create repositories
	imageRepo := image.NewDefaultRepository(db)
	jobRepo := job.NewDefaultRepository(db)
	log.Info(ctx, fmt.Sprintf("Setting up image service (queue: %s)", cfg.Job.QueueName))
	imageService := image.NewDefaultService(cfg, imageRepo, jobRepo)

	s := http.NewServer(db, s3Service, imageService, cfg.Auth0.Domain, cfg.Auth0.Audience)
	if err := s.Start(":8080"); err != nil {
		log.Error(ctx, fmt.Sprintf("Failed to start server: %v", err))
	}
}
