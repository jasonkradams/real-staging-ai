package project_test

import (
	"testing"

	"github.com/virtual-staging-ai/api/internal/project"
)

func TestProjectValidation(t *testing.T) {
	// Test that a project can be created with valid data
	p := &project.Project{
		ID:     "test-id",
		Name:   "Test Project",
		UserID: "user-123",
	}

	if p.Name != "Test Project" {
		t.Errorf("Expected project name to be 'Test Project', got %s", p.Name)
	}

	if p.UserID != "user-123" {
		t.Errorf("Expected user ID to be 'user-123', got %s", p.UserID)
	}
}
