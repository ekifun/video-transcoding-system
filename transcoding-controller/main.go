package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	ctx = context.Background()

	resolutionMap = map[string]struct {
		Resolution string
		Bitrate    string
	}{
		"144p":  {"256x144", "200k"},
		"240p":  {"426x240", "300k"},
		"360p":  {"640x360", "800k"},
		"480p":  {"854x480", "1200k"},
		"720p":  {"1280x720", "2500k"},
		"1080p": {"1920x1080", "4500k"},
	}
)

func main() {
	log.Println("📦 Starting transcoding controller...")

	InitKafka()
	InitRedis()

	// ✅ Optional health check endpoint
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Transcoding Controller is up")
	})

	http.HandleFunc("/transcode", handleTranscodeRequest)

	log.Println("🚀 Controller running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleTranscodeRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var req TranscodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("❌ JSON decode error: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	jobID := uuid.New().String()
	log.Printf("🆕 New transcode job: %s", jobID)

	func StoreJobMetadata(jobID string, req TranscodeRequest) error {
		data, err := json.Marshal(req)
		if err != nil {
			log.Printf("❌ JSON marshal error: %v", err)
			return err
		}
	
		key := fmt.Sprintf("job:%s", jobID)
		log.Printf("🔄 Storing key: %s", key)
	
		err = redisClient.Set(context.Background(), key, data, 0).Err()
		if err != nil {
			log.Printf("❌ Redis SET error: %v", err)
		} else {
			log.Printf("✅ Job metadata stored for: %s", key)
		}
		return err
	}
	

	for _, rep := range req.Resolutions {
		info, ok := resolutionMap[rep]
		if !ok {
			log.Printf("⚠️ Unsupported resolution: %s", rep)
			continue
		}

		job := TranscodeJob{
			JobID:          jobID,
			InputURL:       req.InputURL,
			Representation: rep,
			Resolution:     info.Resolution,
			Bitrate:        info.Bitrate,
			Codec:          req.Codec,
			OutputPath:     fmt.Sprintf("s3://output/%s/video_%s.mp4", jobID, rep),
		}

		if err := PublishJob("transcode-jobs", job); err != nil {
			log.Printf("❌ Failed to publish job %s: %v", rep, err)
		} else {
			log.Printf("✅ Published job for resolution: %s", rep)
		}
	}

	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintf(w, `{"job_id": "%s", "status": "submitted"}`, jobID)
}
