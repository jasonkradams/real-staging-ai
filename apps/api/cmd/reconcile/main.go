package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/real-staging-ai/api/internal/config"
	"github.com/real-staging-ai/api/internal/logging"
	"github.com/real-staging-ai/api/internal/reconcile"
	"github.com/real-staging-ai/api/internal/storage"
)

func main() {
	var (
		batchSize   = flag.Int("batch-size", 100, "Number of images to check per batch")
		concurrency = flag.Int("concurrency", 5, "Number of concurrent S3 checks")
		dryRun      = flag.Bool("dry-run", false, "Don't apply changes, only report what would be done")
		projectID   = flag.String("project-id", "", "Optional: filter by project ID")
		status      = flag.String("status", "", "Optional: filter by status (queued, processing, ready, error)")
	)
	flag.Parse()

	ctx := context.Background()
	logger := logging.Default()

	logger.Info(ctx, "starting reconciliation CLI")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Error(ctx, "failed to load configuration", "error", err)
		fmt.Fprintf(os.Stderr, "Error: failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	db, err := storage.NewDefaultDatabase(&cfg.DB)
	if err != nil {
		logger.Error(ctx, "failed to connect to database", "error", err)
		fmt.Fprintf(os.Stderr, "Error: failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	logger.Debug(ctx, "database connection established")

	logger.Debug(ctx, "initializing S3 service", "bucket", cfg.S3.BucketName)
	s3Service, err := storage.NewDefaultS3Service(ctx, &cfg.S3)
	if err != nil {
		logger.Error(ctx, "failed to initialize S3 service", "error", err, "bucket", cfg.S3.BucketName)
		fmt.Fprintf(os.Stderr, "Error: failed to initialize S3 service: %v\n", err)
		os.Exit(1)
	}

	// Create reconcile service
	svc := reconcile.NewDefaultService(db, s3Service)

	// Build options
	opts := reconcile.ReconcileOptions{
		Limit:       *batchSize,
		Concurrency: *concurrency,
		DryRun:      *dryRun,
	}
	if *projectID != "" {
		opts.ProjectID = projectID
	}
	if *status != "" {
		opts.Status = status
	}

	logger.Info(ctx, "starting reconciliation run",
		"dry_run", *dryRun,
		"batch_size", *batchSize,
		"concurrency", *concurrency,
		"project_id", opts.ProjectID,
		"status", opts.Status,
	)

	fmt.Printf("Starting reconciliation (dry_run=%v, batch_size=%d, concurrency=%d)\n", *dryRun, *batchSize, *concurrency)
	if opts.ProjectID != nil {
		fmt.Printf("  Filtering by project_id: %s\n", *opts.ProjectID)
	}
	if opts.Status != nil {
		fmt.Printf("  Filtering by status: %s\n", *opts.Status)
	}

	// Run reconciliation
	result, err := svc.ReconcileImages(ctx, opts)
	if err != nil {
		logger.Error(ctx, "reconciliation failed", "error", err)
		fmt.Fprintf(os.Stderr, "Error: reconciliation failed: %v\n", err)
		os.Exit(1)
	}

	logger.Info(ctx, "reconciliation completed",
		"checked", result.Checked,
		"missing_original", result.MissingOrig,
		"missing_staged", result.MissingStaged,
		"updated", result.Updated,
		"dry_run", result.DryRun,
	)

	// Print results
	fmt.Println("\nReconciliation Results:")
	fmt.Printf("  Checked:         %d images\n", result.Checked)
	fmt.Printf("  Missing original: %d\n", result.MissingOrig)
	fmt.Printf("  Missing staged:   %d\n", result.MissingStaged)
	fmt.Printf("  Updated:         %d\n", result.Updated)
	fmt.Printf("  Dry run:         %v\n", result.DryRun)

	if len(result.Examples) > 0 {
		fmt.Println("\nExample errors (up to 10):")
		for _, ex := range result.Examples {
			fmt.Printf("  - Image %s (status=%s): %s\n", ex.ImageID, ex.Status, ex.Error)
		}
	}

	// Output JSON for scripting
	if jsonOutput := os.Getenv("JSON_OUTPUT"); jsonOutput == "1" {
		jsonBytes, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println("\nJSON Output:")
		fmt.Println(string(jsonBytes))
	}

	if !*dryRun && result.Updated > 0 {
		fmt.Println("\nNote: Changes have been applied to the database.")
	} else if *dryRun && result.Updated > 0 {
		fmt.Println("\nNote: This was a dry run. No changes were applied.")
	}
}
