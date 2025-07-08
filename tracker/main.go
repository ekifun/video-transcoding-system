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

var (
	requiredReps = []string{"144p", "360p", "720p"}
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
		Addr: 	  kafka.TCP(os.Getenv("KAFKA_BROKERS")),
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
		jobID := strings.TrimPrefix(key, "job:progress:")
		progress, err := redisClient.HGetAll(ctx, key).Result()
		if err != nil {
			log.Printf("❌ Failed to read progress: %v", err)
			continue
		}

		if allRepsDone(progress) {
			log.Printf("✅ All done for job: %s", jobID)
			publishReadyForMPD(jobID)
			redisClient.Del(ctx, key) // cleanup
		}
	}
}

func allRepsDone(progress map[string]string) bool {
	for _, rep := range requiredReps {
		if progress[rep] != "done" {
			return false
		}
	}
	return true
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
