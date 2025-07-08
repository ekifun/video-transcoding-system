#!/bin/bash

set -e

APP_NAME="transcode-mobile"

echo "🚀 Deploying $APP_NAME with Expo tunnel..."

# Step 1: Install dependencies
echo "📦 Installing dependencies..."
npm install

# Step 2: Ensure expo-checkbox is installed
if ! npm list expo-checkbox >/dev/null 2>&1; then
  echo "📦 Installing expo-checkbox..."
  npx expo install expo-checkbox
else
  echo "✅ expo-checkbox already installed"
fi

# Step 3: Run Expo health check
echo "🩺 Running expo doctor..."
npx expo doctor || true

# Step 4: Start Expo in tunnel mode
echo "🌐 Starting Expo in tunnel mode..."
npx expo start --tunnel
