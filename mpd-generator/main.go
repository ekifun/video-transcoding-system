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
	ctx         = context.Background()
	segmentsDir = "/segments"
	publicHost  = os.Getenv("PUBLIC_HOST")
)

var redisClient = redis.NewClient(&redis.Options{
	Addr: os.Getenv("REDIS_ADDR"),
})

type MPDMessage struct {
	JobID  string `json:"job_id"`
	Status string `json:"status"`
}

func main() {
	log.Println("🚀 Starting MPD Generator...")

	InitDB() // Only for updating mpd_url

	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		log.Fatalf("❌ Failed to connect to Redis: %v", err)
	}
	log.Println("✅ Connected to Redis")

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{os.Getenv("KAFKA_BROKER")},
		Topic:    "mpd-generation",
		GroupID:  "mpd-generator",
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})

	for {
		m, err := reader.ReadMessage(ctx)
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
	localMPDPath := filepath.Join(jobDir, "manifest.mpd")
	publicMPDURL := fmt.Sprintf("%s/%s/manifest.mpd", strings.TrimRight(publicHost, "/"), jobID)

	os.MkdirAll(jobDir, 0755)

	redisKey := fmt.Sprintf("job:%s", jobID)

	codec, err := redisClient.HGet(ctx, redisKey, "codec").Result()
	if err != nil {
		log.Printf("❌ Failed to read codec from Redis for job %s: %v", jobID, err)
		return
	}
	codec = strings.ToLower(codec)

	requiredListStr, err := redisClient.HGet(ctx, redisKey, "required_resolutions").Result()
	if err != nil {
		log.Printf("❌ Failed to read required_resolutions from Redis for job %s: %v", jobID, err)
		return
	}
	requiredReps := parseRequiredReps(requiredListStr)

	args := []string{
		"-dash", "4000",
		"-rap", "-frag-rap",
		"-out", localMPDPath,
	}

	if codec == "h264" || codec == "avc" {
		args = append([]string{"-profile", "dashavc264:live"}, args...)
	}

	for _, rep := range requiredReps {
		file := filepath.Join(segmentsDir, fmt.Sprintf("%s_%s.mp4", jobID, rep))
		if _, err := os.Stat(file); os.IsNotExist(err) {
			log.Printf("⚠️ Missing segment file: %s", file)
			return
		}
		args = append(args, file)
	}

	cmd := exec.Command("MP4Box", args...)
	log.Printf("📦 Running MP4Box: %s", strings.Join(cmd.Args, " "))

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("❌ MP4Box failed: %v\n%s", err, string(output))
		return
	}

	log.Printf("✅ MPD generated: %s", localMPDPath)

	// Update only MPD URL in DB
	if err := UpdateMPDUrl(jobID, publicMPDURL); err != nil {
		log.Printf("⚠️ Failed to update MPD URL in DB for job %s: %v", jobID, err)
	} else {
		log.Printf("✅ MPD URL updated in DB for job %s", jobID)
	}

	// Mark job as done after MPD generation completes
	if err := UpdateJobStatus(jobID, "done"); err != nil {
		log.Printf("⚠️ Failed to mark job %s as done: %v", jobID, err)
	} else {
		log.Printf("✅ Job %s marked as done in DB", jobID)
	}
}

func parseRequiredReps(input string) []string {
	parts := strings.Split(input, ",")
	var reps []string
	for _, p := range parts {
		rep := strings.TrimSpace(p)
		if rep != "" {
			reps = append(reps, rep)
		}
	}
	return reps
}
