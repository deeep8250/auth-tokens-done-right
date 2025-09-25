# ---------- Builder Stage ----------
FROM golang:1.24-alpine AS builder

# Install tools required for go mod download (git + HTTPS certs)
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum first (better caching for dependencies)
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the Go binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./cmd/server

# ---------- Runtime Stage ----------
FROM alpine:latest

# Install certificates for HTTPS
RUN apk add --no-cache ca-certificates && adduser -D -u 10001 appuser

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/server .

# Run as non-root user
USER appuser

# Expose app port
EXPOSE 8080

# Default command
CMD ["./server"]
