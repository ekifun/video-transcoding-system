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
    go mod tidy
    for dep in "${dependencies[@]}"; do
      echo "📦 Installing Go dependency: $dep"
      go get "$dep"
    done
    go mod tidy
    popd > /dev/null
  else
    echo "✅ Go module already exists in $service_dir."
  fi
}

# Step 1: Initialize Go modules and install dependencies
init_go_mod "./transcoding-controller" "transcoding-controller"
init_go_mod "./transcode-worker" "transcode-worker"
init_go_mod "./tracker" "tracker" "github.com/mattn/go-sqlite3"
init_go_mod "./mpd-generator" "mpd-generator" "
