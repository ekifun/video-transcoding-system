# Stage 1: Build the Go binary
FROM golang:1.21-alpine AS builder

# Install dependencies and prepare build env
RUN apk add --no-cache git

WORKDIR /app

# Copy all Go source files into the container
COPY . .

# Build the Go binary
RUN go build -o transcoding-controller

# Stage 2: Create minimal runtime image
FROM alpine:latest

# Install ca-certificates (needed for HTTPS requests)
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/transcoding-controller .

# Expose the HTTP port
EXPOSE 8080

# Run the binary
ENTRYPOINT ["./transcoding-controller"]
