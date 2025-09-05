# Database Schema

This document provides an overview of the database schema for the Virtual Staging AI project.

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
| `seed`         | BIGINT       | The seed used for the staging process.                              |
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
| `type`         | TEXT        | The type of the job (e.g., `stage_image`).                             |
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

## Relationships

- A `user` can have multiple `projects`.
- A `project` belongs to one `user`.
- A `project` can have multiple `images`.
- An `image` belongs to one `project`.
- An `image` can have multiple `jobs`.
- A `job` belongs to one `image`.
