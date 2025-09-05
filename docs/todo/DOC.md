# Documentation Todos

This document outlines the documentation tasks for the Virtual Staging AI project.

## 1. Technical Specifications (in `docs/`)

-   [x] `docs/architecture.md`: High-level overview of the system architecture, including the `api` and `worker` services, the database, Redis, and S3.
-   [x] `docs/api_service.md`: Detailed documentation of the API service.
    -   [x] Authentication flow with Auth0.
    -   [x] Description of each HTTP handler and its purpose.
    -   [x] Database interaction patterns.
    -   [x] S3 usage for file uploads.
    -   [x] OpenTelemetry integration and tracing strategy.
    -   [x] Data Models: Detailed description of the Go structs and their validation rules.
    -   [x] Error Handling: Explanation of the error handling strategy.
-   [x] `docs/worker_service.md`: Detailed documentation of the worker service.
    -   [x] Redis queue implementation.
    -   [x] Job processing logic.
    -   [x] Telemetry and monitoring for the worker.
    -   [x] Job Payloads: Description of the different job types and their JSON payloads.
-   [x] `docs/database.md`: Documentation of the database schema.
    -   [x] Description of each table and its columns.
    -   [x] Entity-relationship diagram (ERD).

## 2. API Documentation (`web/api/v1/`)

-   [x] Review and complete the OpenAPI specification in `web/api/v1/oas3.yaml`.
-   [x] Add detailed descriptions for all API endpoints.
-   [x] Provide clear request and response examples for every endpoint.
    -   [x] Success cases (2xx).
    -   [x] Error cases (4xx, 5xx).

## 3. Additional Documentation

-   [x] `docs/local_development.md`: A guide on how to set up the project for local development.
-   [x] `docs/deployment.md`: A guide on how to deploy the application to a production environment.
-   [x] `CONTRIBUTING.md`: A guide for new contributors.
-   [x] `docs/configuration.md`: A document explaining all the environment variables used in the project.