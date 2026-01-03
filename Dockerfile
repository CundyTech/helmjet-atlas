# Multi-stage build
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install git so `go` can fetch VCS-backed modules during build
RUN apk add --no-cache git

# Remove submodule go.mod files so the builder uses the single top-level go.mod
RUN rm -f api/go.mod integrations/*/go.mod scripts/go.mod go.work || true

# Copy full source code (workspace includes go.work)
COPY . .

# Build the application
# Build the API package (single binary) from the `cmd` folder
# (project uses `cmd/main.go` as the entrypoint)
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o helmjet-atlas ./cmd

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/helmjet-atlas .
# Copy frontend assets so the binary can serve them at runtime
COPY --from=builder /app/visualization ./visualization

# Expose port
EXPOSE 8080

# Run the application
CMD ["./helmjet-atlas"]
