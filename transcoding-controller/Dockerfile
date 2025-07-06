# Stage 1: Build the Go binary
FROM golang:1.22.3-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc g++ librdkafka-dev musl-dev

WORKDIR /app
COPY . .

# Build the Go binary
RUN go build -o transcoding-controller

# Stage 2: Minimal runtime image
FROM alpine:latest

# Install required runtime dependencies
RUN apk add --no-cache ca-certificates librdkafka

WORKDIR /app

COPY --from=builder /app/transcoding-controller .

EXPOSE 8080
ENTRYPOINT ["./transcoding-controller"]
