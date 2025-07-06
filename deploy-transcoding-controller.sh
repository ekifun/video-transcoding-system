#!/bin/bash

set -e

echo "📁 Navigating to project root..."
cd "$(dirname "$0")"

# Step 1: Initialize Go module if missing
if [ ! -f "./transcoding-controller/go.mod" ]; then
  echo "🧩 go.mod not found. Initializing Go module..."
  cd ./transcoding-controller
  go mod init transcoding-controller
  go mod tidy
  cd ..
else
  echo "✅ Go module already exists."
fi

# Step 2: Build and launch with Docker Compose
echo "🏗️  Building and starting Docker Compose services..."
docker compose up -d --build

echo "✅ Deployment complete. Access at http://<EC2-IP>:8080/transcode"
