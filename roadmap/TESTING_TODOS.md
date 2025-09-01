### Phase 1 Testing Todo List

**1. Setup Testing Infrastructure**
- [x] Create a `Makefile` with `test` (unit) and `test-integration` targets.
- [x] Configure a `docker-compose.test.yml` for a dedicated test database and other services.
- [x] Implement a mechanism to run database migrations as part of the test setup.
- [x] Create seed data fixtures for users, projects, etc., to use in tests.
- [x] Prepare "golden files" to represent the minimal placeholder images for byte-level comparisons.

**2. Authentication Middleware**
- [ ] **Test:** `fail: no JWT` - Write a test to ensure requests without a JWT are rejected with a `401 Unauthorized`.
- [ ] **Test:** `fail: invalid JWT` - Write a test to ensure requests with an invalid or expired JWT are rejected.
- [ ] **Test:** `success: valid JWT` - Write a test using a mock JWKS to confirm valid JWTs are accepted.

**3. Presigned Upload Endpoint (`/v1/uploads/presign`)**
- [ ] **Test:** `fail: requires auth` - Write `TestPresignUpload_RequiresAuth` to verify the endpoint is protected by auth.
- [ ] **Test:** `success: returns presigned URL` - Write `TestPresignUpload_Succeeds` to confirm the endpoint returns a URL and key, and correctly enforces content-type and size limits.

**4. Image Job Endpoint (`/v1/images`)**
- [ ] **Test:** `success: enqueues and persists` - Write `TestCreateImageJob_EnqueuesAndPersists` to verify that a `POST` request correctly creates database records, enqueues a background job, and returns a `202 Accepted` with the new image ID.
- [ ] **Test:** `success: returns status flow` - Write `TestGetImage_ReturnsStatusFlow` to check that a `GET` request for an image shows the status transitioning from `queued` to `processing` to `ready`.

**5. Background Worker**
- [ ] **Test:** `success: creates placeholder and updates DB` - Write `TestStageRun_CreatesPlaceholderAndUpdatesDB` to ensure the worker processes a job, downloads the original file, creates and uploads the staged version, and updates the image's status in the database.

**6. Stripe Webhook Endpoint (`/v1/stripe/webhook`)**
- [ ] **Test:** `success: handles checkout session` - Write a test to simulate a `checkout.session.completed` event and verify that the user's plan and `stripe_customer_id` are updated correctly in the database.

**7. Server-Sent Events (SSE) Endpoint (`/v1/events`)**
- [ ] **Test:** `success: streams job updates` - Write a test to connect to the events endpoint and verify that it streams status updates for a running job.

**8. End-to-End Integration Test**
- [ ] **Test:** `success: happy path` - Create a full integration test that:
    1. Creates a user and project.
    2. Simulates an image upload using a presigned URL.
    3. Creates an image job.
    4. Triggers the worker to run once.
    5. Polls the image endpoint until the status is `ready` and a `staged_url` is present.
