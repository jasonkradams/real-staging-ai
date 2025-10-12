# Phase 2 Rollout Plan — AI Inference

This plan introduces real GPU-backed virtual staging while preserving reliability, cost control, and product velocity. Phase 2 replaces the Phase 1 mocked staging with a production inference service and controlled rollout.

## Objectives

- Deliver real, high-quality virtual staging results for common room types and styles.
- Keep the system reliable and observable under real load (SLOs defined below).
- Control cost via concurrency, autoscaling, and feature gating by plan tiers.
- Preserve user experience: same API, predictable status transitions, clear errors.

## Scope (In / Out)

- In: Inference microservice, worker integration, prompts/style mapping, quality evaluation loop, observability, security, autoscaling, cost controls.
- Out: Full frontend build-out, advanced manual editing tools, multi-image scenes, 3D staging.

## Architecture Changes (Delta from Phase 1)

- Add `apps/inference` (Python): HTTP or gRPC service running Diffusers/PyTorch on GPU nodes.
  - Reads original from S3, runs image-to-image/inpainting pipeline, writes staged image to S3.
  - Exposes `/healthz`, `/readyz`, `/infer` (or gRPC `Infer`), `/metrics` (Prometheus), OTEL tracing.
- Extend Worker: call inference service instead of stubbed “copy”.
  - Request: `{ image_id, original_key, room_type, style, seed, strength?, guidance_scale?, mask_key? }`
  - Response: `{ staged_key, width, height, model_rev, duration_ms, warnings? }`
  - Timeouts, retries with backoff, idempotency by `image_id`.
- Optional schema additions (planned, can be deferred initially):
  - `images.inference_meta JSONB` (model, params, timings)
  - `jobs.metrics JSONB` (duration breakdowns)

## Inference Pipeline (Initial target)

- Base: SDXL image-to-image with ControlNet (depth or edge) to preserve room geometry.
- Mode: inpainting over furniture regions; conservative denoise strength to maintain structure.
- Prompts: template per `(room_type, style)`, with negative prompts for artifacts.
- Resolution: 768–1024 px longest side (configurable). Optional upscaler (paid tiers).
- Reproducibility: deterministic seeds when provided; otherwise random.

## Deployment Strategies

Option A — Managed GPU (easiest first): Replicate/Modal/RunPod API
- Pros: speed to market, zero cluster ops, simple autoscaling.
- Cons: vendor lock-in, cold starts, cost per run.

Option B — Self-hosted GPUs (Kubernetes)
- GPU node pool (L4/A10/A100), NVIDIA drivers, device plugin.
- Images pre-loaded with models; HPA/KEDA scaling by queue depth + GPU utilization.
- Pros: cost control at scale, no vendor dependency; Cons: ops complexity.

Decision checkpoint after M2: select A or B for GA; keep the other as fallback.

## Milestones & Timeline (Indicative)

M0 — Design & Spikes (3–5 days)
- Evaluate model recipes (SDXL img2img + ControlNet depth/edges). Collect sample inputs/outputs.
- Benchmark latency on target hardware (L4/A10) at 768/1024 px.
- Decide initial provider path (A: managed) for faster iteration.

M1 — Contracts & Scaffolding (3–4 days)
- Define worker ↔ inference API contract and error model.
- Scaffold `apps/inference` with health/metrics/trace; local CPU fallback for dev.
- Feature flag in worker: `INFERENCE_MODE=mock|managed|self` with kill switch.

M2 — Managed GPU Integration (4–6 days)
- Implement provider adapter (auth, timeouts, retries, idempotency key = `image_id`).
- Wire S3 I/O (download original, upload staged) with least-privilege keys.
- Add observability: end-to-end spans; metrics for P50/P95, failure rates, queue waits.

M3 — Quality Tuning & Acceptance (1–2 weeks parallelized)
- Prompt templates per room/style; negative prompts; seed behavior.
- Try ControlNet variants (depth vs. edge) and denoise strengths.
- Evaluate outputs with a rubric: structure fidelity, furniture plausibility, style adherence, artifacts.
- Acceptance gate: ≥80% pass on curated test set; define rework plan for fails.

