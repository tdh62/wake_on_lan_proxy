# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o wol-service -ldflags="-s -w" .

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata wget

# Set timezone
ENV TZ=Asia/Shanghai

WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /build/wol-service /app/wol-service

# Make binary executable
RUN chmod +x /app/wol-service

# Expose port
EXPOSE 24000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:24000/ || exit 1

# Run the application
CMD ["/app/wol-service"]
