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
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const MaxConcurrentFFmpeg = 2

var (
	ctx         = context.Background()
	outputDir   = "/segments"
	redisAddr   = os.Getenv("REDIS_ADDR")
	instanceID  = os.Getenv("WORKER_INSTANCE_ID") // Optional: unique per worker instance

	jobTracker  *JobTracker
	redisClient *redis.Client

	ffmpegSemaphore = make(chan struct{}, MaxConcurrentFFmpeg)
)

func init() {
	redisClient = redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		log.Fatalf("❌ Failed to connect to Redis: %v", err)
	}

	jobTracker = NewJobTracker(redisAddr)
	log.Println("✅ Connected to Redis (Job Tracker and Redis Client)")
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
	case "vp9":
		return "libvpx-vp9"
	case "av1":
		return "libaom-av1"
	default:
		log.Printf("⚠️ Unsupported codec '%s'. Defaulting to h264 (libx264)", codec)
		return "libx264"
	}
}

func HandleTranscodeJob(job TranscodeJob) {
	jobTracker.MarkJobWaiting(job.JobID, instanceID)

	log.Printf("⏳ [Job %s] Waiting for FFmpeg slot...", job.JobID)
	ffmpegSemaphore <- struct{}{}
	log.Printf("🚦 [Job %s] FFmpeg slot acquired. Starting job...", job.JobID)

	defer func() {
		<-ffmpegSemaphore
		log.Printf("🔓 [Job %s] FFmpeg slot released.", job.JobID)
	}()

	runTranscode(job)
}

func runTranscode(job TranscodeJob) {
	if job.Codec == "" {
		job.Codec = "h264"
	}

	log.Printf("📥 [Job %s] Processing Job | Codec=%s | Resolution=%s | Bitrate=%s | GOP=%d | KeyintMin=%d",
		job.JobID, job.Codec, job.Resolution, job.Bitrate, job.GopSize, job.KeyintMin)

	jobTracker.MarkJobProcessing(job.JobID)

	ffmpegCodec := MapCodecToFFmpeg(job.Codec)

	localInput, err := DownloadInput(job.InputURL, job.JobID)
	if err != nil {
		log.Printf("❌ [Job %s] Download failed: %v", job.JobID, err)
		jobTracker.MarkJobFailed(job.JobID)
		return
	}
	defer os.Remove(localInput)

	outputPath := filepath.Join(outputDir, fmt.Sprintf("%s_%s.mp4", job.JobID, job.Representation))

	args := buildFFmpegArgs(localInput, outputPath, job, ffmpegCodec)

	cmd := exec.Command("ffmpeg", args...)
	log.Printf("⚙️ [Job %s] Running FFmpeg: %s", job.JobID, strings.Join(cmd.Args, " "))

	stderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("❌ [Job %s] FFmpeg failed: %v\n%s", job.JobID, err, string(stderr))
		jobTracker.MarkJobFailed(job.JobID)
		return
	}

	log.Printf("✅ [Job %s] Segment generated: %s", job.JobID, outputPath)

	jobTracker.MarkJobDone(job.JobID, outputPath)
}

func buildFFmpegArgs(input, output string, job TranscodeJob, codec string) []string {
	args := []string{
		"-i", input,
		"-vf", fmt.Sprintf("scale=%s", job.Resolution),
		"-c:v", codec,
		"-b:v", job.Bitrate,
		"-g", fmt.Sprintf("%d", job.GopSize),
		"-keyint_min", fmt.Sprintf("%d", job.KeyintMin),
		"-sc_threshold", "0",
		"-an",
	}

	if job.Codec == "av1" {
		args = append(args, "-pix_fmt", "yuv420p", "-cpu-used", "4", "-usage", "good")
	}

	args = append(args, "-f", "mp4")

	if job.Codec == "av1" {
		args = append(args, "-movflags", "+faststart+frag_keyframe+separate_moof+omit_tfhd_offset")
	} else {
		args = append(args, "-movflags", "+faststart+frag_keyframe+empty_moov+default_base_moof")
	}

	args = append(args, "-y", output)

	return args
}
