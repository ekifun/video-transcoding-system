package main

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type JobTracker struct {
	redisClient *redis.Client
	ctx         context.Context
}

func NewJobTracker(redisAddr string) *JobTracker {
	client := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	ctx := context.Background()
	if _, err := client.Ping(ctx).Result(); err != nil {
		panic(fmt.Sprintf("❌ Failed to connect to Redis: %v", err))
	}

	return &JobTracker{
		redisClient: client,
		ctx:         ctx,
	}
}

func (jt *JobTracker) SetJobStatus(jobID, status string) {
	key := fmt.Sprintf("job:%s", jobID)
	jt.redisClient.HSet(jt.ctx, key,
		"status", status,
	)
}

func (jt *JobTracker) MarkJobWaiting(jobID, workerID string) {
	key := fmt.Sprintf("job:%s", jobID)
	jt.redisClient.HSet(jt.ctx, key,
		"status", "waiting",
		"worker_id", workerID,
	)
}

func (jt *JobTracker) MarkJobProcessing(jobID string) {
	key := fmt.Sprintf("job:%s", jobID)
	jt.redisClient.HSet(jt.ctx, key,
		"status", "processing",
		"started_at", time.Now().Format(time.RFC3339),
	)
}

func (jt *JobTracker) MarkJobDone(jobID, outputPath string) {
	key := fmt.Sprintf("job:%s", jobID)
	jt.redisClient.HSet(jt.ctx, key,
		"status", "done",
		"completed_at", time.Now().Format(time.RFC3339),
		"output_path", outputPath,
	)
	jt.redisClient.Expire(jt.ctx, key, 24*time.Hour)
}

func (jt *JobTracker) MarkJobFailed(jobID string) {
	key := fmt.Sprintf("job:%s", jobID)
	jt.redisClient.HSet(jt.ctx, key,
		"status", "failed",
		"completed_at", time.Now().Format(time.RFC3339),
	)
	jt.redisClient.Expire(jt.ctx, key, 24*time.Hour)
}
