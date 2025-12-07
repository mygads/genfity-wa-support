# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install git for private dependencies if needed
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# ===================================================================================
# Final stage - minimal image
# ===================================================================================
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates wget curl tzdata

WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Create non-root user for security
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup && \
    chown -R appuser:appgroup /app

USER appuser

# Expose port
EXPOSE 8082

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8082/health || exit 1

# Command to run
CMD ["./main"]
