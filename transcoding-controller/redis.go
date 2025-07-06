package main

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client
var ctx = context.Background()

func InitRedis() {
	redisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
}

func StoreJobMetadata(jobID string, request TranscodeRequest) error {
	key := fmt.Sprintf("job:%s", jobID)
	return redisClient.HSet(ctx, key, map[string]interface{}{
		"input_url": request.InputURL,
		"codec":     request.Codec,
		"resolutions": request.Resolutions,
	}).Err()
}
