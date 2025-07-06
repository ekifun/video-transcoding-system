#!/bin/bash

set -e

echo "🚀 Installing Docker..."

# Install Docker
sudo yum update -y
sudo amazon-linux-extras enable docker
sudo yum install -y docker
sudo systemctl enable docker
sudo systemctl start docker
sudo usermod -aG docker $USER

# Apply Docker group without logout
newgrp docker << END

echo "✅ Docker installed"

# Install docker-compose v2
echo "🚀 Installing Docker Compose v2..."
mkdir -p ~/.docker/cli-plugins
curl -SL https://github.com/docker/compose/releases/download/v2.24.5/docker-compose-linux-x86_64 -o ~/.docker/cli-plugins/docker-compose
chmod +x ~/.docker/cli-plugins/docker-compose
docker compose version

# Create deployment folder
mkdir -p ~/video-transcoding-deploy
cd ~/video-transcoding-deploy

echo "📦 Creating docker-compose.yml..."

cat <<EOF > docker-compose.yml
version: "3"
services:
  redis:
    image: redis:7-alpine
    container_name: redis
    ports:
      - "6379:6379"

  kafka:
    image: bitnami/kafka:latest
    container_name: kafka
    ports:
      - "9092:9092"
    environment:
      KAFKA_CFG_ZOOKEEPER_CONNECT: zookeeper:2181
      KAFKA_CFG_ADVERTISED_LISTENERS: PLAINTEXT://localhost:9092
      KAFKA_CFG_LISTENERS: PLAINTEXT://0.0.0.0:9092
    depends_on:
      - zookeeper

  zookeeper:
    image: bitnami/zookeeper:latest
    container_name: zookeeper
    ports:
      - "2181:2181"

  transcoding-controller:
    image: your-dockerhub/transcoding-controller:latest
    container_name: transcoding-controller
    ports:
      - "8080:8080"
    environment:
      REDIS_ADDR: redis:6379
      KAFKA_BROKERS: kafka:9092
    depends_on:
      - kafka
      - redis
EOF

echo "🚀 Starting services..."
docker compose up -d

echo "✅ All services are up and running!"
echo "🔗 Access your controller at: http://<your-ec2-ip>:8080/transcode"

END