M4 — Autoscaling & Cost Controls (3–5 days)
- Backpressure via asynq priorities and queue depth caps.
- Concurrency controls in inference; guard VRAM to prevent OOM.
- Scale policy: autoscale on queue depth and P95 latency; scale-to-zero cooldown.

M5 — Security, Compliance, and Safety (3–4 days)
- Validate/limit input sizes and mime types; strip EXIF.
- Content constraints policy (e.g., reject faces/people if required); log redactions.
- License review for models/weights; document versions.

M6 — Gradual Rollout (5–7 days)
- Canary: route 5–10% of jobs to real inference; rest to mock.
- SLOs: availability ≥99.5%, P95 time ≤ 60s at 1024px; failure rate ≤ 2%.
- Monitor, iterate prompts and parameters, raise canary to 100%.

M7 — GA & Tiering (3–4 days)
- Enable for paid tiers; add optional upscaler/refiner for Pro/Business.
- Document user-facing guidance; update API docs and examples.

## Detailed Work Items

Contracts & APIs
- Define JSON schema for request/response and error codes (timeout, OOM, invalid input, transient provider error).
- Idempotency: treat duplicate `image_id` as safe replays; return previously produced key.
- Cancellation: optional future support via job cancel endpoint.

Worker Integration
- HTTP/gRPC client with deadlines; exponential backoff; circuit breaker for provider outages.
- Map inference response to DB updates; persist `inference_meta` if enabled.
- Telemetry: span links from worker → inference; attributes include model, params, timings.

Inference Service
- Health probes: `/healthz` (liveness), `/readyz` (model warm), `/metrics` (Prometheus).
- Model mgmt: preload weights; control precision (fp16/bf16); enable xFormers/Flash-Attn if available.
- Concurrency: 1–2 jobs/GPU at 1024px; runtime knob via env; queue inside service off by default.
- Storage: stream from S3; write staged to S3 with content-type and caching headers.

Observability
- Metrics: queue depth, dequeue latency, inference duration, GPU utilization/VRAM, OOM count, error rates.
- Tracing: spans for S3 GET/PUT, model forward pass, preprocessing, upscaler.
- Logs: structured JSON; include `image_id`, model rev, seed; redact prompts if necessary.

Testing & Validation
- Unit tests: request/response validation, timeouts, retries.
- Integration: golden inputs; deterministic seeds; snapshot outputs checksum for regression.
- Chaos: inject provider 5xx/timeouts; ensure graceful degradation to mock if kill switch flips.

Security & Privacy
- Secrets via env/Secrets Manager; never commit tokens/keys.
- S3 paths scoped per user/project; enforce allowed prefixes.
- Validate and cap resolution; reject oversized payloads early.

Cost & Capacity Planning
- Track cost per image by tier/provider; alert on threshold breaches.
- Autoscale policy tuned to maintain P95 SLO; prefer spot instances if self-hosting.
- Throttle free tier; queue length caps; user-level rate limits if necessary.

Rollout & Rollback
- Feature flags: `INFERENCE_ENABLED`, `INFERENCE_MODE`, `UPSCALE_ENABLED`.
- Canary: start low %, expand as KPIs hold; publish runbook.
- Rollback: flip to `mock` mode instantly; drain GPU jobs; preserve queued requests.

Risks & Mitigations
- GPU cold starts → warm pools; preloading; keep-alive.
- OOM at high res → cap resolution; adapt batch size; reduce denoise.
- Output quality variance → prompt tuning, LoRAs; fallback to alternative ControlNet.
- Provider outage → circuit breaker; automatic fallback to mock; status page.

SLOs & KPIs (initial)
- Availability (inference endpoint): ≥ 99.5%
- P50 inference time: ≤ 20s @ 768px; P95 ≤ 60s @ 1024px
- Failure rate: ≤ 2% (excludes invalid input)
- Queue wait P95: ≤ 15s under typical load

Definition of Done
- All M1–M6 completed; SLOs met during canary for 7 consecutive days.
- Quality rubric pass rate ≥ 80% on curated set; no critical regressions.
- Docs updated (API usage, limits, styles); runbooks in place.

