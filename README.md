# video-transcoding-system

Here’s a full README.md file for your video transcoding system that includes:
Overview of the architecture and components
Functional responsibilities
Deployment instructions
Testing guidelines
# 🎥 Video Transcoding System with H.264 and HEVC Support

This project provides an end-to-end video transcoding pipeline supporting multiple resolutions and codecs (H.264 and HEVC). It is built with **Kafka**, **Redis**, **Go microservices**, and **MP4Box** for DASH packaging.

---

## 🧱 System Architecture Overview

```plaintext
     +--------------+        +----------------+         +------------------+
     |              |        |                |         |                  |
     |  Mobile App  +------->+ Transcode Ctrl +-------->+ Kafka (Job Topic)|
     | (ReactNative)|        | (Go HTTP API)  |         |                  |
     +--------------+        +----------------+         +------------------+
                                    |
                                    v
                       +------------------------+
                       |   Transcode Workers    |
                       |   (Go + FFmpeg)        |
                       +------------------------+
                                    |
                                    v
                         +--------------------+
                         | Redis (Job Hashes) |
                         +--------------------+
                                    |
                                    v
                         +--------------------+
                         | Output Tracker     |
                         | (Go)               |
                         +--------------------+
                                    |
                                    v
                      +---------------------------+
                      | Kafka (mpd-generation topic)
                      +---------------------------+
                                    |
                                    v
                            +------------------+
                            | MPD Generator    |
                            | (Go + MP4Box)    |
                            +------------------+

📦 Components
1. mobile-app/
React Native app for users to submit a video URL, select target resolutions (144p, 360p, 720p), and choose codec (h264 or hevc).
2. controller/
A Go-based HTTP server that:
Accepts POST /transcode requests
Validates inputs
Creates a Redis job entry: job:<jobID> hash with codec, resolutions, status, etc.
Publishes a Kafka message per resolution to the transcode-jobs topic
3. transcode-worker/
Stateless Go service that:
Subscribes to transcode-jobs Kafka topic
Downloads input video
Invokes ffmpeg to transcode into target resolution
Stores MP4 segment in /segments/
Updates Redis job status (e.g., 144p: done)
4. tracker/
Go service that:
Periodically scans Redis job:* hashes
Calls publishReadyForMPD(jobID) if all resolutions are complete
Publishes message to mpd-generation Kafka topic
5. mpd-generator/
Go service that:
Subscribes to mpd-generation Kafka topic
Looks up codec from Redis
Uses MP4Box to generate DASH manifest.mpd for all ready resolutions
Skips -profile flag for HEVC jobs (to avoid MP4Box profile mismatch)

🚀 Deployment
Pre-requisites:
Docker + Docker Compose
(Optional) Kubernetes for production
Node.js + Expo CLI (for mobile app testing)
🔁 Step 1: Clone the Repository
git clone https://github.com/your-org/video-transcoding-system.git
cd video-transcoding-system
🖥️ Step 2: Deploy Backend (Redis, Kafka, Services)
1. Start Core Infrastructure
docker compose -f infra/docker-compose.infra.yml up -d
This brings up:
redis for job coordination
kafka + zookeeper for message queue
nginx for static segment delivery on port 8081
2. Deploy Backend Services
./deploy.sh
This script builds and starts the following services:
controller: Receives transcode requests
transcode-worker: Performs transcoding via FFmpeg
tracker: Monitors Redis to check job completion
mpd-generator: Creates DASH manifest.mpd using MP4Box
All output segments are stored under /segments and served via nginx.
📱 Step 3: Deploy the Mobile App (Transcode Client)
cd transcode-mobile
./deploy-transcode-mobile-app.sh
This script starts the Expo development server. You can open the app in Expo Go (iOS/Android) by scanning the QR code shown in terminal or browser.
The mobile app allows users to:
Input a video URL
Select resolutions (144p, 360p, 720p)
Choose codec (h264 or hevc)
Submit the transcode job to your backend API

✅ Testing the System
1. Submit a Transcode Job (Mobile or CURL)
Use the mobile app or run:
curl -X POST http://localhost:8080/transcode \
  -H "Content-Type: application/json" \
  -d '{
    "input_url": "https://example.com/video.mp4",
    "resolutions": ["144p", "360p", "720p"],
    "codec": "hevc"
  }'
2. Monitor Logs
docker compose logs -f transcode-worker
docker compose logs -f tracker
docker compose logs -f mpd-generator
3. Inspect Redis
docker exec -it redis redis-cli
> keys job:*
> hgetall job:<jobID>
4. Verify Segments and Manifest
http://<your-ec2-ip>:8081/segments/<jobID>/manifest.mpd
Use tools like Shaka Player to test playback.
🔍 How to Identify Codec of a Segment
ffprobe -v error -select_streams v:0 -show_entries stream=codec_name -of default=noprint_wrappers=1:nokey=1 file.mp4

🏷️ Versioning
v1.0.0: Initial H.264 support
v2.0.0: Added HEVC support and unified Redis job structure

👥 Authors
Chenghao Liu — Architect & Developer
Contributors welcome!

📄 License
MIT License. See LICENSE for more details.