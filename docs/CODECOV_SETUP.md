# Codecov Setup

This document explains how to set up Codecov for coverage reporting.

## Setup Steps

1. **Sign up for Codecov**
   - Go to [codecov.io](https://codecov.io)
   - Sign in with your GitHub account
   - Authorize Codecov to access your repository

2. **Get Your Codecov Token**
   - Navigate to your repository in Codecov
   - Go to Settings → General
   - Copy the "Repository Upload Token"

3. **Add Token to GitHub Secrets**
   - Go to your GitHub repository
   - Navigate to Settings → Secrets and variables → Actions
   - Click "New repository secret"
   - Name: `CODECOV_TOKEN`
   - Value: Paste your Codecov token
   - Click "Add secret"

4. **Update README Badge**
   - Get your badge token from Codecov Settings → Badge
   - Replace `YOUR_TOKEN_HERE` in README.md with the actual token:
     ```markdown
     [![codecov](https://codecov.io/gh/jasonkradams/real-staging-ai/graph/badge.svg?token=YOUR_ACTUAL_TOKEN)](https://codecov.io/gh/jasonkradams/real-staging-ai)
     ```

## Local Testing

Generate coverage reports locally:

```bash
# Generate coverage reports (excludes mocks)
make coverage

# Generate HTML reports and open in browser
make coverage-html

# Show summary
make coverage-summary
```

## Coverage Configuration

Coverage settings are configured in `codecov.yml`:
- Mock files (`*_mock.go`) are excluded
- Test files are excluded
- Command-line tools (`cmd/`) are excluded
- Target: Auto with 1% threshold

## CI Integration

The GitHub Actions workflow (`.github/workflows/coverage.yml`) automatically:
1. Runs tests with coverage on every push to `main` and PRs
2. Excludes `*_mock.go` files from coverage reports
3. Uploads coverage to Codecov
4. Displays coverage summary in PR checks

## Viewing Coverage

- **Online**: Visit https://codecov.io/gh/jasonkradams/real-staging-ai
- **Badge**: Shows overall coverage on README
- **PR Comments**: Codecov automatically comments on PRs with coverage changes
