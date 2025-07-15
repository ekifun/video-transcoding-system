package main

import (
	"context"
	"encoding/json"
	"log"
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

	for {
		checkCompletedJobs()
		time.Sleep(5 * time.Second)
	}
}

func checkCompletedJobs() {
	keys, err := redisClient.Keys(ctx, "job:*").Result()
	if err != nil {
		log.Printf("❌ Failed to scan Redis: %v", err)
		return
	}

	for _, key := range keys {
		jobID := strings.TrimPrefix(key, "job:")
		jobData, err := redisClient.HGetAll(ctx, key).Result()
		if err != nil {
			log.Printf("❌ Failed to read Redis hash for %s: %v", key, err)
			continue
		}

		if allRepsDone(jobData) {
			log.Printf("✅ All done for job: %s", jobID)
			publishReadyForMPD(jobID)

			err := redisClient.HSet(ctx, key, "status", "ready_for_mpd").Err()
			if err != nil {
				log.Printf("⚠️ Failed to update job status in Redis: %v", err)
			}
		}
	}
}

func allRepsDone(progress map[string]string) bool {
	requiredListStr, ok := progress["required_resolutions"]
	if !ok || requiredListStr == "" {
		log.Printf("⚠️ Missing or empty required_resolutions field in Redis job metadata")
		return false
	}

	requiredReps := parseRequiredReps(requiredListStr)

	for _, rep := range requiredReps {
		status := progress[rep]
		if status != "done" {
			log.Printf("⏳ Representation %s is not done yet (status=%s)", rep, status)
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
		log.Printf("❌ Kafka publish failed: %v", err)
	} else {
		log.Printf("📤 Kafka published: jobID=%s, topic=mpd-generation", jobID)
	}
}
