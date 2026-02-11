# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /bot \
    cmd/main.go

# Final stage - using alpine instead of scratch
FROM alpine:latest

# Install CA certificates (already present but ensures latest)
RUN apk --no-cache add ca-certificates

# Copy the binary
COPY --from=builder /bot /bot

# Cloud Run sets PORT environment variable
EXPOSE 8080

# Run the binary
ENTRYPOINT ["/bot"]