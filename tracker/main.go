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
	log.Println("🚀 Starting transcode-complete-tracker...")

	// Start background Redis monitoring
	go func() {
		for {
			checkCompletedJobs()
			time.Sleep(5 * time.Second)
		}
	}()

	// Start HTTP API
	http.HandleFunc("/job-summary", handleJobSummary)
	log.Println("📡 Job Tracker API available at :9000/job-summary")
	log.Fatal(http.ListenAndServe(":9000", nil))
}

func handleJobSummary(w http.ResponseWriter, r *http.Request) {
	counts := aggregateJobStatuses()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(counts)
}

func aggregateJobStatuses() map[string]int {
	counts := map[string]int{
		"waiting":    0,
		"processing": 0,
		"done":       0,
		"failed":     0,
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
