#!/bin/bash

set -e

echo "🚀 Preparing to build and deploy transcode-worker..."

cd transcode-worker

# ✅ Step 1: Ensure go.sum exists
if [ ! -f "go.sum" ]; then
  echo "📦 go.sum not found. Running 'go mod tidy' to generate it..."
  go mod tidy
else
  echo "✅ go.sum already exists."
fi

# ✅ Step 2: Build Docker image
echo "🐳 Building Docker image: transcode-worker:latest"
docker build -t transcode-worker:latest .

# ✅ Step 3: Start container using docker-compose
echo "🧩 Restarting transcode-worker service via docker-compose..."
cd ..
docker compose up -d --build transcode-worker

echo "🎉 transcode-worker deployed successfully!"
