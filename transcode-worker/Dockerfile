# Stage 1: Builder
FROM golang:1.21 AS builder

WORKDIR /app
COPY . .

RUN go build -o transcode-worker

# Stage 2: Runner (same base OS to match glibc version)
FROM debian:bullseye-slim  # or use ubuntu:20.04 if needed

WORKDIR /app
COPY --from=builder /app/transcode-worker .

ENTRYPOINT ["./transcode-worker"]
