# syntax=docker/dockerfile:1

# --- Builder stage ---
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the binary
RUN go build -o quickpulse-server .

# --- Runtime stage ---
FROM alpine:latest

WORKDIR /app

# Copy the binary from the builder
COPY --from=builder /app/quickpulse-server .

# Copy entrypoint script
COPY entrypoint.sh .
RUN chmod +x entrypoint.sh

# Expose ports for gRPC (50051), WebSocket (8081), and Prometheus metrics (8080)
EXPOSE 50051 8081 8080

# Set default environment variables (can be overridden at runtime)
ENV WS_MODE=0
ENV RPC_MODE=1
ENV RPC_STREAM_MODE=0

# Run the entrypoint script (prints banner, then runs server)
CMD ["./entrypoint.sh"]