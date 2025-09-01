package main

import (
	"context"
	"log"

	"github.com/virtual-staging-ai/api/internal/http"
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

	s := http.NewServer(db)
	s.Start(":8080")
}
