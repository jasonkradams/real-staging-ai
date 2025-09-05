# Architecture

This document provides a high-level overview of the system architecture for the Virtual Staging AI project.

## System Components

The system is composed of the following services:

-   **API Service (`api`):** A Go service that provides the main RESTful API for the application. It handles user authentication, project management, and file uploads.
-   **Worker Service (`worker`):** A Go service that processes background jobs, such as image processing and rendering.
-   **PostgreSQL Database (`postgres`):** The primary database for storing user data, projects, and other application state.
-   **Redis (`redis`):** Used for caching and as a message broker for the job queue between the API and worker services.
-   **MinIO (`minio`):** An S3-compatible object storage service used for storing user-uploaded images and rendered outputs.
-   **OpenTelemetry Collector (`otel`):** Collects and exports telemetry data (metrics, traces, logs) for monitoring and observability.

## Component Interaction

A typical workflow looks like this:

1.  A user interacts with the web application (not yet implemented), which communicates with the **API service**.
2.  The user authenticates with Auth0, and the API service validates the JWT token.
3.  The user creates a project and uploads an image. The API service stores the project data in the **PostgreSQL database** and gets a presigned URL from the **MinIO** service to upload the image.
4.  The API service creates a new job and puts it in the **Redis** queue.
5.  The **Worker service** picks up the job from the Redis queue.
6.  The worker processes the image, which may involve downloading the image from **MinIO**, performing some computation, and uploading the result back to **MinIO**.
7.  Throughout this process, both the API and worker services send telemetry data to the **OpenTelemetry Collector**.

This architecture is designed to be scalable and resilient. The separation of the API and worker services allows for independent scaling of the components based on the workload.
