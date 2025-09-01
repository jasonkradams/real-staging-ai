package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/virtual-staging-ai/worker/internal/processor"
	"github.com/virtual-staging-ai/worker/internal/queue"
	"github.com/virtual-staging-ai/worker/internal/telemetry"
)

func main() {
	ctx := context.Background()

	// Initialize OpenTelemetry
	shutdown, err := telemetry.InitTracing(ctx, "virtual-staging-worker")
	if err != nil {
		log.Fatal("Failed to initialize tracing:", err)
	}
	defer func() {
		if err := shutdown(ctx); err != nil {
			log.Printf("Failed to shutdown tracing: %v", err)
		}
	}()

	log.Println("Starting Virtual Staging AI Worker...")

	// Create context that listens for the interrupt signal from the OS
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Initialize the job processor
	processor := processor.NewImageProcessor()

	// Initialize the queue client (Redis/asynq in production)
	queueClient := queue.NewMockQueueClient()

	// Start processing jobs
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("Shutting down worker...")
				return
			default:
				// Poll for jobs
				job, err := queueClient.GetNextJob(ctx)
				if err != nil {
					log.Printf("Error getting next job: %v", err)
					time.Sleep(5 * time.Second)
					continue
				}

				if job == nil {
					// No jobs available, wait a bit
					time.Sleep(2 * time.Second)
					continue
				}

				// Process the job
				log.Printf("Processing job %s of type %s", job.ID, job.Type)
				if err := processor.ProcessJob(ctx, job); err != nil {
					log.Printf("Error processing job %s: %v", job.ID, err)
					if markErr := queueClient.MarkJobFailed(ctx, job.ID, err.Error()); markErr != nil {
						log.Printf("Failed to mark job %s as failed: %v", job.ID, markErr)
					}
				} else {
					log.Printf("Successfully processed job %s", job.ID)
					if markErr := queueClient.MarkJobCompleted(ctx, job.ID); markErr != nil {
						log.Printf("Failed to mark job %s as completed: %v", job.ID, markErr)
					}
				}
			}
		}
	}()

	log.Println("Worker started. Press Ctrl+C to stop.")
	<-ctx.Done()
	log.Println("Worker stopped.")
}
