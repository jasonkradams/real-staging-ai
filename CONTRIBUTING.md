# Contributing to Real Staging AI

First off, thank you for considering contributing to this project! We welcome any contributions, from bug reports and feature requests to code contributions.

## How to Contribute

### Reporting Bugs

If you find a bug, please open an issue on our GitHub repository. When you report a bug, please include the following information:

-   A clear and descriptive title.
-   A detailed description of the bug, including the steps to reproduce it.
-   The expected behavior and the actual behavior.
-   Any relevant logs or screenshots.

### Suggesting Features

If you have an idea for a new feature, please open an issue on our GitHub repository. When you suggest a feature, please include the following information:

-   A clear and descriptive title.
-   A detailed description of the feature and why it would be useful.
-   Any potential implementation ideas.

### Pull Requests

We welcome pull requests for bug fixes, new features, and improvements to the documentation.

When you submit a pull request, please make sure that:

-   Your code follows the project's coding style and conventions.
-   You have added or updated the necessary tests.
-   You have updated the documentation if necessary.
-   Your pull request has a clear and descriptive title and description.

## Coding Style and Conventions

-   **Go:** We follow the standard Go coding style. Please run `gofmt` on your code before submitting a pull request. We also use `golangci-lint` for linting. You can run the linter with `make lint`.
-   **Git Commits:** We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification.

## Makefile Targets

The `Makefile` provides a set of useful targets for common development tasks:

-   `make up`: Start the development environment.
-   `make down`: Stop the development environment.
-   `make test`: Run the unit tests.
-   `make test-integration`: Run the integration tests.
-   `make lint`: Run the linter.
-   `make generate`: Generate mocks and other code.

For a full list of available targets, run `make help`.

Thank you for your contributions!
