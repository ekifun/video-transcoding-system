package main

import (
	"context"
	"fmt"
	"strings"
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
	jt.redisClient.HSet(jt.ctx, key, "status", status)
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

func (jt *JobTracker) MarkJobFailed(jobID string) {
	key := fmt.Sprintf("job:%s", jobID)
	jt.redisClient.HSet(jt.ctx, key,
		"status", "failed",
		"completed_at", time.Now().Format(time.RFC3339),
	)
	jt.redisClient.Expire(jt.ctx, key, 24*time.Hour)
}

// ✅ New: Track per-representation status and output
func (jt *JobTracker) UpdateRepresentationStatus(jobID, resolution, status, outputPath string) {
	key := fmt.Sprintf("job:%s", jobID)

	// Example:
	// 360p = done
	// 360p_output = /segments/jobID_360p.mp4
	jt.redisClient.HSet(jt.ctx, key,
		resolution, status,
		fmt.Sprintf("%s_output", resolution), outputPath,
	)

	// Check if parent job can be marked done
	jt.checkIfJobCompleted(jobID)
}

// ✅ Helper to check completion across all required representations
func (jt *JobTracker) checkIfJobCompleted(jobID string) {
	key := fmt.Sprintf("job:%s", jobID)

	requiredListStr, err := jt.redisClient.HGet(jt.ctx, key, "required_resolutions").Result()
	if err != nil || requiredListStr == "" {
		return // Cannot verify completeness without required_resolutions
	}

	requiredReps := strings.Split(requiredListStr, ",")
	allDone := true

	for _, rep := range requiredReps {
		rep = strings.TrimSpace(rep)
		status, _ := jt.redisClient.HGet(jt.ctx, key, rep).Result()
		if status != "done" {
			allDone = false
			break
		}
	}

	if allDone {
		jt.redisClient.HSet(jt.ctx, key,
			"status", "done",
			"completed_at", time.Now().Format(time.RFC3339),
		)
		jt.redisClient.Expire(jt.ctx, key, 24*time.Hour)
	}
}
