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

# Ensure go.sum exists
if [ ! -f go.sum ]; then
  echo "📦 go.sum not found. Running 'go mod tidy' to generate it..."
  go mod tidy
fi

cd ..  # go back to root so docker build context is correct

echo "🐳 Building transcode-worker Docker image..."
docker build -t transcode-worker:latest -f transcode-worker/Dockerfile ./transcode-worker

echo "✅ Docker image built!"

# Remove existing container if exists
docker rm -f transcode-worker 2>/dev/null || true

echo "🚀 Starting transcode-worker container..."
docker run -d \
  --name transcode-worker \
  --network=host \
  -e REDIS_ADDR=localhost:6379 \
  -e KAFKA_BROKERS=localhost:9092 \
  transcode-worker:latest

echo "✅ transcode-worker is now running."
