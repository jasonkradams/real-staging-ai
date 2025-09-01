### Overview
A DIY virtual staging SaaS for real-estate photos. Upload a room photo → receive a *mocked* staged image in Phase 1 (no GPU yet). We’ll validate flows, pricing, and UX while keeping the image pipeline simple.

### Tech Summary
- **Backend/API**: Go (Echo), Postgres (pgx), Redis (asynq), S3, Stripe, Auth0 (OIDC/JWT), OpenTelemetry
- **Frontend**: Next.js + Tailwind + shadcn/ui
- **Infra**: Docker Compose (dev), GitHub Actions (CI), Fly.io/Render/Neon/Supabase/Cloudflare R2 (later)

### Phase 1 Goals
- End-to-end flow working: auth → upload → job → result placeholder → billing → usage tracking
- **Test-driven** from the start (unit + integration)
- Image generation is **stubbed**: returns a watermarked placeholder based on the uploaded image to exercise data flow, S3, and queueing.

### High-Level Flow
1. User logs in via Auth0 → frontend gets JWT
2. Upload original image via **S3 presigned PUT** from API
3. Create an **image job** → enqueued to Redis (asynq)
4. Worker processes job (Phase 1: copy original to a `-staged.jpg` variant + watermark)
5. API marks image **ready** → client fetches results / receives event updates
