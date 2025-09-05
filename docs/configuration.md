# Configuration

This document provides a detailed explanation of all the environment variables used in the Virtual Staging AI project.

## API Service (`api`)

| Variable | Description | Default Value |
| --- | --- | --- |
| `APP_ENV` | The application environment. | `dev` |
| `AUTH0_DOMAIN` | Your Auth0 domain. | |
| `AUTH0_AUDIENCE` | The audience for your Auth0 API. | `https://api.virtualstaging.local` |
| `PGHOST` | The hostname of the PostgreSQL database. | `postgres` |
| `PGPORT` | The port of the PostgreSQL database. | `5432` |
| `PGUSER` | The username for the PostgreSQL database. | `postgres` |
| `PGPASSWORD` | The password for the PostgreSQL database. | `postgres` |
| `PGDATABASE` | The name of the PostgreSQL database. | `virtualstaging` |
| `REDIS_ADDR` | The address of the Redis server. | `redis:6379` |
| `S3_ENDPOINT` | The endpoint of the S3-compatible storage. | `http://minio:9000` |
| `S3_REGION` | The region of the S3 bucket. | `us-west-1` |
| `S3_BUCKET` | The name of the S3 bucket. | `virtual-staging` |
| `S3_ACCESS_KEY` | The access key for the S3 bucket. | `minioadmin` |
| `S3_SECRET_KEY` | The secret key for the S3 bucket. | `minioadmin` |
| `S3_USE_PATH_STYLE` | Whether to use path-style addressing for S3. | `true` |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | The endpoint of the OpenTelemetry Collector. | `http://otel:4318` |

## Worker Service (`worker`)

| Variable | Description | Default Value |
| --- | --- | --- |
| `APP_ENV` | The application environment. | `dev` |
| `PGHOST` | The hostname of the PostgreSQL database. | `postgres` |
| `PGPORT` | The port of the PostgreSQL database. | `5432` |
| `PGUSER` | The username for the PostgreSQL database. | `postgres` |
| `PGPASSWORD` | The password for the PostgreSQL database. | `postgres` |
| `PGDATABASE` | The name of the PostgreSQL database. | `virtualstaging` |
| `REDIS_ADDR` | The address of the Redis server. | `redis:6379` |
| `S3_ENDPOINT` | The endpoint of the S3-compatible storage. | `http://minio:9000` |
| `S3_REGION` | The region of the S3 bucket. | `us-west-1` |
| `S3_BUCKET` | The name of the S3 bucket. | `virtual-staging` |
| `S3_ACCESS_KEY` | The access key for the S3 bucket. | `minioadmin` |
| `S3_SECRET_KEY` | The secret key for the S3 bucket. | `minioadmin` |
| `S3_USE_PATH_STYLE` | Whether to use path-style addressing for S3. | `true` |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | The endpoint of the OpenTelemetry Collector. | `http://otel:4318` |
