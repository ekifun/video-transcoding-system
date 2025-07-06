# Stage 1: Build
FROM golang:1.22-bookworm AS builder

RUN apt-get update && apt-get install -y \
    librdkafka-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY . .

RUN go build -o transcoding-controller

# Stage 2: Runtime
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y \
    ca-certificates librdkafka1 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /app/transcoding-controller .

EXPOSE 8080
ENTRYPOINT ["./transcoding-controller"]
