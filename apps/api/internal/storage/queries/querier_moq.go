package queries

// Generate a moq for the Querier interface in this package so it's reusable across other packages.
//
// Usage:
//   go generate ./apps/api/internal/storage/queries
//
//go:generate go run github.com/matryer/moq@v0.5.3 -out querier_mock.go . Querier
