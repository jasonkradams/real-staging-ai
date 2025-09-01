# Mocks

This directory contains generated mock implementations for testing. The mocks are automatically generated using [gomock](https://github.com/uber-go/mock) based on interfaces defined in the parent storage package.

## Generating Mocks

To regenerate all mocks, run from the project root:

```bash
make generate
```

This will:
1. Generate sqlc code from SQL queries
2. Generate mocks from interfaces using `go:generate` comments

## Available Mocks

### MockProjectRepository
Generated from `storage.ProjectRepository` interface in `interfaces.go`
- Provides mock implementations for all project CRUD operations
- Use for testing handlers and services that depend on project data access

### MockS3Service  
Generated from `storage.S3Service` interface in `s3.go`
- Provides mock implementations for S3 file operations
- Use for testing upload/download functionality without AWS dependencies

## Usage Example

```go
func TestCreateProject(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    // Create mock repository
    mockRepo := mocks.NewMockProjectRepository(ctrl)

    // Set expectations
    mockRepo.EXPECT().
        CreateProject(gomock.Any(), gomock.Any(), "user123").
        Return(&project.Project{
            ID:     "project-id",
            Name:   "Test Project",
            UserID: "user123",
        }, nil).
        Times(1)

    // Test your service/handler that uses the repository
    service := NewProjectService(mockRepo)
    result, err := service.CreateProject(ctx, "Test Project", "user123")
    
    assert.NoError(t, err)
    assert.Equal(t, "Test Project", result.Name)
}
```

## Adding New Mockable Interfaces

1. Define your interface in the appropriate file (e.g., `interfaces.go`)
2. Add a `//go:generate` comment at the top of the file:
   ```go
   //go:generate go run go.uber.org/mock/mockgen@latest -source=myfile.go -destination=./mocks/my_mock.go -package=mocks
   ```
3. Run `make generate` to create the mock
4. Ensure your concrete implementations declare interface compliance:
   ```go
   var _ MyInterface = (*MyImplementation)(nil)
   ```

## Best Practices

- Always call `ctrl.Finish()` in test cleanup (use `defer`)
- Set specific expectations rather than using `gomock.Any()` when possible
- Use `Times(n)` to verify methods are called the expected number of times
- Use `DoAndReturn()` for complex mock behavior that needs to inspect arguments
- Keep mocks focused on testing business logic, not implementation details

## Dependency Injection

To effectively use mocks, your handlers and services should accept interfaces as dependencies rather than concrete types. This allows easy substitution of mocks during testing.

Example refactor:
```go
// Before: Hard to test
type Handler struct {
    db *storage.DB
}

func (h *Handler) CreateProject() {
    repo := storage.NewProjectStorage(h.db) // concrete dependency
    // ... business logic
}

// After: Easy to test with mocks
type Handler struct {
    projectRepo storage.ProjectRepository // interface dependency
}

func (h *Handler) CreateProject() {
    // ... business logic using h.projectRepo
}
```

## Troubleshooting

- **Mock not generated**: Check that the `//go:generate` comment is correctly formatted
- **Interface not found**: Ensure the source file is in the same package as the generate comment
- **Import errors**: Run `go mod tidy` after generating mocks
- **Compilation errors**: Verify that concrete implementations satisfy their interfaces

## Files in this directory

- `project_repository_mock.go` - Mock for ProjectRepository interface
- `s3_service_mock.go` - Mock for S3Service interface  
- `README.md` - This documentation file

**Note**: Never edit generated files directly. They will be overwritten on the next `make generate`.