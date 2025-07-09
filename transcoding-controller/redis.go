package main

import (
	"context"
	"fmt"
	"log"
	"os" // ✅ ADD THIS

	"github.com/redis/go-redis/v9"
)

var (
	redisClient *redis.Client
	ctx         = context.Background()
)

// InitRedis initializes the Redis client using REDIS_ADDR or defaults to localhost
func InitRedis() {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	log.Printf("🔗 Connecting to Redis at: %s", redisAddr)

	redisClient = redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("❌ Could not connect to Redis: %v", err)
	}
	log.Println("✅ Redis connection successful")
}

// StoreJobMetadata saves the TranscodeRequest under a Redis key "job:<jobID>"
func StoreJobMetadata(jobID string, req TranscodeRequest) error {
	key := fmt.Sprintf("job:%s", jobID)
	log.Printf("🔄 Storing job metadata with key: %s", key)

	data := map[string]interface{}{
		"stream_name": req.StreamName,
		"input_url":   req.InputURL,
		"codec":       req.Codec,
	}

	if err := redisClient.HSet(ctx, key, data).Err(); err != nil {
		log.Printf("❌ Redis HSET error: %v", err)
		return err
	}

	log.Printf("✅ Job metadata stored in Redis hash for jobID %s", jobID)
	return nil
}


