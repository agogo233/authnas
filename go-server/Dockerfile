# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/server ./cmd/server

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates sqlite-libs

# Copy binary from builder
COPY --from=builder /app/server /app/server

# Copy static files (frontend SPA)
COPY --from=builder /app/static /app/static

# Copy keys directory (should be mounted as volume in production)
COPY --from=builder /app/keys /app/keys

# Copy migrations directory
COPY --from=builder /app/migrations /app/migrations

# Copy config directory
COPY --from=builder /app/config /app/config

# Expose port
EXPOSE 8080

# Run the server
CMD ["/app/server"]
