package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

var (
	ctx         = context.Background()
	redisClient *redis.Client
	kafkaWriter *kafka.Writer
)

func init() {
	// Redis
	redisClient = redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})
	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		log.Fatalf("❌ Redis error: %v", err)
	}

	// Kafka
	kafkaWriter = &kafka.Writer{
		Addr:     kafka.TCP(os.Getenv("KAFKA_BROKERS")),
		Topic:    "mpd-generation",
		Balancer: &kafka.LeastBytes{},
	}

	// SQLite DB
	dbPath := os.Getenv("SQLITE_DB_PATH")
	if dbPath == "" {
		dbPath = "/app/db/data/jobs.db"
	}
	InitDB(dbPath)
}

func main() {
	log.Println("🚀 Starting tracker (monitor + API)...")

	go func() {
		for {
			checkCompletedJobs()
			time.Sleep(5 * time.Second)
		}
	}()

	http.HandleFunc("/job-summary", handleJobSummary)
	log.Println("📡 Tracker API running on :9000/job-summary")
	log.Fatal(http.ListenAndServe(":9000", nil))
}

func checkCompletedJobs() {
	keys, err := redisClient.Keys(ctx, "job:*").Result()
	if err != nil {
		log.Printf("❌ Redis scan failed: %v", err)
		return
	}

	for _, key := range keys {
		jobID := strings.TrimPrefix(key, "job:")
		jobData, err := redisClient.HGetAll(ctx, key).Result()
		if err != nil {
			log.Printf("❌ Redis read failed (%s): %v", key, err)
			continue
		}

		// Extract job fields from Redis
		streamName := jobData["stream_name"]
		inputURL := jobData["input_url"]
		codec := jobData["codec"]
		representations := jobData["required_resolutions"]
		workerID := jobData["worker_id"]
		currentStatus := jobData["status"]

		// Safely update DB metadata
		err = SafeUpdateJobMetadata(jobID, streamName, inputURL, codec, representations, workerID, currentStatus)
		if err != nil {
			log.Printf("⚠️ Failed to sync metadata to DB for job %s: %v", jobID, err)
		}

		// Promote from waiting → transcoding if any representation is processing
		if currentStatus == "waiting" && hasActiveRepresentation(jobData) {
			log.Printf("🚧 Job %s entering transcoding...", jobID)
			redisClient.HSet(ctx, key, "status", "transcoding")
			_ = UpdateJobStatus(jobID, "transcoding")
		}

		// Skip completed jobs
		if jobData["mpd_published"] == "true" {
			continue
		}

		// If all representations done, mark job as ready_for_mpd
		if allRepsDone(jobData) {
			log.Printf("✅ Job %s all representations done. Marking ready_for_mpd.", jobID)
			publishReadyForMPD(jobID)

			redisClient.HSet(ctx, key, map[string]interface{}{
				"status":        "ready_for_mpd",
				"mpd_published": "true",
			})

			_ = UpdateJobStatus(jobID, "ready_for_mpd")
		}
	}
}

func hasActiveRepresentation(jobData map[string]string) bool {
	requiredListStr, ok := jobData["required_resolutions"]
	if !ok || requiredListStr == "" {
		return false
	}
	requiredReps := parseRequiredReps(requiredListStr)
	for _, rep := range requiredReps {
		if jobData[rep] == "processing" {
			return true
		}
	}
	return false
}

func allRepsDone(jobData map[string]string) bool {
	requiredListStr, ok := jobData["required_resolutions"]
	if !ok || requiredListStr == "" {
		return false
	}
	requiredReps := parseRequiredReps(requiredListStr)
	for _, rep := range requiredReps {
		if jobData[rep] != "done" {
			return false
		}
	}
	return true
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

func publishReadyForMPD(jobID string) {
	msg := map[string]string{
		"job_id": jobID,
		"status": "ready_for_mpd",
	}
	payload, _ := json.Marshal(msg)

	err := kafkaWriter.WriteMessages(ctx, kafka.Message{
		Key:   []byte(jobID),
		Value: payload,
	})
	if err != nil {
		log.Printf("❌ Kafka publish failed for job %s: %v", jobID, err)
	} else {
		log.Printf("📤 Kafka published (mpd-generation): job_id=%s", jobID)
	}
}

func handleJobSummary(w http.ResponseWriter, r *http.Request) {
	counts := aggregateJobStatuses()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(counts)
}

func aggregateJobStatuses() map[string]int {
	counts := map[string]int{
		"waiting":       0,
		"transcoding":   0,
		"processing":    0,
		"done":          0,
		"failed":        0,
		"ready_for_mpd": 0,
	}

	keys, _ := redisClient.Keys(ctx, "job:*").Result()
	for _, key := range keys {
		status, err := redisClient.HGet(ctx, key, "status").Result()
		if err == nil {
			counts[status]++
		}
	}

	return counts
}
