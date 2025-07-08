package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ctx         = context.Background()
	redisAddr   = os.Getenv("REDIS_ADDR") // e.g. "redis:6379"
	redisClient *redis.Client
	outputDir   = "/segments" // Shared volume path
)

func init() {
	redisClient = redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		log.Fatalf("❌ Failed to connect to Redis: %v", err)
	}
	log.Println("✅ Connected to Redis")
}

// DownloadInput downloads the input file to /segments/{jobID}_input.mp4 and returns its local path.
func DownloadInput(inputURL string, jobID string) (string, error) {
	log.Printf("🌐 Downloading input from: %s", inputURL)
	localPath := filepath.Join(outputDir, fmt.Sprintf("%s_input.mp4", jobID))

	outFile, err := os.Create(localPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer outFile.Close()

	resp, err := http.Get(inputURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch input: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	log.Printf("📥 Downloaded to: %s", localPath)
	return localPath, nil
}

// MapCodecToFFmpeg maps "h264" and "hevc" to FFmpeg codec names
func MapCodecToFFmpeg(codec string) string {
	switch codec {
	case "hevc", "h265":
		return "libx265"
	case "h264":
		return "libx264"
	default:
		log.Printf("⚠️ Unsupported codec '%s'. Defaulting to h264 (libx264)", codec)
		return "libx264"
	}
}

func HandleTranscodeJob(job TranscodeJob) {
	if job.Codec == "" {
		job.Codec = "h264" // Default to H.264
	}

	// 🧩 Enhancement 1: Log input job parameters
	log.Printf("📥 Received Job: ID=%s | Codec=%s | Resolution=%s | Bitrate=%s | InputURL=%s | Representation=%s",
		job.JobID, job.Codec, job.Resolution, job.Bitrate, job.InputURL, job.Representation)

	// 🧩 Enhancement 2: Validate input fields
	if job.Resolution == "" || job.Bitrate == "" {
		log.Printf("⚠️ Missing resolution or bitrate. Skipping job: %+v", job)
		return
	}

	// ✅ Write codec to Redis early so it's available before FFmpeg starts
	redisKey := fmt.Sprintf("job:progress:%s", job.JobID)
	if err := redisClient.HSet(ctx, redisKey, "codec", job.Codec).Err(); err != nil {
		log.Printf("⚠️ Failed to write codec to Redis: %v", err)
	}

	// Map to FFmpeg encoder
	ffmpegCodec := MapCodecToFFmpeg(job.Codec)

	localInput, err := DownloadInput(job.InputURL, job.JobID)
	if err != nil {
		log.Printf("❌ Download failed: %v", err)
		return
	}
	defer os.Remove(localInput)

	outputPath := filepath.Join(outputDir, fmt.Sprintf("%s_%s.mp4", job.JobID, job.Representation))

	cmd := exec.Command("ffmpeg",
		"-i", localInput,
		"-vf", fmt.Sprintf("scale=%s", job.Resolution),
		"-c:v", ffmpegCodec,
		"-b:v", job.Bitrate,
		"-g", "48",
		"-keyint_min", "48",
		"-sc_threshold", "0",
		"-an",
		"-movflags", "+faststart+frag_keyframe+empty_moov+default_base_moof",
		"-y", outputPath,
	)

	log.Printf("⚙️ Running FFmpeg: %v", cmd.String())

	stderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("❌ FFmpeg failed: %v\n%s", err, string(stderr))
		return
	}

	log.Printf("✅ DASH segments generated at: %s", outputPath)

	// ✅ Update Redis progress (status + output path)
	if err := redisClient.HSet(ctx, redisKey,
		job.Representation, "done",
		fmt.Sprintf("%s_output", job.Representation), outputPath,
	).Err(); err != nil {
		log.Printf("❌ Failed to update Redis: %v", err)
		return
	}

	redisClient.Expire(ctx, redisKey, 1*time.Hour)

	log.Printf("📦 Updated Redis: %s → %s = done, %s_output = %s",
		redisKey, job.Representation, job.Representation, outputPath)
}


