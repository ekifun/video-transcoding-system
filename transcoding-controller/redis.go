package main

import (
	"context"
	"encoding/json"
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
	data, err := json.Marshal(req)
	if err != nil {
		log.Printf("❌ JSON marshal error: %v", err)
		return err
	}

	key := fmt.Sprintf("job:%s", jobID)
	log.Printf("🔄 Storing job metadata with key: %s", key)

	err = redisClient.Set(ctx, key, data, 0).Err()
	if err != nil {
		log.Printf("❌ Redis SET error: %v", err)
		return err
	}

	log.Printf("✅ Job metadata stored for jobID %s", jobID)
	return nil
}
