// Package storage provides functionality for interacting with the database.
package storage

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB represents a connection to the database.
type DB struct {
	pool *pgxpool.Pool
}

// NewDB creates a new DB instance.
func NewDB(ctx context.Context) (*DB, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://testuser:testpassword@localhost:5433/testdb?sslmode=disable"
	}

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	return &DB{pool: pool}, nil
}

// Close closes the database connection.
func (db *DB) Close() {
	db.pool.Close()
}

// GetPool returns the underlying pgxpool.Pool.
func (db *DB) GetPool() *pgxpool.Pool {
	return db.pool
}
