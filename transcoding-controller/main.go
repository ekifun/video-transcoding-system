package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"log"
	"net/http"
)

var resolutionMap = map[string]struct {
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

var validCodecs = map[string]bool{
	"h264": true,
	"hevc": true,
}

func main() {
	log.Println("📦 Starting transcoding controller...")

	InitKafka()
	InitRedis()
	InitDB() // ✅ Initialize SQLite DB connection

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Transcoding Controller is up")
	})

	http.HandleFunc("/transcode", handleTranscodeRequest)
	http.HandleFunc("/jobs", handleListJobs) // ✅ Add /jobs endpoint

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

	log.Printf("📥 Received transcode request: %+v", req)

	// ✅ Validate fields
	if req.StreamName == "" || req.InputURL == "" || len(req.Resolutions) == 0 || req.Codec == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}
	if !validCodecs[req.Codec] {
		http.Error(w, "Unsupported codec", http.StatusBadRequest)
		return
	}

	jobID := uuid.New().String()
	log.Printf("🆕 New transcode job: %s", jobID)

	if err := StoreJobMetadata(jobID, req); err != nil {
		http.Error(w, "Failed to store metadata", http.StatusInternalServerError)
		log.Printf("❌ Failed to store metadata: %v", err)
		return
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

func handleListJobs(w http.ResponseWriter, r *http.Request) {
	jobs, err := GetAllTranscodedJobs(50)
	if err != nil {
		http.Error(w, "Failed to fetch jobs", http.StatusInternalServerError)
		log.Printf("❌ Failed to fetch jobs: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jobs)
}
