# Server-Sent Events (SSE)

This document explains how Virtual Staging AI streams realtime image job updates to clients using Server‑Sent Events (SSE).

The API exposes a per-image event stream that emits:
- An initial "connected" event
- Periodic "heartbeat" events
- Minimal "job_update" events containing status-only payloads

The worker publishes status-only updates to Redis Pub/Sub on a per-image channel, and the API relays those updates over SSE to the client.

---

## Endpoint

GET /api/v1/events?image_id={IMAGE_ID}

- Query params:
  - image_id (required) — The image identifier you want to subscribe to.
- Auth:
  - Production: This endpoint is protected; a valid JWT is required (see Auth documentation). The test server may expose it publicly for test convenience.
- Success: HTTP 200 with Content-Type: text/event-stream
- Error responses:
  - 400 Bad Request — missing image_id
  - 503 Service Unavailable — Pub/Sub not configured (e.g., Redis unavailable or misconfigured)

Typical response headers
- Content-Type: text/event-stream
- Cache-Control: no-cache
- Connection: keep-alive
- Access-Control-Allow-Origin: *
- Access-Control-Allow-Headers: Cache-Control

---

## Event types and payloads

Events are formatted per the SSE wire protocol:

- Event name: a line like event: <name>
- Data: a line like data: <json>
- Blank line separating events

1) connected
- Emitted when the subscription is established.
- Example:
  event: connected
  data: {"message":"Connected to image stream"}

2) heartbeat
- Emitted periodically to keep connections healthy and to signal liveness.
- Example:
  event: heartbeat
  data: {"timestamp": 1700000000}

3) job_update
- Emitted when a new status-only update is published for the image.
- Minimal payload shape:
  {"status":"processing" | "ready" | "error"}
- Example:
  event: job_update
  data: {"status":"processing"}

Notes
- Malformed inbound pub/sub messages are ignored to keep the stream healthy.
- When the client disconnects (context canceled), the stream ends gracefully.

---

## Pub/Sub topology

- Transport: Redis Pub/Sub
- Channel convention (per-image):
  jobs:image:{IMAGE_ID}

- Payloads: minimal status-only JSON
  {"status":"processing" | "ready" | "error"}

- Producer:
  - The Worker publishes status updates on the per-image channel as it processes the job (processing → ready | error).

- Consumer:
  - The API subscribes to the per-image channel and forwards messages to the SSE client.

---

## Client usage examples

JavaScript (EventSource)
```js
const imageId = "abc123";
const url = `/api/v1/events?image_id=${encodeURIComponent(imageId)}`;

const es = new EventSource(url, { withCredentials: true }); // include cookies if needed

es.addEventListener("connected", (e) => {
  console.log("SSE connected:", e.data);
});

es.addEventListener("heartbeat", (e) => {
  const data = JSON.parse(e.data);
  // Optionally update UI with last-seen heartbeat timestamp
});

es.addEventListener("job_update", (e) => {
  const { status } = JSON.parse(e.data);
  // Update UI: processing | ready | error
});

es.onerror = (err) => {
  console.warn("SSE error:", err);
  // The browser will auto-reconnect; consider backoff and retry UX if needed.
};
```

cURL (for quick inspection)
```bash
curl -N -H "Accept: text/event-stream" \
     -H "Authorization: Bearer <YOUR_JWT>" \
     "https://your-api.example.com/api/v1/events?image_id=abc123"
```

---

## Configuration

Environment variables
- REDIS_ADDR (required): Redis address, e.g., localhost:6379
  - If missing or Redis is unavailable, the SSE endpoint returns 503 Service Unavailable.

Runtime tuning (server-side)
- HeartbeatInterval (default: 30s): interval for heartbeat events.
- SubscribeTimeout (default: inherited from request/handler; configured in server wiring): time to wait when establishing the Redis subscription before failing.

Notes
- Current implementation reads REDIS_ADDR from environment and constructs a Redis client for Pub/Sub. Heartbeat and subscribe timeout are set via code-level configuration.
- If you want to change heartbeat cadence or subscribe timeout globally, update the SSE Config passed in your HTTP server setup.

---

## Behavior and lifecycle

Connection setup
1) Client requests GET /api/v1/events?image_id=IMAGE_ID
2) Server:
   - Validates image_id
   - Subscribes to jobs:image:IMAGE_ID
   - Sends event: connected
   - Starts heartbeat ticker

Event loop
- On each heartbeat tick → event: heartbeat
- On each pub/sub message for the channel:
  - Parse minimal status-only JSON
  - If well-formed → event: job_update
  - If malformed → ignore (no termination)
- On subscription channel closure or client cancel → terminate stream

Close conditions
- Client closed connection (browser navigates away/refresh)
- Server context canceled (route timeout or shutdown)
- Redis subscription closed or errored (stream ends gracefully)

---

## Operational considerations

- Proxies/load balancers:
  - Ensure they support long-lived HTTP responses and do not buffer SSE. Disable response buffering and set idle timeouts high enough to cover heartbeat intervals and client reconnect behavior.
- Scaling:
  - Each client connection uses a Redis subscription for a single image channel. Ensure Redis and the API instances are provisioned for expected concurrency.
- CORS:
  - Default handler sets permissive CORS header (Access-Control-Allow-Origin: *). Adjust as needed in your API gateway or application if you want stricter policies.
- Backpressure:
  - SSE is one-way, server→client. If the client is slow, the implementation flushes after each event; the OS/socket buffers apply. Consider monitoring connection counts and error rates.

---

## Troubleshooting

- I get 503 from /events
  - Check REDIS_ADDR is set and reachable by the API.
- I get 400 missing image_id
  - Provide the image_id query param: /api/v1/events?image_id=...
- No job_update events, only connected/heartbeat
  - Ensure the Worker publishes updates on jobs:image:{IMAGE_ID} with payload like {"status":"processing"}.
  - Confirm you’re subscribing to the correct IMAGE_ID.
- Browser stops receiving after some minutes
  - Verify your reverse proxy/ingress doesn’t terminate idle connections too aggressively. Increase idle timeouts or reduce heartbeat interval.

---

## Example test flow (local)

1) Start Redis (or use in-memory test Redis in unit tests).
2) Start API with REDIS_ADDR set.
3) Connect a client to /api/v1/events?image_id=img-123 (e.g., curl).
4) Publish messages:
   - redis-cli PUBLISH jobs:image:img-123 '{"status":"processing"}'
   - redis-cli PUBLISH jobs:image:img-123 '{"status":"ready"}'
5) Observe SSE stream events: connected → heartbeat (periodic) → job_update (processing) → job_update (ready)

---

## Compatibility and references

- SSE is broadly supported by modern browsers via EventSource.
- Implementation is based on:
  - Redis Pub/Sub for transport
  - Minimal JSON payloads for status-only updates
  - Echo framework for HTTP routing

This design favors simplicity and reliability for real-time job status updates without the complexity of WebSockets where not needed.