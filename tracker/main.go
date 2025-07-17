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
}

func main() {
	log.Println("🚀 Starting tracker (monitor + API)...")

	// Start background job completion monitor
	go func() {
		for {
			checkCompletedJobs()
			time.Sleep(5 * time.Second)
		}
	}()

	// Start HTTP monitoring API
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

		if jobData["mpd_published"] == "true" {
			continue
		}

		if allRepsDone(jobData) {
			log.Printf("✅ Job %s: all required representations done.", jobID)
			publishReadyForMPD(jobID)

			// Mark job as ready and avoid duplicate publishing
			redisClient.HSet(ctx, key, map[string]interface{}{
				"status":        "ready_for_mpd",
				"mpd_published": "true",
			})
		}
	}
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
