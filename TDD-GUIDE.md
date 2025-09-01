> Example: we’ll add tests before code. Below are minimal skeletons to start implementing.

### API Handler Tests (Echo)
```go
// apps/api/handlers/uploads_presign_test.go
func TestPresignUpload_RequiresAuth(t *testing.T) { /* expect 401 */ }
func TestPresignUpload_Succeeds(t *testing.T) { /* mock S3 presigner; expect 200 json */ }
```

```go
// apps/api/handlers/images_test.go
func TestCreateImageJob_EnqueuesAndPersists(t *testing.T) {}
func TestGetImage_ReturnsStatusFlow(t *testing.T) {}
```

### Worker Tests
```go
// apps/worker/worker_test.go
func TestStageRun_CreatesPlaceholderAndUpdatesDB(t *testing.T) {}
```

### Integration (happy path)
```go
// apps/api/integration/integration_test.go
// 1) create user, project
// 2) presign upload (simulate PUT to minio)
// 3) create image job
// 4) run worker once
// 5) get image -> ready with staged_url
```
