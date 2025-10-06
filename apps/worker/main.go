package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/virtual-staging-ai/worker/internal/config"
	"github.com/virtual-staging-ai/worker/internal/events"
	"github.com/virtual-staging-ai/worker/internal/logging"
	"github.com/virtual-staging-ai/worker/internal/processor"
	"github.com/virtual-staging-ai/worker/internal/queue"
	"github.com/virtual-staging-ai/worker/internal/repository"
	"github.com/virtual-staging-ai/worker/internal/staging"
	"github.com/virtual-staging-ai/worker/internal/telemetry"
)

func main() {
	log := logging.Default()
	ctx := context.Background()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Error(ctx, fmt.Sprintf("Failed to load configuration: %v", err))
		os.Exit(1)
	}
	log.Info(ctx, fmt.Sprintf("Loaded configuration for environment: %s", cfg.App.Env))

	// Initialize OpenTelemetry
	shutdown, err := telemetry.InitTracing(ctx, "virtual-staging-worker")
	if err != nil {
		log.Error(ctx, "Failed to initialize tracing:", err)
	}
	defer func() {
		if err := shutdown(ctx); err != nil {
			log.Error(ctx, fmt.Sprintf("Failed to shutdown tracing: %v", err))
		}
	}()

	// Initialize database connection using config
	dsn := cfg.DatabaseURL()
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Error(ctx, fmt.Sprintf("Failed to open database: %v", err))
		os.Exit(1)
	}
	if err := db.PingContext(ctx); err != nil {
		log.Error(ctx, fmt.Sprintf("Failed to connect to database: %v", err))
		os.Exit(1)
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Error(ctx, fmt.Sprintf("Failed to close database: %v", err))
		}
	}()

	imgRepo := repository.NewImageRepository(db)

	// Initialize the staging service with config
	stagingCfg := &staging.ServiceConfig{
		BucketName:     cfg.S3Bucket(),
		ReplicateToken: cfg.Replicate.APIToken,
		ModelVersion:   cfg.Replicate.ModelVersion,
		S3Endpoint:     cfg.S3.Endpoint,
		S3Region:       cfg.S3.Region,
		S3AccessKey:    cfg.S3.AccessKey,
		S3SecretKey:    cfg.S3.SecretKey,
		S3UsePathStyle: cfg.S3.UsePathStyle,
		AppEnv:         cfg.App.Env,
	}
	stagingService, err := staging.NewDefaultService(ctx, stagingCfg)
	if err != nil {
		log.Error(ctx, fmt.Sprintf("Failed to initialize staging service: %v", err))
		os.Exit(1)
	}

	log.Info(ctx, "Starting Virtual Staging AI Worker...")
	// Create context that listens for the interrupt signal from the OS
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Initialize events publisher (Redis) if configured
	var pub events.Publisher
	if p, err := events.NewDefaultPublisher(cfg); err == nil {
		pub = p
		log.Info(ctx, "Events publisher enabled")
	} else {
		log.Info(ctx, "Events publisher disabled (no REDIS_ADDR)")
		pub = &events.NoopPublisher{}
	}

	// Initialize the job processor
	proc := processor.NewImageProcessor(imgRepo, stagingService, pub)

	// Initialize the queue client (Redis/asynq in production)
	var queueClient queue.QueueClient
	// Log queue-related configuration for clarity
	redisAddr := cfg.Redis.Addr
	queueName := cfg.Job.QueueName
	concurrency := cfg.Job.WorkerConcurrency
	log.Info(ctx, "Queue configuration", "redis_addr", redisAddr, "queue", queueName, "concurrency", concurrency)
	if qc, err := queue.NewAsynqQueueClient(cfg); err == nil {
		queueClient = qc
		log.Info(ctx, "Using Asynq queue backend")
	} else {
		queueClient = queue.NewMockQueueClient()
		log.Info(ctx, "Using mock queue backend (no Redis Address configured)")
	}

	// Start processing jobs
	go func() {
		log.Info(ctx, "Job polling loop started")
		pollCount := 0
		for {
			select {
			case <-ctx.Done():
				log.Info(ctx, "Shutting down worker...")
				return
			default:
				// Poll for jobs
				job, err := queueClient.GetNextJob(ctx)
				if err != nil {
					log.Error(ctx, fmt.Sprintf("Error getting next job: %v", err))
					time.Sleep(5 * time.Second)
					continue
				}

				if job == nil {
					// No jobs available, wait a bit
					pollCount++
					if pollCount%30 == 0 {
						log.Info(ctx, "Still polling for jobs...", "poll_count", pollCount)
					}
					time.Sleep(2 * time.Second)
					continue
				}

				// Reset counter when job is found
				pollCount = 0

				// Process the job
				log.Info(ctx, fmt.Sprintf("Processing job %s of type %s", job.ID, job.Type))

				// The processor handles all DB updates and SSE events internally
				if err := proc.ProcessJob(ctx, job); err != nil {
					log.Error(ctx, fmt.Sprintf("Error processing job %s: %v", job.ID, err))
					if markErr := queueClient.MarkJobFailed(ctx, job.ID, err.Error()); markErr != nil {
						log.Error(ctx, fmt.Sprintf("Failed to mark job %s as failed: %v", job.ID, markErr))
					}
				} else {
					log.Info(ctx, fmt.Sprintf("Successfully processed job %s", job.ID))
					if markErr := queueClient.MarkJobCompleted(ctx, job.ID); markErr != nil {
						log.Error(ctx, fmt.Sprintf("Failed to mark job %s as completed: %v", job.ID, markErr))
					}
				}
			}
		}
	}()

	log.Info(ctx, "Worker started. Press Ctrl+C to stop.")
	<-ctx.Done()
	log.Info(ctx, "Worker stopped.")
}
