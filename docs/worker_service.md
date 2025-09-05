# Worker Service

This document provides a detailed documentation of the worker service.

## Redis Queue

The worker service uses Redis as a message broker for a job queue. When the API service needs to perform a long-running task, such as image processing, it enqueues a job in Redis.

The job queue is implemented using Redis Lists, which provide a simple and reliable way to implement a FIFO queue.

## Job Processing

The worker service continuously polls the Redis queue for new jobs. When a new job is received, the worker performs the following steps:

1.  Deserializes the job payload.
2.  Performs the job's task (e.g., image processing).
3.  Updates the job status in the database.
4.  Sends a notification to the user (e.g., via Server-Sent Events).

The worker is designed to be resilient to failures. If a job fails, it can be retried or marked as failed in the database.

## Telemetry and Monitoring

Like the API service, the worker service is instrumented with OpenTelemetry to provide tracing and metrics.

Traces are used to track jobs as they are processed by the worker, providing visibility into the performance and potential bottlenecks of the system.

Metrics are used to monitor the health and performance of the worker, such as the number of jobs processed, the number of failed jobs, and the average job processing time.

## Job Payloads

This section describes the different job types and their JSON payloads.

### `stage:run`

This job type is used to process an image and generate a staged version.

**Payload:**

```json
{
  "image_id": "a1b2c3d4-e5f6-7890-1234-567890abcdef",
  "original_url": "https://bucket.s3.amazonaws.com/uploads/uuid/original.jpg",
  "room_type": "living_room",
  "style": "modern",
  "seed": 1234567890
}
```

| Field | Type | Description |
| --- | --- | --- |
| `image_id` | UUID | The ID of the image to be processed. |
| `original_url` | string | The URL of the original uploaded image. |
| `room_type` | string | The type of the room in the image. |
| `style` | string | The staging style. |
| `seed` | integer | The seed for the staging process. |
