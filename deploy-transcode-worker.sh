#!/bin/bash

set -e

echo "🚀 Preparing to build and deploy transcode-worker..."

cd transcode-worker

# Ensure go.mod and go.sum exist
if [ ! -f go.sum ]; then
  echo "📦 go.sum not found. Running 'go mod tidy' to generate it..."
  go mod tidy
fi

cd ..

# Build Docker image
docker build -t transcode-worker:latest -f transcode-worker/Dockerfile .

echo "✅ transcode-worker Docker image built successfully!"

# Run container (remove existing one first if needed)
docker rm -f transcode-worker 2>/dev/null || true

echo "🐳 Running transcode-worker container..."
docker run -d \
  --name transcode-worker \
  --network=host \
  -e REDIS_ADDR=localhost:6379 \
  -e KAFKA_BROKERS=localhost:9092 \
  transcode-worker:latest

echo "✅ transcode-worker is now running."
