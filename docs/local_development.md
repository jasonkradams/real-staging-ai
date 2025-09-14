# Local Development Setup

This guide provides instructions on how to set up the Virtual Staging AI project for local development.

## Prerequisites

Before you begin, make sure you have the following tools installed on your system:

-   **Go:** Version 1.22 or higher.
-   **Docker and Docker Compose:** For running the project's services.
-   **Node.js:** Version 20 or higher (for future web development).
-   **Auth0 Account:** A free account from [Auth0](https://auth0.com/) for handling authentication.
-   **Stripe Account:** A free account from [Stripe](https://stripe.com/) for payment processing.
-   **AWS CLI:** The AWS Command Line Interface for interacting with S3-compatible storage.

## Environment Setup

1.  **Clone the repository:**

    ```bash
    git clone https://github.com/virtual-staging-ai/api.git
    cd api
    ```

2.  **Set up environment variables:**

    The project uses environment variables for configuration. You can create a `.env` file in the root of the project and the `docker-compose.yml` file will automatically load it.

    For a detailed explanation of all the environment variables, please refer to the [`docs/configuration.md`](configuration.md) document.

3.  **Run the services:**

    To start all the services (API, worker, database, Redis, MinIO), run the following command:

    ```bash
    make up
    ```

    This will start all the services in the background. To stop the services, run:

    ```bash
    make down
    ```

## Running Tests

-   **Unit Tests:** To run the unit tests for all the services, use the following command:

    ```bash
    make test
    ```

-   **Integration Tests:** To run the integration tests, which require the database and other services to be running, use the following command:

    ```bash
    make test-integration
    ```

## Authentication

To test the authenticated endpoints, you will need to generate a test token from Auth0.

For detailed instructions on how to do this, please refer to the [`AUTH_TOKEN_GENERATION.md`](guides/AUTH_TOKEN_GENERATION.md) document.
