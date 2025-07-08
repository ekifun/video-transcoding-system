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

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

var (
	ctx          = context.Background()
	requiredReps = []string{"144p", "360p", "720p"}
	segmentsDir  = "/segments" // Shared volume for transcoded files
)

// Redis client for codec lookup
var redisClient = redis.NewClient(&redis.Options{
	Addr: os.Getenv("REDIS_ADDR"), // Example: "redis:6379"
})

// Kafka message structure
type MPDMessage struct {
	JobID  string `json:"job_id"`
	Status string `json:"status"`
}

func main() {
	log.Println("🚀 Starting MPD Generator...")

	// ✅ Verify Redis connectivity
	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		log.Fatalf("❌ Failed to connect to Redis: %v", err)
	}
	log.Println("✅ Connected to Redis")

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

	// ✅ Get codec from Redis hash (shared job structure)
	redisKey := fmt.Sprintf("job:%s", jobID)
	codec, err := redisClient.HGet(ctx, redisKey, "codec").Result()
	if err != nil {
		log.Printf("❌ Failed to read codec from Redis for job %s: %v", jobID, err)
		return
	}

	// ✅ Select DASH profile based on codec
	var profile string
	switch strings.ToLower(codec) {
	case "hevc", "h265":
		profile = "dashh265:live"
	case "h264", "avc":
		profile = "dashavc264:live" // ✅ Correct profile for H.264
	default:
		log.Printf("⚠️ Unknown codec '%s' for job %s. Defaulting to h264 profile", codec, jobID)
		profile = "dashavc264:live"
	}

	args := []string{
		"-dash", "4000",
		"-rap", "-frag-rap",
		"-profile", profile,
		"-out", outputPath,
	}

	// ✅ Append each representation file
	for _, rep := range requiredReps {
		file := filepath.Join(segmentsDir, fmt.Sprintf("%s_%s.mp4", jobID, rep))
		if _, err := os.Stat(file); os.IsNotExist(err) {
			log.Printf("⚠️ Missing representation file: %s", file)
			return
		}
		args = append(args, file)
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

