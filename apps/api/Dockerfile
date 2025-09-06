# ---- Builder ----
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum to download dependencies first
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

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
