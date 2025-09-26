# Phase 1 (P1) Checklist

This document tracks post-P0 tasks to polish DevEx, Observability, and Docs. Items can be re-scoped as we converge on frontend needs.

## Completed

- [x] GitHub Pages workflow to publish API docs from `apps/api/web/api/v1/`
- [x] OpenTelemetry spans for queue enqueuer, SSE streaming, worker events publisher, and worker processor

## In Progress

- [ ] Add structured logging around queue enqueue → worker process → DB updates → SSE publish
- [ ] Add log correlation fields (trace_id, span_id) to logs in API and worker
- [ ] Document local OTEL setup and collector config (`docs/observability.md`)

## Next

- [ ] E2E happy path (optional env-gated) documentation and CI toggle
- [ ] API Docs publishing: add link to README and repo description
- [ ] Frontend bootstrap (Next.js + Auth0):
  - [ ] App scaffold under `apps/web`
  - [ ] Auth0 login flow and token storage
  - [ ] Pages: Dashboard, Projects, Upload, Images (with SSE client)
- [ ] Security polish:
  - [ ] Document `STRIPE_WEBHOOK_SECRET` rotation steps (added)
  - [ ] Review auth scopes for protected routes
- [ ] CI enhancements:
  - [ ] Lint and unit tests matrix for api/worker
  - [ ] Optional integration tests on labels or nightly

