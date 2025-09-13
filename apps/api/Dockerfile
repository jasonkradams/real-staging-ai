# ---- Builder ----
FROM golang:1.25.1-alpine AS builder

WORKDIR /app

# Copy module files and download deps
COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download

# Copy API source code
COPY apps/api/ ./

# Build the application
# CGO_ENABLED=0: build a statically linked binary
# -o /api-server: specify the output file name
# ./cmd/api: specify the main package to build
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /api-server ./cmd/api

# ---- Runner ----
FROM alpine:latest

# Copy the compiled binary from the builder stage
COPY --from=builder /api-server /api-server

# Expose the port the application runs on
EXPOSE 8080

# Set the entrypoint for the container
ENTRYPOINT ["/api-server"]
