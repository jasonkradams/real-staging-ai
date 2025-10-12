# Real Staging AI Documentation

This directory contains the source for the Real Staging AI documentation site, built with [Material for MkDocs](https://squidfunk.github.io/mkdocs-material/).

## Local Development

### Prerequisites
- Python 3.11+
- pip

### Quick Start

1. **Install dependencies:**
   ```bash
   make docs-install
   ```

2. **Serve locally:**
   ```bash
   make docs-serve
   ```
   
   Documentation will be available at http://localhost:8000

3. **Or use Docker:**
   ```bash
   make docs-up
   ```

## Building

Build the static site:
```bash
make docs-build
```

Output will be in `site/` directory.

## Structure

```
docs/                        # Source markdown files
├── index.md                # Landing page
├── getting-started/        # Quick start guides
├── architecture/           # System design & components
├── guides/                 # How-to guides
├── operations/             # Deployment & operations
├── api-reference/          # API documentation
├── development/            # Contributing & development
├── implementation-notes/   # Technical implementation details
├── security/               # Security documentation
├── project-history/        # Roadmap & historical docs
└── legal/                  # License & legal
overrides/                  # Custom theme overrides
└── stylesheets/
    └── extra.css          # Custom branding CSS
```

## Deployment

The documentation is designed to be deployed at `docs.real-staging.ai` via Docker container.

Production build:
```bash
docker build -t real-staging-ai/docs:latest .
docker run -p 8000:8000 real-staging-ai/docs:latest
```

## Contributing

When adding or updating documentation:

1. Follow the existing structure
2. Use descriptive filenames
3. Add admonitions for important notes/warnings
4. Test locally before committing
5. Ensure all links work

## Branding

The documentation uses custom branding matching the main Real Staging AI web app:
- Blue (#4F46E5) to Indigo (#6366F1) gradients
- Modern, clean design
- Dark mode support
- Consistent typography
