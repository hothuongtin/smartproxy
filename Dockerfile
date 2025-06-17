# Multi-stage build for minimal SmartProxy image

# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN adduser -D -u 10001 appuser

WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY *.go ./

# Build statically linked binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags='-s -w -extldflags "-static"' \
    -a -installsuffix cgo -o smartproxy .

# Compress binary with UPX for even smaller size
RUN apk add --no-cache upx && \
    upx --best --lzma smartproxy || true

# Final stage using distroless
FROM gcr.io/distroless/static-debian12:nonroot

# Copy CA certificates for HTTPS
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the binary
COPY --from=builder /build/smartproxy /smartproxy

# Copy config files
COPY config.yaml /app/config.yaml
COPY ad_domains.yaml /app/ad_domains.yaml

# Set working directory
WORKDIR /app

# Use non-root user (distroless nonroot has UID 65532)
USER nonroot

# Expose proxy port
EXPOSE 8888

# Set config path environment variable
ENV SMARTPROXY_CONFIG=/app/config.yaml

# Run the proxy
ENTRYPOINT ["/smartproxy"]