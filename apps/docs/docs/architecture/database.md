# Database Schema

This document provides an overview of the database schema for the Real Staging AI project.

## Tables

### `users`

Stores information about users.

| Column               | Type        | Description                                                              |
| -------------------- | ----------- | ------------------------------------------------------------------------ |
| `id`                 | UUID        | Primary key for the user.                                                |
| `auth0_sub`          | TEXT        | The user's subject (sub) claim from Auth0. This is unique for each user. |
| `stripe_customer_id` | TEXT        | The user's customer ID from Stripe.                                      |
| `role`               | TEXT        | The user's role (e.g., `user`, `admin`).                                 |
| `created_at`         | TIMESTAMPTZ | The timestamp when the user was created.                                 |

### `projects`

Stores information about user projects.

| Column       | Type        | Description                                 |
| ------------ | ----------- | ------------------------------------------- |
| `id`         | UUID        | Primary key for the project.                |
| `user_id`    | UUID        | Foreign key to the `users` table.           |
| `name`       | TEXT        | The name of the project.                    |
| `created_at` | TIMESTAMPTZ | The timestamp when the project was created. |

### `images`

Stores information about the images in each project.

| Column         | Type         | Description                                                         |
| -------------- | ------------ | ------------------------------------------------------------------- |
| `id`           | UUID         | Primary key for the image.                                          |
| `project_id`   | UUID         | Foreign key to the `projects` table.                                |
| `original_url` | TEXT         | The URL of the original uploaded image.                             |
| `staged_url`   | TEXT         | The URL of the staged (processed) image.                            |
| `room_type`    | TEXT         | The type of the room in the image (e.g., `living_room`, `bedroom`). |
| `style`        | TEXT         | The staging style (e.g., `modern`, `scandinavian`).                 |
| `status`       | image_status | The status of the image (`queued`, `processing`, `ready`, `error`). |
| `error`        | TEXT         | Any error message if the processing failed.                         |
| `created_at`   | TIMESTAMPTZ  | The timestamp when the image was created.                           |
| `updated_at`   | TIMESTAMPTZ  | The timestamp when the image was last updated.                      |

### `jobs`

Stores information about the background jobs for image processing.

| Column         | Type        | Description                                                            |
| -------------- | ----------- | ---------------------------------------------------------------------- |
| `id`           | UUID        | Primary key for the job.                                               |
| `image_id`     | UUID        | Foreign key to the `images` table.                                     |
| `type`         | TEXT        | The type of the job (e.g., `stage:run`).                               |
| `payload_json` | JSONB       | The JSON payload for the job.                                          |
| `status`       | TEXT        | The status of the job (`queued`, `processing`, `completed`, `failed`). |
| `error`        | TEXT        | Any error message if the job failed.                                   |
| `created_at`   | TIMESTAMPTZ | The timestamp when the job was created.                                |
| `started_at`   | TIMESTAMPTZ | The timestamp when the job started processing.                         |
| `finished_at`  | TIMESTAMPTZ | The timestamp when the job finished processing.                        |

### `plans`

Stores information about the subscription plans.

| Column          | Type | Description                                      |
| --------------- | ---- | ------------------------------------------------ |
| `id`            | UUID | Primary key for the plan.                        |
| `code`          | TEXT | The code for the plan (e.g., `free`, `pro`).     |
| `price_id`      | TEXT | The price ID from Stripe.                        |
| `monthly_limit` | INT  | The number of images a user can stage per month. |

### `processed_events`

Records processed Stripe webhook events to enforce idempotency.

| Column            | Type        | Description                                     |
| ----------------- | ----------- | ----------------------------------------------- |
| `id`              | UUID        | Primary key for the processed event.            |
| `stripe_event_id` | TEXT        | Stripe event ID (unique).                       |
| `type`            | TEXT        | Event type (e.g., `invoice.payment_succeeded`). |
| `payload`         | JSONB       | Raw event payload for auditing.                 |
| `received_at`     | TIMESTAMPTZ | When the event was received.                    |

### `subscriptions`

Tracks Stripe subscription state per user.

| Column                   | Type        | Description                                       |
| ------------------------ | ----------- | ------------------------------------------------- |
| `id`                     | UUID        | Primary key for the subscription row.             |
| `user_id`                | UUID        | Foreign key to `users`.                           |
| `stripe_subscription_id` | TEXT        | Stripe subscription ID (unique).                  |
| `status`                 | TEXT        | Subscription status (e.g., `active`, `past_due`). |
| `price_id`               | TEXT        | Stripe price/plan identifier.                     |
| `current_period_start`   | TIMESTAMPTZ | Current billing period start.                     |
| `current_period_end`     | TIMESTAMPTZ | Current billing period end.                       |
| `cancel_at`              | TIMESTAMPTZ | Scheduled cancel time, if any.                    |
| `canceled_at`            | TIMESTAMPTZ | Time when the subscription was canceled, if any.  |
| `cancel_at_period_end`   | BOOLEAN     | Whether to cancel at period end.                  |
| `created_at`             | TIMESTAMPTZ | Row creation time.                                |
| `updated_at`             | TIMESTAMPTZ | Last update time.                                 |

### `invoices`

Stores Stripe invoice records per user for history and analytics.

| Column                   | Type        | Description                               |
| ------------------------ | ----------- | ----------------------------------------- |
| `id`                     | UUID        | Primary key for the invoice row.          |
| `user_id`                | UUID        | Foreign key to `users`.                   |
| `stripe_invoice_id`      | TEXT        | Stripe invoice ID (unique).               |
| `stripe_subscription_id` | TEXT        | Optional Stripe subscription ID.          |
| `status`                 | TEXT        | Invoice status (e.g., `paid`, `open`).    |
| `amount_due`             | INTEGER     | Amount due in cents.                      |
| `amount_paid`            | INTEGER     | Amount paid in cents.                     |
| `currency`               | TEXT        | Currency (e.g., `usd`).                   |
| `invoice_number`         | TEXT        | Human-readable invoice number (optional). |
| `created_at`             | TIMESTAMPTZ | Row creation time.                        |
| `updated_at`             | TIMESTAMPTZ | Last update time.                         |

## Relationships

- A `user` can have multiple `projects`.
- A `project` belongs to one `user`.
- An `image` belongs to one `project`.
- An `image` can have multiple `jobs`.
- A `job` belongs to one `image`.
