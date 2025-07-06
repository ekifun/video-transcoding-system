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

// HandleTranscodeJob downloads input, transcodes it using FFmpeg, and publishes status
func HandleTranscodeJob(job TranscodeJob) {
	localInput, err := DownloadInput(job.InputURL, job.JobID)
	if err != nil {
		log.Printf("❌ Download failed: %v", err)
		return
	}
	defer os.Remove(localInput) // optional: clean up input file

	outputPath := fmt.Sprintf("/tmp/output_%s.mp4", job.Representation)
	defer os.Remove(outputPath) // optional: clean up output file

	cmd := exec.Command("ffmpeg",
		"-i", localInput,
		"-vf", fmt.Sprintf("scale=%s", job.Resolution),
		"-b:v", job.Bitrate,
		"-c:v", job.Codec,
		"-y", outputPath,
	)

	// Capture FFmpeg stderr output
	stderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("❌ FFmpeg failed: %v\n%s", err, string(stderr))
		return
	}

	log.Printf("✅ Transcoding done: %s", outputPath)

	// Notify Kafka that job is done
	PublishStatus(job.JobID, job.Representation, "done")
}
