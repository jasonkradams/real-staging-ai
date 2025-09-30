package reconcile

import "context"

//go:generate go run github.com/matryer/moq@v0.5.3 -out service_mock.go . Service

// Service defines the interface for storage reconciliation operations.
type Service interface {
	ReconcileImages(ctx context.Context, opts ReconcileOptions) (*ReconcileResult, error)
}

// ReconcileOptions configures a reconciliation run.
type ReconcileOptions struct {
	ProjectID   *string // Optional filter by project
	Status      *string // Optional filter by status
	Limit       int     // Max number of images to check
	Cursor      *string // Optional cursor for pagination
	DryRun      bool    // If true, don't apply changes
	Concurrency int     // Worker pool size for S3 checks
}

// ReconcileResult summarizes what was checked and updated.
type ReconcileResult struct {
	Checked       int              `json:"checked"`
	MissingOrig   int              `json:"missing_original"`
	MissingStaged int              `json:"missing_staged"`
	Updated       int              `json:"updated"`
	Examples      []ReconcileError `json:"examples,omitempty"` // Up to 10 example errors
	DryRun        bool             `json:"dry_run"`
}

// ReconcileError captures an example error for reporting.
type ReconcileError struct {
	ImageID string `json:"image_id"`
	Status  string `json:"status"`
	Error   string `json:"error"`
}
