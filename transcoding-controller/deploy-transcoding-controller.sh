#!/bin/bash

set -e

echo "📁 Ensuring we're in the project directory..."
cd "$(dirname "$0")"

# Step 1: Initialize Go module if missing
if [ ! -f "go.mod" ]; then
  echo "🧩 go.mod not found. Initializing Go module..."
  go mod init transcoding-controller
  go mod tidy
else
  echo "✅ Go module detected. Skipping go mod init."
fi

# Step 2: Build the Docker image locally
echo "🏗️  Building Docker image: transcoding-controller:latest"
docker build -t transcoding-controller:latest .

# Step 3: Run Docker Compose
echo "🚀 Starting Docker Compose stack..."
docker compose up -d --no-deps --build transcoding-controller

echo "✅ Deployment complete. Access service at http://<EC2-IP>:8080/transcode"
