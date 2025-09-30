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

- [ ] Storage Reconciliation: Add integration tests with mocked S3 and DB
- [ ] Storage Reconciliation: Add unit tests for extractS3Key and parseUUID helpers
- API Docs publishing:
  - [x] Add link to README
  - [x] Update repo description with docs link (manual action: add https://jasonkradams.github.io/virtual-staging-ai/ to GitHub repo settings)
- [ ] Frontend bootstrap (Next.js + Auth0):
  - [ ] App scaffold under `apps/web`
  - [ ] Auth0 login flow and token storage
  - [ ] Pages: Dashboard, Projects, Upload, Images (with SSE client)
- [ ] Security polish:
  - [ ] Document `STRIPE_WEBHOOK_SECRET` rotation steps (added)
  - [ ] Review auth scopes for protected routes
- [x] CI enhancements:
  - [x] Lint and unit tests matrix for api/worker
  - [x] Optional integration tests on labels or nightly (workflow_dispatch + schedule)
