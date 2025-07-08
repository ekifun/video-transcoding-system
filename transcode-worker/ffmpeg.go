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
	log.Printf("📥 [Job %s] Received Job | Codec=%s | Resolution=%s | Bitrate=%s | InputURL=%s | Representation=%s",
		job.JobID, job.Codec, job.Resolution, job.Bitrate, job.InputURL, job.Representation)

	// 🧩 Enhancement 2: Validate input fields
	if job.Resolution == "" || job.Bitrate == "" {
		log.Printf("⚠️ [Job %s] Missing resolution or bitrate. Skipping job.", job.JobID)
		return
	}

	// ✅ Write codec to Redis early so it's available before FFmpeg starts
	redisKey := fmt.Sprintf("job:%s", job.JobID)
	if err := redisClient.HSet(ctx, redisKey, "codec", job.Codec).Err(); err != nil {
		log.Printf("❌ [Job %s] Failed to write codec to Redis: %v", job.JobID, err)
	} else {
		log.Printf("✅ [Job %s] Wrote codec to Redis: %s", job.JobID, job.Codec)
	}

	// Map to FFmpeg encoder
	ffmpegCodec := MapCodecToFFmpeg(job.Codec)

	localInput, err := DownloadInput(job.InputURL, job.JobID)
	if err != nil {
		log.Printf("❌ [Job %s] Download failed: %v", job.JobID, err)
		return
	}
	defer os.Remove(localInput)

	log.Printf("📥 [Job %s] Input downloaded to: %s", job.JobID, localInput)

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

	log.Printf("⚙️ [Job %s] Running FFmpeg: %v", job.JobID, cmd.String())

	stderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("❌ [Job %s] FFmpeg failed: %v\n%s", job.JobID, err, string(stderr))
		return
	}

	log.Printf("✅ [Job %s] DASH segment generated: %s", job.JobID, outputPath)

	// ✅ Update Redis progress (status + output path)
	if err := redisClient.HSet(ctx, redisKey,
		job.Representation, "done",
		fmt.Sprintf("%s_output", job.Representation), outputPath,
	).Err(); err != nil {
		log.Printf("❌ [Job %s] Failed to update Redis progress: %v", job.JobID, err)
		return
	}

	if err := redisClient.Expire(ctx, redisKey, 1*time.Hour).Err(); err != nil {
		log.Printf("⚠️ [Job %s] Failed to set TTL for Redis key: %v", job.JobID, err)
	} else {
		log.Printf("⏳ [Job %s] Set Redis TTL: 1 hour", job.JobID)
	}

	log.Printf("📦 [Job %s] Updated Redis → %s = done, %s_output = %s",
		job.JobID, job.Representation, job.Representation, outputPath)
}


