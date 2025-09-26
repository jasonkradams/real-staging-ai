# Project Todos

This document outlines the systematic plan and checklist to implement the Virtual Staging AI project.

## Milestone 1: Project Setup & Foundation

- [x] Set up monorepo layout (`/apps`, `/infra`, `/docs`)
- [x] Define project architecture and services (`API`, `Worker`, `Postgres`, `Redis`, `S3`)
- [x] Create initial `docker-compose.yml` for development environment (Postgres, Redis, MinIO)
- [x] Define environment variables and setup instructions in `SETUP.md`
- [x] Define contribution guidelines in `CONTRIBUTING.md`
- [x] Set up `Makefile` with `test` and `test-integration` targets
- [x] Set up `docker-compose.test.yml` for integration tests
- [x] Implement golangci-lint with v2 standards and make target
- [x] Implement database migrations for the initial schema
- [x] Create seed data fixtures for the test database
- [x] Prepare "golden files" for image comparisons

## Milestone 2: API Documentation

- [x] Create `web/api/v1/oas3.yaml` with the OpenAPI specification.
- [x] Update all API endpoints to be prefixed with `/api/v1`.
- [x] Serve the API documentation at the `/api/v1/docs` endpoint.
- [x] Create a `make docs` target to validate the OAS3 file.
- [ ] Configure GitHub Pages to publish the API documentation.

## Milestone 3: Core Backend Features

### 3.1. API Server (Go + Echo)

- [x] Implement basic project CRUD endpoints (`GET /api/v1/projects`, `POST /api/v1/projects`, `GET /api/v1/projects/{id}`, `PUT /api/v1/projects/{id}`, `DELETE /api/v1/projects/{id}`)
- [x] Implement API for presigned S3 uploads (`POST /api/v1/uploads/presign`)
- [x] Implement API for creating image jobs (`POST /api/v1/images`)
- [x] Implement API for fetching image status (`GET /api/v1/images/{id}`)
- [x] Implement Server-Sent Events (SSE) for real-time job updates (`GET /api/v1/events`)

### 3.2. Authentication (Auth0)

- [x] Set up Auth0 API and SPA application
- [x] Implement JWT validation middleware in the API (Echo middleware)

### 3.3. Database & Schema (Postgres + pgx + sqlc)

- [x] Create initial database schema with `users`, `projects`, `images`, `jobs`, and `plans` tables
- [x] Set up `golang-migrate` for managing database migrations
- [x] Set up `sqlc` to generate type-safe queries from SQL

### 3.4. Background Jobs (Worker + asynq)

- [x] Implement the `stage:run` task in the worker
- [x] Replace mock queue with Redis/asynq in API and worker; define task payload contract
- [x] API enqueues `stage:run` to Redis/asynq on image creation
- [x] Worker updates images table on start/success/failure; idempotent transitions
- [x] Wire Redis pub/sub for SSE and stream updates in API (define channels/payload schema, ensure API subscribes and emits, add heartbeats)


### 3.5. File Uploads (S3)

- [x] Implement the logic to generate presigned PUT URLs for direct browser uploads

### 3.6. Billing (Stripe)

- [x] Implement the Stripe webhook handler (`POST /api/v1/stripe/webhook`)

### 3.7. Observability

- [x] Set up OpenTelemetry collector in `docker-compose.yml`
- [x] Instrument the API server and other components with OpenTelemetry

## Milestone 4: Frontend Implementation (Next.js)

- [ ] Set up a new Next.js application in `/apps/web`
- [ ] Implement user authentication with Auth0
- [ ] Implement the image upload flow
- [ ] Implement a dashboard to view projects and images
- [ ] Implement the checkout flow with Stripe
- [ ] Implement real-time updates for image status

## Milestone 5: Testing

- [ ] **Achieve 100% Coverage on all packages**
- [ ] **Authentication Middleware**
    - [ ] `fail: no JWT`
    - [ ] `fail: invalid JWT`
    - [ ] `success: valid JWT`
- [ ] **Presigned Upload Endpoint**
    - [ ] `fail: requires auth`
    - [ ] `success: returns presigned URL`
- [ ] **Image Job Endpoint**
    - [x] `success: enqueues and persists`
    - [x] `success: returns status flow`
- [ ] **Background Worker**
    - [x] `success: creates placeholder and updates DB`
- [ ] **Stripe Webhook Endpoint**
    - [ ] `success: handles checkout session`
- [ ] **Server-Sent Events (SSE) Endpoint**
    - [x] `success: streams job updates`
- [ ] **End-to-End Integration Test**
    - [ ] `success: happy path`

## Milestone 6: Deployment

- [ ] Create a preview environment for the application
- [ ] Deploy the application to the preview environment
- [ ] Set up CI/CD pipeline with GitHub Actions

## Milestone 7: Refactor
- [ ] Refactor http handlers and break out handlers into their respective packages
