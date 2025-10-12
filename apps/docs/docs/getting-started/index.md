# Getting Started

Welcome to Real Staging AI! This guide will help you get up and running with the platform quickly.

## What You'll Learn

In this section, you'll learn how to:

- Install and configure your development environment
- Start the local development stack
- Create your first virtual staging project
- Understand the basic workflow

## Prerequisites

Before you begin, ensure you have the following installed:

- **Docker & Docker Compose** - For running services
- **Go 1.22+** - Backend development
- **Node.js 18+** - Frontend development
- **Make** - Build automation

## Quick Start

The fastest way to get started:

```bash
# Clone the repository
git clone https://github.com/jasonkradams/real-staging-ai.git
cd real-staging-ai

# Get a Replicate API token
# Sign up at https://replicate.com
# Get your token from https://replicate.com/account/api-tokens

# Set environment variable
export REPLICATE_API_TOKEN=r8_your_token_here

# Start the stack
make up

# Web app available at: http://localhost:3000
# API available at: http://localhost:8080
```

That's it! You now have Real Staging AI running locally.

## What's Next?

1. **[Installation](installation.md)** - Detailed installation and configuration steps
2. **[Your First Project](first-project.md)** - Create your first staging project
3. **[Configuration Guide](../guides/configuration.md)** - Learn about environment variables

## Development Workflow

For active development:

```bash
# Run tests
make test

# Run integration tests
make test-integration

# Lint code
make lint

# Generate mocks and code
make generate
```

## Common Commands

| Command | Description |
|---------|-------------|
| `make up` | Start development stack |
| `make down` | Stop all services |
| `make test` | Run unit tests |
| `make test-integration` | Run integration tests |
| `make lint` | Run linters |
| `make migrate` | Run database migrations |
| `make token` | Generate Auth0 test token |

## Getting Help

- **Documentation**: You're reading it! Check the sidebar for specific topics
- **GitHub Issues**: [Report bugs or request features](https://github.com/jasonkradams/real-staging-ai/issues)
- **Contributing**: See the [Contributing Guide](../development/contributing.md)

---

Continue to [Installation →](installation.md)
