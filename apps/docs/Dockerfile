FROM python:3.11-slim

LABEL maintainer="Real Staging AI <hello@real-staging.ai>"
LABEL description="Documentation site for Real Staging AI"

WORKDIR /docs

# Install system dependencies (git for git-revision-date plugin)
RUN apt-get update && apt-get install -y --no-install-recommends \
    git \
    && rm -rf /var/lib/apt/lists/*

# Install Python dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy documentation source
COPY . .

# Expose default MkDocs serve port
EXPOSE 8000

# Serve documentation
CMD ["mkdocs", "serve", "-a", "0.0.0.0:8000"]
