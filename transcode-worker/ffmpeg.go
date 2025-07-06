package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

// DownloadInput downloads the input file to /tmp/{jobID}_input.mp4 and returns its local path.
func DownloadInput(inputURL string, jobID string) (string, error) {
	log.Printf("🌐 Downloading input from: %s", inputURL)

	localPath := filepath.Join("/tmp", fmt.Sprintf("%s_input.mp4", jobID))

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
	// 🛠️ Default codec to libx264 if not specified
	if job.Codec == "" {
		job.Codec = "libx264"
	}

	localInput, err := DownloadInput(job.InputURL, job.JobID)
	if err != nil {
		log.Printf("❌ Download failed: %v", err)
		return
	}
	defer os.Remove(localInput)

	// Output segment pattern: /tmp/{jobID}_{rep}_%03d.mp4
	outputPattern := fmt.Sprintf("/tmp/%s_%s_%%03d.mp4", job.JobID, job.Representation)

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

	// Notify Kafka that job is done
	PublishStatus(job.JobID, job.Representation, "done")
}
