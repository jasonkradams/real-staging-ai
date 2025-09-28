# Virtual Staging AI — Web (Next.js)

A minimal Next.js 14 + Tailwind web app for Phase 1:
- Upload an image via presign → PUT to S3
- Create an image job against the API
- Watch job status via Server-Sent Events (SSE)

## Prereqs
- Node.js 18.18+ (recommended: 18.x LTS)
- API running locally at http://localhost:8080 (via `make up` or your preferred workflow)

## Getting Started

```bash
# From repo root
cd apps/web
npm install
npm run dev
```

Open http://localhost:3000

## API Proxy
Requests to `/api/*` are proxied to `http://localhost:8080/api/*` via `next.config.js` rewrites.

You can override with env:

```
# .env.local
NEXT_PUBLIC_API_BASE=/api
```

## Authentication
For local testing, paste a bearer token via the Token bar (top-right). The token is stored in `localStorage` and sent as `Authorization: Bearer <token>`.

You can generate a token via:

```bash
make token
```

## Pages
- `/` — landing
- `/upload` — presign + PUT upload + create image
- `/images` — SSE viewer: connect to `/api/v1/events?image_id=<id>` and stream `job_update` events

## Styling
Tailwind CSS is configured via `tailwind.config.ts` and `app/globals.css`.

## Scripts
- `npm run dev` — Next.js dev server on :3000
- `npm run build` — production build
- `npm start` — start production server
- `npm run lint` — Next.js ESLint

## Notes
- This app is intentionally minimal to validate backend flows. Auth0 login/UI will be added in a later milestone.
