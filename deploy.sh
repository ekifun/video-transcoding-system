#!/bin/bash

set -e

echo "📁 Navigating to project root..."
cd "$(dirname "$0")"

# Function to initialize Go module in a given directory
init_go_mod() {
  local service_dir=$1
  local module_name=$2

  echo "🔍 Checking $service_dir for go.mod..."
  if [ ! -f "$service_dir/go.mod" ]; then
    echo "🧩 Initializing Go module in $service_dir..."
    pushd "$service_dir" > /dev/null
    go mod init "$module_name"
    go mod tidy
    popd > /dev/null
  else
    echo "✅ Go module already exists in $service_dir."
  fi
}

# Step 1: Initialize Go modules if missing
init_go_mod "./transcoding-controller" "transcoding-controller"
init_go_mod "./transcode-worker" "transcode-worker"
init_go_mod "./tracker" "tracker"

# Step 2: Build and start services with Docker Compose
echo "🏗️  Building and starting Docker Compose services..."
docker compose up -d --build

echo "✅ Deployment complete."
echo "🌐 Access Transcoding Controller: http://13.57.143.121:8080/transcode"
