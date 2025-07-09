package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func InitDB() {
	dbPath := os.Getenv("SQLITE_DB_PATH")
	if dbPath == "" {
		dbPath = "/app/db/data/jobs.db" // fallback
	}
	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("❌ Failed to open DB: %v", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatalf("❌ Failed to connect to DB: %v", err)
	}
	log.Printf("✅ Connected to DB: %s", dbPath)
}

type TranscodedJob struct {
	JobID         string `json:"job_id"`
	StreamName    string `json:"stream_name"`
	InputURL      string `json:"input_url"`
	Codec         string `json:"codec"`
	Representations string `json:"representations"`
	MPDURL        string `json:"mpd_url"`
	CreatedAt     string `json:"created_at"`
}

func GetAllTranscodedJobs(limit int) ([]TranscodedJob, error) {
	rows, err := db.Query(`
		SELECT job_id, stream_name, input_url, codec, representations, mpd_url, created_at
		FROM transcoded_jobs
		ORDER BY created_at DESC
		LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []TranscodedJob
	for rows.Next() {
		var job TranscodedJob
		err := rows.Scan(&job.JobID, &job.StreamName, &job.InputURL, &job.Codec, &job.Representations, &job.MPDURL, &job.CreatedAt)
		if err != nil {
			log.Printf("⚠️ Scan error: %v", err)
			continue
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}
