# Real Staging AI — Documentation

This project is an AI-driven Real Staging platform for real estate agents and photographers. Users upload photos of empty rooms; the system returns the same photos virtually furnished. Phase 1 focuses on a production-grade backend with a mocked image “staging” step to validate the full flow end-to-end (auth → upload → job → result → billing) before GPU inference is introduced.

## Quick Start

1) Prereqs: Go 1.22+, Docker + Compose, Node 20+, Auth0, Stripe

2) Start the stack
```
make up
```

3) Get a test token (Auth0)
- See guides: `docs/guides/AUTH_TOKEN_GENERATION.md`
- Or run `make token` for a local helper, if configured

4) Typical flow
- Create a project
- Request presigned upload → PUT original image to S3/MinIO
- Create image job (`stage:run`) → worker processes → image status becomes `ready`
- Fetch image status/result or listen via SSE

Key commands: `make test`, `make test-integration`, `make lint`, `make migrate` (see Makefile for more)

## Monorepo Structure

- `apps/api` — Go HTTP API (Echo). Domain logic in `internal/<domain>`; sqlc for DB access
- `apps/worker` — Go background worker (asynq) for image jobs
- `infra/migrations` — SQL migrations (up/down)
- `web/api/v1` — OpenAPI spec (`oas3.yaml`) and reference docs (planned)
- `docs` — Architecture, configuration, guides, and roadmap (this directory)

## Architecture Overview

- API (Go + Echo): Auth (Auth0/JWT), projects/images, presigned S3 uploads, Stripe webhooks, SSE
- Worker (Go + asynq): consumes `stage:run`; Phase 1 produces a watermarked/stubbed staged image
- Postgres: users, projects, images, jobs, plans, subscriptions, invoices (see DB doc)
- Redis: queue + retries (asynq)
- S3/MinIO: object storage for originals and staged outputs
- OpenTelemetry: traces/metrics/logs via collector

See details:
- Architecture: `docs/architecture.md`
- API service: `docs/api_service.md`
- Worker service: `docs/worker_service.md`
- Database schema: `docs/database.md`

## API Overview

Authenticated via Auth0 (JWT Bearer). Core endpoints:
- Projects: `/api/v1/projects`
- Uploads (presign): `/api/v1/uploads/presign`
- Images: `/api/v1/images`, `/api/v1/images/{id}`
- Events (SSE): `/api/v1/events`
- Stripe webhooks: `/api/v1/stripe/webhook`
- Billing (added): `/api/v1/billing/subscriptions`, `/api/v1/billing/invoices`

OpenAPI: lives under `web/api/v1/oas3.yaml` and is published via GitHub Pages at https://jasonkradams.github.io/real-staging-ai/. The roadmap contains an OpenAPI sketch under `docs/roadmap/phase1/API.md`.

## Configuration & Environments

All env vars documented here: `docs/configuration.md`

Typical variables: Postgres (`PG*`), Redis (`REDIS_ADDR`), S3/MinIO (`S3_*`), Auth0 (`AUTH0_*`), Stripe (`STRIPE_*`), OTEL.

## Local Development

Step-by-step setup, commands, and test flows: `docs/local_development.md`

## Deployment

Docker Compose for simple deployments; Kubernetes recommended for scale. See `docs/deployment.md` for strategies and production variable guidance.

## Testing

- Philosophy: tests first (unit + integration)
- How-to and coverage goals: `docs/guides/TESTING.md`, `docs/guides/TDD_GUIDE.md`
- Commands: `make test`, `make test-integration`, `make test-cover`

## Security & Observability

- Security guidelines (JWT/issuer/audience, presign validation, Stripe signature + idempotency, S3 key scoping): `docs/roadmap/phase1/SECURITY.md`
- Observability plan (OTEL traces, metrics, structured logs): `docs/roadmap/phase1/OBSERVABILITY.md`

## Roadmap

Start with Phase 1 to validate flows and billing with a mocked staging worker. Detailed plans and specs live under `docs/roadmap/phase1/`:
- API & OpenAPI sketch: `docs/roadmap/phase1/API.md`
- Architecture & Services: `docs/roadmap/phase1/ARCHITECTURE.md`
- Auth: `docs/roadmap/phase1/AUTH.md`
- Queue: `docs/roadmap/phase1/QUEUE.md`
- S3: `docs/roadmap/phase1/S3.md`
- Schema: `docs/roadmap/phase1/SCHEMA.md`
- Stripe: `docs/roadmap/phase1/STRIPE.md`
- Security: `docs/roadmap/phase1/SECURITY.md`
- Observability: `docs/roadmap/phase1/OBSERVABILITY.md`
- Setup & Env: `docs/roadmap/phase1/SETUP.md`

High-level milestones and checklists: `docs/todo/TODOS.md` and `docs/todo/DOC.md`.

## Recent Review Notes

Latest deep-dive and Stripe progress summaries are captured in `docs/review/2025-09-29-analysis.md`. Highlights:
- Stripe webhook signature verification + idempotency
- Subscriptions & invoices tables and repositories
- New billing endpoints: subscriptions, invoices

## Contributor Guide

- Conventions: Conventional Commits, linting, tests required for PRs
- See `CONTRIBUTING.md` and `AGENTS.md` at repo root for commands and standards

