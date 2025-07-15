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
	redisAddr   = os.Getenv("REDIS_ADDR")
	redisClient *redis.Client
	outputDir   = "/segments"
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

func MapCodecToFFmpeg(codec string) string {
	switch codec {
	case "hevc", "h265":
		return "libx265"
	case "h264":
		return "libx264"
	case "vvc", "h266":
		return "libvvenc"
	default:
		log.Printf("⚠️ Unsupported codec '%s'. Defaulting to h264 (libx264)", codec)
		return "libx264"
	}
}

func HandleTranscodeJob(job TranscodeJob) {
	if job.Codec == "" {
		job.Codec = "h264"
	}

	log.Printf("📥 [Job %s] Received Job | Codec=%s | Resolution=%s | Bitrate=%s | GOP=%d | KeyintMin=%d",
		job.JobID, job.Codec, job.Resolution, job.Bitrate, job.GopSize, job.KeyintMin)

	if job.Resolution == "" || job.Bitrate == "" {
		log.Printf("⚠️ [Job %s] Missing resolution or bitrate. Skipping job.", job.JobID)
		return
	}

	redisKey := fmt.Sprintf("job:%s", job.JobID)

	valType, err := redisClient.Type(ctx, redisKey).Result()
	if err == nil && valType != "hash" && valType != "none" {
		log.Printf("⚠️ [Job %s] Redis key has wrong type (%s), deleting it...", job.JobID, valType)
		redisClient.Del(ctx, redisKey)
	}

	if err := redisClient.HSet(ctx, redisKey, "codec", job.Codec).Err(); err != nil {
		log.Printf("❌ [Job %s] Failed to write codec to Redis: %v", job.JobID, err)
	} else {
		log.Printf("✅ [Job %s] Wrote codec to Redis: %s", job.JobID, job.Codec)
	}

	ffmpegCodec := MapCodecToFFmpeg(job.Codec)

	localInput, err := DownloadInput(job.InputURL, job.JobID)
	if err != nil {
		log.Printf("❌ [Job %s] Download failed: %v", job.JobID, err)
		return
	}
	defer os.Remove(localInput)

	outputPath := filepath.Join(outputDir, fmt.Sprintf("%s_%s.mp4", job.JobID, job.Representation))

	args := []string{
		"-i", localInput,
		"-vf", fmt.Sprintf("scale=%s", job.Resolution),
		"-c:v", ffmpegCodec,
		"-b:v", job.Bitrate,
		"-g", fmt.Sprintf("%d", job.GopSize),
		"-keyint_min", fmt.Sprintf("%d", job.KeyintMin),
		"-sc_threshold", "0",
		"-an",
	}

	if job.Codec == "vvc" || job.Codec == "h266" {
		args = append(args,
			"-preset", "medium",
			"-threads", "4",
			"-f", "mp4",
		)
	}

	args = append(args,
		"-movflags", "+faststart+frag_keyframe+empty_moov+default_base_moof",
		"-y", outputPath,
	)

	cmd := exec.Command("ffmpeg", args...)
	log.Printf("⚙️ [Job %s] Running FFmpeg: %s", job.JobID, cmd.String())

	stderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("❌ [Job %s] FFmpeg failed: %v\n%s", job.JobID, err, string(stderr))
		return
	}

	log.Printf("✅ [Job %s] Segment generated: %s", job.JobID, outputPath)

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
