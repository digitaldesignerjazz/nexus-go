# Dockerfile for nexus-go
# Multi-stage build for minimal, secure image
# Part of the Nexus Ecosystem orchestrator (Esslinger & Co.)

FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Copy go mod files first for better caching
COPY go.mod ./
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /nexus-go .

# Final minimal image
FROM alpine:3.19

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /nexus-go /usr/local/bin/nexus-go

# Create non-root user for security
RUN addgroup -g 1000 nexus && \
    adduser -D -u 1000 -G nexus nexus
USER nexus

ENTRYPOINT ["nexus-go"]
CMD ["help"]
