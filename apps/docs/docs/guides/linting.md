# Linting Configuration Guide

This document describes our golangci-lint configuration and standards for maintaining code quality.

## Overview

We use [golangci-lint](https://golangci-lint.run/) with a comprehensive set of linters to ensure:
- **Code Quality**: Consistent, idiomatic Go code
- **Security**: Detection of security vulnerabilities and unsafe patterns
- **Maintainability**: Prevention of code drift and technical debt
- **Error Handling**: Proper error wrapping and handling patterns
- **Performance**: Detection of inefficient code patterns

## Configuration

Location: `.golangci.yml`

### Core Linters (Always Enabled)

These fundamental linters catch critical bugs and are always enabled:

- **errcheck**: Checks for unchecked errors (prevents silent failures)
- **govet**: Official Go static analyzer (detects suspicious constructs)
- **ineffassign**: Detects ineffectual assignments
- **unused**: Finds unused constants, variables, functions, and types

### Quality Linters

#### Code Style & Best Practices
- **gocritic**: Opinionated source code checker with multiple diagnostic tags
- **gocyclo**: Cyclomatic complexity checker (threshold: 15)
- **dupl**: Detects code duplication (threshold: 60 lines)
- **misspell**: Finds spelling errors in comments
- **nakedret**: Prevents naked returns in long functions (max: 30 lines)
- **unconvert**: Removes unnecessary type conversions

#### Error Handling
- **errorlint**: Ensures proper error wrapping (Go 1.13+ patterns)
- **errname**: Enforces sentinel error naming (Err prefix)
- **nilerr**: Catches returning nil when error is not nil
- **nilnil**: Prevents simultaneous return of nil error and nil value

#### Security
- **gosec**: Inspects source for security problems
- **bodyclose**: Ensures HTTP response bodies are closed
- **sqlclosecheck**: Ensures SQL connections are properly closed
- **rowserrcheck**: Checks that database row errors are handled

#### Resource Management
- **bodyclose**: HTTP response body leak detection
- **sqlclosecheck**: SQL connection leak detection
- **rowserrcheck**: Database row error checking

#### Modern Go Patterns
- **copyloopvar**: Loop variable copying detection (Go 1.22+)
- **mirror**: Detects incorrect use of reverse patterns
- **usestdlibvars**: Suggests using stdlib constants

#### Code Organization
- **goconst**: Finds repeated strings that should be constants
- **godot**: Ensures comments end with periods
- **nolintlint**: Validates nolint directives (requires explanations)
- **testpackage**: Encourages separate `_test` packages
- **testifylint**: Validates testify usage patterns

#### API & Interfaces
- **revive**: Fast, configurable linter (golint replacement)
- **noctx**: Ensures context.Context is passed to functions
- **nosprintfhostport**: Detects incorrect URL construction

## Linter Settings

### Duplication Detection
```yaml
dupl:
  threshold: 60  # Lines of code considered duplicate
```

### Error Checking
```yaml
errcheck:
  check-type-assertions: true
  check-blank: true
  exclude-functions:
    - (io.Closer).Close
    - (*database/sql.Rows).Close
```

### Complexity Limits
```yaml
gocyclo:
  min-complexity: 15  # Maximum cyclomatic complexity
```

### Security
```yaml
gosec:
  excludes:
    - G104  # Covered by errcheck
  confidence: medium
```

### Code Review Standards
```yaml
nolintlint:
  allow-unused: false
  require-explanation: true  # Must explain why linting is disabled
  require-specific: true     # Must specify which linter to disable
```

## Exclusions & Special Cases

### Test Files
Test files have relaxed rules for:
- `gocyclo`: Tests can be more complex
- `dupl`: Test setup code may be duplicated
- `gosec`: Security checks less critical in tests
- `goconst`: Test data can have repeated strings

### Generated Code
Generated files (`*_mock.go`, `queries/*.go`) are excluded from:
- Style checks
- Complexity checks
- Line length limits

### Special Patterns
- Lines starting with `//go:generate` are exempt from line length limits
- Mock files are exempt from complexity and duplication checks

## Running Linters

### Local Development
```bash
# Run all linters
make lint

# Auto-fix issues where possible
golangci-lint run --fix
```

### CI/CD
The linting runs automatically:
- On every commit (via git hooks)
- In pull requests
- Before merging to main

Exit code 0 = no issues, CI passes
Exit code 1 = issues found, CI fails

## Disabling Linters

When you must disable a linter, follow these rules:

### Inline Disabling (Preferred)
```go
//nolint:lintername // Reason: explanation of why this is necessary
problematicCode()
```

### File-Level Disabling
```go
//nolint:lintername // Reason: explanation
package mypackage
```

### Configuration Disabling (Last Resort)
Add to `.golangci.yml` under `issues.exclude-rules` with clear documentation.

## Common Issues & Solutions

### Error Handling
**Issue**: `comparing with == will fail on wrapped errors`
```go
// Bad
if err == pgx.ErrNoRows { }

// Good
if errors.Is(err, pgx.ErrNoRows) { }
```

### Type Assertions
**Issue**: `type assertion on error will fail on wrapped errors`
```go
// Bad
httpErr, ok := err.(*echo.HTTPError)

// Good
var httpErr *echo.HTTPError
if errors.As(err, &httpErr) { }
```

### Stdlib Constants
**Issue**: `"POST" can be replaced by http.MethodPost`
```go
// Bad
req, _ := http.NewRequest("POST", url, body)

// Good
req, _ := http.NewRequest(http.MethodPost, url, body)
```

### Imports
Imports should be organized in three groups with blank lines:
1. Standard library
2. External dependencies
3. Internal packages (github.com/real-staging-ai/*)

## Severity Levels

- **ERROR**: Fails CI, must be fixed
- **WARNING**: Informational, consider fixing

Current warnings:
- `dupl`: Code duplication (refactor when possible)
- `lll`: Long lines (improve readability)

## Adding New Linters

When enabling a new linter:

1. Enable in `.golangci.yml`
2. Run `make lint` to see issues
3. Fix issues or add justified exclusions
4. Document the linter in this file
5. Commit with message: `chore(lint): enable <linter-name>`

## Maintenance

### Review Schedule
- **Weekly**: Review new linter violations
- **Monthly**: Evaluate if linter thresholds need adjustment
- **Quarterly**: Review and update linter configuration

### Metrics
Track these over time:
- Total issues found
- Issues by linter type
- Issues fixed vs. suppressed
- Code coverage trends

## Resources

- [golangci-lint Documentation](https://golangci-lint.run/)
- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)

## Version History

- **2025-01**: Initial comprehensive configuration with 48+ linters
- Thresholds set based on current codebase analysis
- Formatters integrated (goimports)
- Security-focused settings enabled
