#!/bin/bash

set -e

echo "📁 Navigating to project root..."
cd "$(dirname "$0")"

# Ensure build dependencies are installed (Amazon Linux 2, RHEL/CentOS)
install_build_tools_if_needed() {
  if ! command -v cmake &> /dev/null; then
    echo "🔧 cmake not found, installing build tools..."

    if [ -f /etc/os-release ] && grep -qi "amzn" /etc/os-release; then
      sudo yum update -y
      sudo yum groupinstall "Development Tools" -y
      sudo yum install cmake3 -y
      sudo alternatives --install /usr/bin/cmake cmake /usr/bin/cmake3 1 --force
    elif [ -f /etc/debian_version ]; then
      sudo apt update
      sudo apt install -y build-essential cmake
    else
      echo "❌ Unsupported Linux distribution. Please install cmake manually."
      exit 1
    fi
  else
    echo "✅ cmake is already installed."
  fi
}

# Step 0: Build FFmpeg with libvvenc for VVC support
echo "🛠️ Building FFmpeg with libvvenc (H.266/VVC support)..."
install_build_tools_if_needed

if [ ! -d "./ffmpeg-vvc-build" ]; then
  echo "📦 Cloning and building vvenc and FFmpeg..."
  mkdir -p ffmpeg-vvc-build && cd ffmpeg-vvc-build

  # Clone vvenc
  git clone https://github.com/fraunhoferhhi/vvenc.git
  cd vvenc
  mkdir build && cd build
  cmake .. -DCMAKE_BUILD_TYPE=Release
  make -j$(nproc)
  sudo make install
  cd ../../

  # Clone FFmpeg
  git clone https://github.com/FFmpeg/FFmpeg.git
  cd FFmpeg
  ./configure --enable-gpl --enable-nonfree --enable-libvvenc
  make -j$(nproc)
  sudo make install
  cd ../..
else
  echo "✅ ffmpeg-vvc-build directory already exists. Skipping rebuild."
fi

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
init_go_mod "./mpd-generator" "mpd-generator"

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
