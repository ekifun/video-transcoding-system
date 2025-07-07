package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/segmentio/kafka-go"
)

var (
	ctx          = context.Background()
	requiredReps = []string{"144p", "360p", "720p"}
	segmentsDir  = "/segments" // Shared volume
)

type MPDMessage struct {
	JobID  string `json:"job_id"`
	Status string `json:"status"`
}

func main() {
	log.Println("🚀 Starting MPD Generator...")

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{os.Getenv("KAFKA_BROKER")},
		Topic:    "mpd-generation",
		GroupID:  "mpd-generator",
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})

	for {
		m, err := r.ReadMessage(ctx)
		if err != nil {
			log.Fatalf("❌ Kafka read error: %v", err)
		}

		var msg MPDMessage
		if err := json.Unmarshal(m.Value, &msg); err != nil {
			log.Printf("❌ JSON parse error: %v", err)
			continue
		}

		if msg.Status == "ready_for_mpd" {
			generateMPD(msg.JobID)
		}
	}
}

func generateMPD(jobID string) {
	jobDir := filepath.Join(segmentsDir, jobID)
	outputPath := filepath.Join(jobDir, "manifest.mpd")
	os.MkdirAll(jobDir, 0755)

	args := []string{
		"-dash", "4000",
		"-rap", "-frag-rap",
		"-profile", "dashavc264:live",
		"-out", outputPath,
	}

	for _, rep := range requiredReps {
		pattern := filepath.Join(segmentsDir, fmt.Sprintf("%s_%s_*.mp4", jobID, rep))
		args = append(args, pattern)
	}

	cmd := exec.Command("MP4Box", args...)
	log.Printf("📦 Running MP4Box: %s", strings.Join(cmd.Args, " "))

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("❌ MP4Box error: %v\n%s", err, string(out))
		return
	}

	log.Printf("✅ MPD generated: %s", outputPath)
}
