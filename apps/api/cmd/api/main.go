package main

import (
	"github.com/virtual-staging-ai/api/internal/http"
)

// main is the entrypoint of the API server.
func main() {
	e := http.NewServer()
	e.Logger.Fatal(e.Start(":8080"))
}
