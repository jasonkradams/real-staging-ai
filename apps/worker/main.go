package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"encoding/json"

	_ "github.com/lib/pq"
	"github.com/virtual-staging-ai/worker/internal/events"
	"github.com/virtual-staging-ai/worker/internal/logging"
	"github.com/virtual-staging-ai/worker/internal/processor"
	"github.com/virtual-staging-ai/worker/internal/queue"
	"github.com/virtual-staging-ai/worker/internal/repository"
	"github.com/virtual-staging-ai/worker/internal/telemetry"
)

func main() {
	log := logging.Default()
	ctx := context.Background()

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

	// Initialize database connection
	host := os.Getenv("PGHOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("PGPORT")
	if port == "" {
		port = "5432"
	}
	user := os.Getenv("PGUSER")
	if user == "" {
		user = "postgres"
	}
	pass := os.Getenv("PGPASSWORD")
	if pass == "" {
		pass = "postgres"
	}
	dbname := os.Getenv("PGDATABASE")
	if dbname == "" {
		dbname = "virtualstaging"
	}
	sslmode := os.Getenv("PGSSLMODE")
	if sslmode == "" {
		sslmode = "disable"
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, pass, host, port, dbname, sslmode)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Error(ctx, fmt.Sprintf("Failed to open database: %v", err))
	}
	if err := db.PingContext(ctx); err != nil {
		log.Error(ctx, fmt.Sprintf("Failed to connect to database: %v", err))
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Error(ctx, fmt.Sprintf("Failed to close database: %v", err))
		}
	}()

	imgRepo := repository.NewImageRepository(db)

	log.Info(ctx, "Starting Virtual Staging AI Worker...")
	// Create context that listens for the interrupt signal from the OS
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Initialize the job processor
	proc := processor.NewImageProcessor()

	// Initialize the queue client (Redis/asynq in production)
	var queueClient queue.QueueClient
	// Log queue-related configuration for clarity
	redisAddr := os.Getenv("REDIS_ADDR")
	queueName := os.Getenv("JOB_QUEUE_NAME")
	if queueName == "" {
		queueName = "default"
	}
	concStr := os.Getenv("WORKER_CONCURRENCY")
	log.Info(ctx, "Queue configuration", "redis_addr", redisAddr, "queue", queueName, "concurrency", concStr)
	if qc, err := queue.NewAsynqQueueClientFromEnv(); err == nil {
		queueClient = qc
		log.Info(ctx, "Using Asynq queue backend")
	} else {
		queueClient = queue.NewMockQueueClient()
		log.Info(ctx, "Using mock queue backend (no REDIS_ADDR configured)")
	}

	// Initialize events publisher (Redis) if configured
	var pub events.Publisher
	if p, err := events.NewDefaultPublisherFromEnv(); err == nil {
		pub = p
		log.Info(ctx, "Events publisher enabled", "redis_addr", redisAddr)
	} else {
		log.Info(ctx, "Events publisher disabled (no REDIS_ADDR)")
	}

	log.Info(ctx, "Worker started. Press Ctrl+C to stop.")

	// Start processing jobs
	go func() {
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
					time.Sleep(2 * time.Second)
					continue
				}

				// Process the job
				log.Info(ctx, fmt.Sprintf("Processing job %s of type %s", job.ID, job.Type))

				// Decode payload (image id + original url)
				var payload struct {
					ImageID     string `json:"image_id"`
					OriginalURL string `json:"original_url"`
				}
				if err := json.Unmarshal(job.Payload, &payload); err != nil {
					log.Error(ctx, fmt.Sprintf("Failed to decode job payload: %v", err))
				}

				// Update DB: processing
				if err := imgRepo.SetProcessing(ctx, payload.ImageID); err != nil {
					log.Error(ctx, fmt.Sprintf("Failed to set image %s processing: %v", payload.ImageID, err))
				}
				// Publish 'processing' status
				if pub != nil {
					_ = pub.PublishJobUpdate(ctx, events.JobUpdateEvent{JobID: job.ID, ImageID: payload.ImageID, Status: "processing"})
				}
				// Process the job
				if err := proc.ProcessJob(ctx, job); err != nil {
					log.Error(ctx, fmt.Sprintf("Error processing job %s: %v", job.ID, err))
					if markErr := queueClient.MarkJobFailed(ctx, job.ID, err.Error()); markErr != nil {
						log.Error(ctx, fmt.Sprintf("Failed to mark job %s as failed: %v", job.ID, markErr))
				}
					// Update DB: error
					if setErr := imgRepo.SetError(ctx, payload.ImageID, err.Error()); setErr != nil {
						log.Error(ctx, fmt.Sprintf("Failed to set image %s error: %v", payload.ImageID, setErr))
					}
					// Publish 'error' status
					if pub != nil {
						if err := pub.PublishJobUpdate(ctx, events.JobUpdateEvent{JobID: job.ID, ImageID: payload.ImageID, Status: "error", Error: err.Error()}); err != nil {
							log.Error(ctx, fmt.Sprintf("Failed to publish job %s error: %v", job.ID, err))
						}
					}
				} else {
					log.Info(ctx, fmt.Sprintf("Successfully processed job %s", job.ID))
					if markErr := queueClient.MarkJobCompleted(ctx, job.ID); markErr != nil {
						log.Error(ctx, fmt.Sprintf("Failed to mark job %s as completed: %v", job.ID, markErr))
					}
					// Update DB: ready + staged_url
					stagedURL := fmt.Sprintf("%s-staged.jpg", payload.OriginalURL)
					if setErr := imgRepo.SetReady(ctx, payload.ImageID, stagedURL); setErr != nil {
						log.Error(ctx, fmt.Sprintf("Failed to set image %s ready: %v", payload.ImageID, setErr))
					}
					// Publish 'ready' status
					if pub != nil {
						_ = pub.PublishJobUpdate(ctx, events.JobUpdateEvent{JobID: job.ID, ImageID: payload.ImageID, Status: "ready"})
					}
				}
			}
		}
	}()

	log.Info(ctx, "Worker started. Press Ctrl+C to stop.")
	<-ctx.Done()
	log.Info(ctx, "Worker stopped.")
}
