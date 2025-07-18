#!/bin/bash

set -e

echo "📁 Navigating to project root..."
cd "$(dirname "$0")"

# Function to initialize Go module and install dependencies in a given directory
init_go_mod() {
  local service_dir=$1
  local module_name=$2
  shift 2
  local dependencies=("$@")

  echo "🔍 Checking $service_dir for go.mod..."
  if [ ! -f "$service_dir/go.mod" ]; then
    echo "🧩 Initializing Go module in $service_dir..."
    pushd "$service_dir" > /dev/null
    go mod init "$module_name"
  else
    echo "✅ go.mod exists in $service_dir. Tidying dependencies..."
    pushd "$service_dir" > /dev/null
  fi

  go mod tidy

  for dep in "${dependencies[@]}"; do
    echo "📦 Ensuring dependency: $dep"
    go get "$dep"
  done

  go mod tidy
  popd > /dev/null
}

# Step 1: Initialize Go modules and install dependencies
init_go_mod "./transcoding-controller" "transcoding-controller"
init_go_mod "./transcode-worker" "transcode-worker"
init_go_mod "./tracker" "tracker" "github.com/mattn/go-sqlite3"
init_go_mod "./mpd-generator" "mpd-generator" "github.com/mattn/go-sqlite3"

# Step 2: Build and start services with Docker Compose
echo "🏗️  Building and starting Docker Compose services..."
docker compose up -d --build

# Step 3: Wait for Kafka to be ready
echo "⏳ Waiting for Kafka to be ready..."
sleep 10  # Adjust delay as needed for your environment

# Step 4: Create required Kafka topics
echo "🌀 Creating Kafka topic: mpd-generation"
docker exec -i kafka kafka-topics.sh \
  --create \
  --if-not-exists \
  --topic mpd-generation \
  --bootstrap-server localhost:9092 \
  --partitions 1 \
  --replication-factor 1 2>/dev/null

if [ $? -eq 0 ]; then
  echo "✅ Kafka topic 'mpd-generation' created or already exists."
else
  echo "⚠️ Kafka topic 'mpd-generation' may already exist or creation failed."
fi

echo "✅ Deployment complete."
echo "🌐 Access Transcoding Controller: http://13.57.143.121:8080/transcode"
