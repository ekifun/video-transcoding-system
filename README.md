# video-transcoding-system

[Live Demo](https://ekifun.github.io/video-transcoding-system/)  
[Source Code](https://github.com/ekifun/video-transcoding-system/)


This project provides an end-to-end video transcoding pipeline supporting multiple resolutions and codecs (H.264/AVC, HEVC, VVC, VP9, and AV1). It is built with Kafka, Redis, Go microservices, FFmpeg, and MP4Box for DASH packaging and adaptive streaming.

## How to Use the App with Expo Go
#### Install Expo Go on your mobile phone
iOS: Download from App Store
https://apps.apple.com/app/expo-go/id982107779
Android: Download from Google Play
https://play.google.com/store/apps/details?id=host.exp.exponent
#### Scan the QR code
Open the Expo Go app on your device

Scan this QR code to open the app in Expo Go

<img src="./qrcode.png" width="200" height="200" alt="QR Code">

Point your camera at the QR code shown in the terminal or browser

The app will load and run on your phone automatically

## Current Scalability Improvements in Progress
#### Per-Worker Concurrency Control (feature/concurrency-limit-worker)
Implementing a semaphore-based worker pool to limit concurrent FFmpeg jobs within each container, ensuring safe and controlled processing.
#### Redis-Based Centralized Job State Tracking (feature/centralized-redis-job-tracking)
Using Redis to centrally track job progress and state across the entire system for better visibility and management.
#### Optional (Recommended): Prometheus Metrics + Grafana Dashboards
Adding Prometheus and Grafana to monitor real-time system load, active jobs, queue depth, and processing performance through interactive dashboards.

## 1. System Architecture Overview
     +--------------+        +---------------------+     +------------------+
     |              |        |                     |     |                  |
     |  Mobile App  +------->+ Transcode Server    +---->+ Kafka (Job Topic)|
     | (ReactNative)|        | (Go HTTP API)       |     |                  |
     +--------------+        +---------------------+     +------------------+
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

## 2. Components

1. mobile-app/  
React Native app for users to submit a video URL, select target resolutions (144p, 360p, 720p), and choose codec (h264, hevc, or vvc).

2. transcode-server/  
Go-based HTTP server that:
- Accepts POST /transcode requests  
- Validates inputs  
- Creates a Redis job entry: job:<jobID> hash with codec, resolutions, status, etc.  
- Publishes a Kafka message per resolution to the `transcode-jobs` topic  

3. transcode-worker/  
Stateless Go service that:
- Subscribes to `transcode-jobs` Kafka topic  
- Downloads input video  
- Invokes FFmpeg to transcode into target resolution using selected codec  
- Stores MP4 segment in `/segments/`  
- Updates Redis job status  

4. tracker/  
Go service that:
- Periodically scans Redis job hashes  
- Publishes to the `mpd-generation` Kafka topic once all resolutions are complete  

5. mpd-generator/  
Go service that:
- Subscribes to `mpd-generation` Kafka topic  
- Looks up codec from Redis  
- Uses MP4Box to generate `manifest.mpd` for all available outputs  
- Supports AVC, HEVC, and VVC DASH profile output  

## 3. Deployment

### Prerequisites
- Docker + Docker Compose  
- (Optional) Kubernetes for production  
- Node.js + Expo CLI (for mobile app testing)  

### Step 1: Clone the Repository
```bash
git clone https://github.com/ekifun/video-transcoding-system.git
cd video-transcoding-system
```
### Step 2: Deploy Backend
```bash
sh ./deploy.sh
```
Services started:
Redis, Kafka, Zookeeper, and Nginx
transcode-server, transcode-worker, tracker, mpd-generator
### Step 3: Deploy the Mobile App
```bash
cd transcode-mobile
./deploy-transcode-mobile-app.sh
```
Use the Expo app to scan the QR code and interact with the system.
## 4. Testing the System
Submit a Transcode Job
```bash
curl -X POST http://localhost:8080/transcode \
  -H "Content-Type: application/json" \
  -d '{
    "input_url": "https://example.com/video.mp4",
    "resolutions": ["144p", "360p", "720p"],
    "codec": "hevc"
  }'
```
Monitor Logs
```bash
docker compose logs -f transcode-worker
docker compose logs -f tracker
docker compose logs -f mpd-generator
```
Inspect Redis
```bash
docker exec -it redis redis-cli
> keys job:*
> hgetall job:<jobID>
```
### Verify Output

#### [DASH.js Reference Player](https://reference.dashif.org/dash.js/latest/samples/dash-if-reference-player/index.html)

You can use the official [DASH.js Reference Player](https://reference.dashif.org/dash.js/latest/samples/dash-if-reference-player/index.html) to test DASH playback of your transcoded streams. Simply paste the URL to your `manifest.mpd` (e.g., `http://<your-ec2-ip>:8081/segments/<jobID>/manifest.mpd`) into the player’s input field and click **Load**.

Check Codec of Segment
```bash
ffprobe -v error -select_streams v:0 -show_entries stream=codec_name -of default=noprint_wrappers=1:nokey=1 file.mp4
```
## 5. Versioning

v1.0.0: Initial H.264/AVC support and DASH output

v2.0.0: Added HEVC support, unified Redis structure, and DASH output

v3.0.0: Added support for VVC transcoding and DASH output

v4.0.0: Added support for VP9 transcoding and DASH output

v5.0.0: Added support for AV1 transcoding and DASH output

## Authors

Chenghao Liu — Architect & Developer

Contributors welcome

## License

MIT License. See LICENSE for details.