# Stage 1: Build
FROM golang:1.22-bookworm AS builder

RUN apt-get update && apt-get install -y \
    gcc \
    libc6-dev \
    libsqlite3-dev \
    librdkafka-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY . .

# ✅ Enable CGO for sqlite3 support
ENV CGO_ENABLED=1

RUN go build -o transcoding-controller

# Stage 2: Runtime
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y \
    ca-certificates \
    libsqlite3-0 \
    librdkafka1 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /app/transcoding-controller .

EXPOSE 8080
ENTRYPOINT ["./transcoding-controller"]
