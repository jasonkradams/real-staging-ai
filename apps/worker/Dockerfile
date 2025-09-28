# ---- Builder ----
FROM golang:1.25.1-alpine AS builder

WORKDIR /app

# Copy module files and download deps
COPY apps/worker/go.mod apps/worker/go.sum ./
COPY apps/api/ ../api
RUN go mod download

# Copy worker source code
COPY apps/worker/ ./

# Build the application
# CGO_ENABLED=0: build a statically linked binary
# -o /worker-server: specify the output file name
# ./cmd/worker: specify the main package to build
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /worker-server ./main.go

# ---- Runner ----
FROM alpine:latest

# Copy the compiled binary from the builder stage
COPY --from=builder /worker-server /worker-server

# Expose the port the application runs on
EXPOSE 8080

# Set the entrypoint for the container
ENTRYPOINT ["/worker-server"]
