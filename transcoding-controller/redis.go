package main

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client
var ctx = context.Background()

var redisClient *redis.Client

func InitRedis() {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	log.Printf("🔗 Connecting to Redis at: %s", redisAddr)

	redisClient = redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("❌ Could not connect to Redis: %v", err)
	}
	log.Println("✅ Redis connection successful")
}


func StoreJobMetadata(jobID string, request TranscodeRequest) error {
	key := fmt.Sprintf("job:%s", jobID)
	return redisClient.HSet(ctx, key, map[string]interface{}{
		"input_url": request.InputURL,
		"codec":     request.Codec,
		"resolutions": request.Resolutions,
	}).Err()
}
