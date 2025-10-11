# Observability (Phase 1)

This guide describes how to use tracing and structured logging locally for the API and Worker.

## Traces

- OTEL Collector runs via Docker Compose as service `otel`. Config: `infra/otelcol.yaml` (debug exporter).
- Default endpoint for SDKs is `http://otel:4318` in Docker and `http://localhost:4318` from the host.
- Environment variable used by both services: `OTEL_EXPORTER_OTLP_ENDPOINT`.

Local run with Docker Compose automatically wires the endpoint:

```
docker compose up -d api worker otel
```

Viewing traces (debug exporter):

```
docker compose logs -f otel
```

We currently trace key paths:

- API: queue enqueue, SSE streaming
- Worker: event publisher, processor

## Logs

We use Go `slog` for structured JSON logs across API and Worker and include OTEL trace correlation when available.

- Environment: `LOG_LEVEL` controls verbosity. Supported values: `debug`, `info` (default), `warn`, `error`.
- Output: JSON to stdout with stable keys including `time`, `level`, `msg`, `service`, optional `trace_id`, `span_id`.

Example output:

```
{"time":"2025-09-26T14:08:45-07:00","level":"WARN","msg":"events publish failed","service":"real-staging-worker","image_id":"img-123","status":"processing","attempt":2,"error":"dial tcp 127.0.0.1:6390: connect: connection refused","trace_id":"...","span_id":"..."}
```

## Metrics (P2)

- Planned: request counts, queue depth, job latency, image sizes.

## End-to-End Testing Aids

- Integration tests run via `make test-integration` and start dockerized Postgres/Redis/LocalStack.
- Full upload → ready SSE flow (optional) is gated by `RUN_E2E_UPLOAD_READY=1`:

```
RUN_E2E_UPLOAD_READY=1 make test-integration
```
