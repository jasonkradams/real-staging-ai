# Phase 1 (P1) Checklist

This document tracks post-P0 tasks to polish DevEx, Observability, and Docs. Items can be re-scoped as we converge on frontend needs.

## Completed

- [x] GitHub Pages workflow to publish API docs from `apps/api/web/api/v1/`
- [x] OpenTelemetry spans for queue enqueuer, SSE streaming, worker events publisher, and worker processor
- [x] Add structured logging around queue enqueue → worker process → DB updates → SSE publish
- [x] Add log correlation fields (trace_id, span_id) to logs in API and worker
- [x] Document local OTEL setup and collector config (`docs/roadmap/phase1/OBSERVABILITY.md`)
- [x] E2E happy path (optional env-gated) tests and CI toggle implemented
- [x] Storage Reconciliation: Module, service, SQL queries, admin endpoint, and CLI
- [x] Storage Reconciliation: Operations runbook (`docs/operations/reconciliation.md`)
- [x] Storage Reconciliation: Makefile target `reconcile-images`
- [x] CI matrix for api/worker (test and lint jobs)
## In Progress

## Next

- API Docs publishing:
  - [x] Add link to README
  - [x] Update repo description with docs link (manual action: add https://jasonkradams.github.io/real-staging-ai/ to GitHub repo settings)
- Frontend bootstrap (Phase 1):
  - [x] App scaffold under `apps/web` (Next.js 15 + TypeScript + Tailwind)
  - [x] Pages: Dashboard (`/`), Upload (`/upload`), Images (`/images`)
  - [x] Project creation and selection UI
  - [x] Image upload flow (presign → S3 → create image record)
  - [x] SSE client for real-time job updates
  - [x] Image listing with status badges and presigned URL viewing
  - [x] Rudimentary auth (manual token paste via localStorage)
  - [x] API client library with bearer token auth (`lib/api.ts`)
- [x] CI enhancements:
  - [x] Lint and unit tests matrix for api/worker
  - [x] Optional integration tests on labels or nightly

## Phase 2 (Complete) / Phase 3 (In Progress)

- [x] Storage Reconciliation: Add integration tests with mocked S3 and DB
- [x] Fix dry-run bug in reconciliation service (Updated counter was incorrectly incremented)
- [x] Security polish:
  - [x] Document `STRIPE_WEBHOOK_SECRET` rotation steps (`docs/security/STRIPE_WEBHOOK_ROTATION.md`)
- [x] Auth0 SDK integration for proper OAuth flow
  - [x] Install `@auth0/nextjs-auth0` SDK
  - [x] Configure Auth0 client and middleware
  - [x] Login/logout with Auth0 Universal Login
  - [x] Protected routes with automatic redirect
  - [x] Session management with encrypted cookies
  - [x] Token refresh via SDK
  - [x] Update API client to use Auth0 access tokens
  - [x] Update SSE viewer to use Auth0 access tokens
  - [x] Documentation (`docs/frontend/AUTH0_INTEGRATION.md`)
- [ ] Additional security improvements:
  - [ ] Review auth scopes for protected routes
  - [ ] Add CSRF protection
  - [ ] Document general secrets management procedures
