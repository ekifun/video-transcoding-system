#!/bin/bash

set -e

echo "🚀 Preparing to build and deploy transcode-worker..."

cd transcode-worker

# Initialize go.mod if not present
if [ ! -f go.mod ]; then
  echo "📦 go.mod not found. Initializing Go module..."
  go mod init transcode-worker
fi

# Ensure Kafka dependency is present in go.mod
if ! grep -q "github.com/confluentinc/confluent-kafka-go" go.mod 2>/dev/null; then
  echo "📦 Adding Kafka dependency..."
  go get github.com/confluentinc/confluent-kafka-go/kafka
fi

# Ensure go.sum exists and is synced
echo "📦 Tidying Go modules..."
go mod tidy

cd ..

echo "🐳 Building transcode-worker Docker image..."
docker build -t transcode-worker:latest -f transcode-worker/Dockerfile ./transcode-worker

echo "✅ Docker image built!"

# Clean up stale container if needed
if docker ps -a --format '{{.Names}}' | grep -q '^transcode-worker$'; then
  echo "🧹 Removing existing transcode-worker container..."
  docker rm -f transcode-worker
fi

echo "🚀 Starting transcode-worker container..."
docker run -d \
  --name transcode-worker \
  --network=host \
  -e REDIS_ADDR=localhost:6379 \
  -e KAFKA_BROKERS=localhost:9092 \
  transcode-worker:latest

echo "✅ transcode-worker is now running."

echo "📄 Logs: run 'docker logs -f transcode-worker' to view output"
