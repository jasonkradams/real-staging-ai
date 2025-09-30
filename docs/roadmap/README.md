# Roadmap

This roadmap outlines Phase 1 for Virtual Staging AI: validate core flows end-to-end with a mocked staging worker, solid auth, billing, and ops foundations. Later phases will introduce true AI staging and a full-featured web UI.

## Phase 1 — Goals

- Authenticated API with project/image flows
- Direct-to-S3 uploads via presigned URLs
- Background image jobs (`stage:run`) processed by worker
- Result retrieval via polling and SSE
- Billing foundation with Stripe webhooks, subscriptions, invoices
- Security baselines (JWT validation, S3 key scoping, webhook signatures)
- Observability (OTEL traces/metrics/logs)

## Specs & Guides (Phase 1)

- API & OpenAPI sketch: `API.md`
- Architecture overview: `ARCHITECTURE.md`
- Auth: `AUTH.md`
- Queue/tasks: `QUEUE.md`
- S3 strategy: `S3.md`
- Schema/migrations: `SCHEMA.md`
- Stripe integration plan: `STRIPE.md`
- Security checklist: `SECURITY.md`
- Observability checklist: `OBSERVABILITY.md`
- Setup & environment: `SETUP.md`

## Milestones & Tasks

Use these living lists for progress tracking and next steps:
- Engineering milestones and checklists: `../../todo/TODOS.md`
- Documentation tasks: `../../todo/DOC.md`
- Latest deep-dive review and Stripe progress: `../review/2025-09-29-analysis.md`

## Next Phases (High-Level)

- Phase 2: Real image generation (GPU-backed), scalable storage/CDN, stronger quotas — see `phase2/ROLLOUT.md`
- Phase 3: Rich web UI (upload, project management, checkout), collaboration, audit and usage analytics
