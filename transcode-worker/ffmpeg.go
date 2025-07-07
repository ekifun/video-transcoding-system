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

// HandleTranscodeJob runs DASH-compliant segmented FFmpeg job for one representation
func HandleTranscodeJob(job TranscodeJob) {
	if job.Codec == "" {
		job.Codec = "libx264"
	}

	localInput, err := DownloadInput(job.InputURL, job.JobID)
	if err != nil {
		log.Printf("❌ Download failed: %v", err)
		return
	}
	defer os.Remove(localInput)

	outputPattern := filepath.Join(outputDir, fmt.Sprintf("%s_%s_%%03d.mp4", job.JobID, job.Representation))

	cmd := exec.Command("ffmpeg",
		"-i", localInput,
		"-vf", fmt.Sprintf("scale=%s", job.Resolution),
		"-c:v", job.Codec,
		"-b:v", job.Bitrate,
		"-an",
		"-f", "segment",
		"-segment_time", "4",
		"-reset_timestamps", "1",
		"-g", "48",
		"-sc_threshold", "0",
		"-force_key_frames", "expr:gte(t,n_forced*4)",
		"-flags", "+cgop",
		"-movflags", "+faststart",
		"-y", outputPattern,
	)

	log.Printf("⚙️ Running FFmpeg: %v", cmd.String())

	stderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("❌ FFmpeg failed: %v\n%s", err, string(stderr))
		return
	}

	log.Printf("✅ DASH segments generated at: %s", outputPattern)

	// ✅ Update Redis progress
	redisKey := fmt.Sprintf("job:progress:%s", job.JobID)
	if err := redisClient.HSet(ctx, redisKey, job.Representation, "done").Err(); err != nil {
		log.Printf("❌ Failed to update Redis: %v", err)
		return
	}
	// Optional TTL for cleanup
	redisClient.Expire(ctx, redisKey, 1*time.Hour)

	log.Printf("📦 Updated Redis: %s → %s = done", redisKey, job.Representation)
}
