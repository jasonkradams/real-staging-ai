# Deployment

This document provides a guide on how to deploy the Real Staging AI application to a production environment.

## Overview

The application is designed to be deployed as a set of containerized services using Docker. The main components to be deployed are:

-   The **API service**
-   The **Worker service**
-   A **PostgreSQL database**
-   A **Redis instance**
-   An **S3-compatible object storage** service (e.g., AWS S3, MinIO)
-   An **OpenTelemetry Collector** (optional, but recommended for production)

## Building Docker Images

The `docker-compose.yml` file is configured to build the Docker images for the `api` and `worker` services automatically. To build the images manually, you can use the following commands from the root of the project:

```bash
docker build -t real-staging-api ./apps/api
docker build -t real-staging-worker ./apps/worker
```

## Deployment Strategies

### Docker Compose

For simple deployments, you can adapt the existing `docker-compose.yml` file for a production environment. This would involve:

-   Removing the development-specific configurations (e.g., local volumes for code).
-   Using a managed PostgreSQL and Redis service instead of running them in containers.
-   Configuring the environment variables for the production environment.

### Kubernetes

For a more scalable and resilient deployment, it is recommended to use Kubernetes. This would involve creating Kubernetes manifests (Deployments, Services, ConfigMaps, Secrets, etc.) for each of the services.

A detailed guide on how to deploy the application to Kubernetes is beyond the scope of this document, but the general steps would be:

1.  Create a Kubernetes cluster.
2.  Create a namespace for the application.
3.  Create a PostgreSQL database and a Redis instance (e.g., using a cloud provider's managed service or by deploying them in the cluster).
4.  Create Kubernetes Secrets for the database credentials, Auth0 credentials, and other sensitive information.
5.  Create Kubernetes ConfigMaps for the non-sensitive configuration.
6.  Create Kubernetes Deployments and Services for the `api` and `worker` services.
7.  Set up an Ingress controller to expose the `api` service to the internet.

## Production Configuration

For a production environment, you will need to configure the following environment variables:

-   **`DATABASE_URL`:** The connection string for your production PostgreSQL database.
-   **`AUTH0_DOMAIN` and `AUTH0_AUDIENCE`:** Your Auth0 domain and API audience.
-   **`S3_*` variables:** The configuration for your production S3 bucket.
-   **`STRIPE_*` variables:** Your Stripe API keys and webhook secret.
-   **`OTEL_EXPORTER_OTLP_ENDPOINT`:** The endpoint of your OpenTelemetry Collector.

For a detailed explanation of all the environment variables, please refer to the [`docs/configuration.md`](configuration.md) document.
