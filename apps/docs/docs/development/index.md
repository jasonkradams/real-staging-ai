# Development

Resources for contributing to and developing Real Staging AI.

## Getting Started

- **[Contributing Guide](contributing.md)** - How to contribute to the project
- **[Repository Guidelines](repository-guidelines.md)** - Code standards and conventions

## Advanced Topics

- **[Cost Tracking](cost-tracking.md)** - Monitor and optimize AI processing costs
- **[Model Registry](model-registry.md)** - AI model architecture and management

## Quick Reference

### Essential Commands

```bash
# Development
make up              # Start dev environment
make down            # Stop services
make test            # Run unit tests
make test-integration # Run integration tests
make lint            # Run linters
make generate        # Generate mocks and code

# Database
make migrate         # Run migrations
make migrate-down-dev # Rollback migrations

# Utilities
make token           # Generate Auth0 token
make clean           # Clean build artifacts
```

### Project Structure

```
apps/
├── api/            # Go HTTP API
│   ├── cmd/        # Entry points
│   ├── internal/   # Domain packages
│   └── tests/      # Integration tests
├── worker/         # Background job processor
│   ├── internal/   # Job handlers
│   └── tests/      # Integration tests
├── web/            # Next.js frontend
└── docs/           # This documentation site

infra/
└── migrations/     # SQL migrations

config/             # Environment configs
└── *.yml           # Dev, test, prod configs
```

### Coding Standards

**Go:**
- Format with `gofmt`
- Lint with `golangci-lint`
- Follow [Effective Go](https://go.dev/doc/effective_go)
- Write tests for all code
- Use table-driven tests

**Git:**
- Follow [Conventional Commits](https://www.conventionalcommits.org/)
- Examples: `feat(api): add presign endpoint`, `fix(worker): handle timeout`
- Create focused, single-purpose commits

**Documentation:**
- Update docs with code changes
- Use clear, concise language
- Include code examples
- Keep diagrams up to date

## Testing

See the comprehensive [Testing Guide](../guides/testing.md) for details.

**Quick test commands:**
```bash
make test               # Unit tests
make test-cover         # With coverage report
make test-integration   # Integration tests
```

## Development Workflow

1. **Create feature branch**
   ```bash
   git checkout -b feat/my-feature
   ```

2. **Write tests first (TDD)**
   ```bash
   # Write failing test
   go test -v ./internal/mypackage
   ```

3. **Implement feature**
   ```bash
   # Write minimal code to pass
   ```

4. **Run linters**
   ```bash
   make lint
   ```

5. **Commit with conventional format**
   ```bash
   git commit -m "feat(api): add new endpoint"
   ```

6. **Push and create PR**
   ```bash
   git push origin feat/my-feature
   ```

## Code Review Checklist

Before submitting a PR:

- [ ] Tests pass (`make test`)
- [ ] Integration tests pass (`make test-integration`)
- [ ] Linters pass (`make lint`)
- [ ] Code coverage maintained/improved
- [ ] Documentation updated
- [ ] Conventional commit messages used
- [ ] No secrets or credentials committed
- [ ] Error handling implemented
- [ ] Logging added for important operations
- [ ] OpenAPI spec updated (if API changes)

## Resources

- [Contributing Guide](contributing.md) - Full contribution process
- [Repository Guidelines](repository-guidelines.md) - Detailed coding standards
- [Testing Guide](../guides/testing.md) - Comprehensive testing docs
- [Architecture](../architecture/) - System design and components

## Getting Help

- **Documentation Issues**: Open an issue with `docs` label
- **Bug Reports**: Use the bug report template
- **Feature Requests**: Use the feature request template
- **Questions**: Start a discussion in GitHub Discussions

---

**Ready to contribute?** Start with the [Contributing Guide →](contributing.md)
