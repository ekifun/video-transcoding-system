package main

import (
	"database/sql"
	"log"
	"os"
	"strings"

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
	JobID           string `json:"job_id"`
	StreamName      string `json:"stream_name"`
	InputURL        string `json:"input_url"`
	Codec           string `json:"codec"`
	Representations string `json:"representations"`
	MPDURL          string `json:"mpd_url"`
	Status          string `json:"status"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

// InsertJobToDB inserts a job immediately upon submission with status like "waiting"
func InsertJobToDB(jobID, streamName, inputURL, codec string, resolutions []string, status string) error {
	reps := strings.Join(resolutions, ",")
	stmt := `
	INSERT INTO transcoding_jobs
	(job_id, stream_name, input_url, codec, representations, status, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	ON CONFLICT(job_id) DO UPDATE SET
		stream_name=excluded.stream_name,
		input_url=excluded.input_url,
		codec=excluded.codec,
		representations=excluded.representations,
		status=excluded.status,
		updated_at=CURRENT_TIMESTAMP;
	`
	_, err := db.Exec(stmt, jobID, streamName, inputURL, codec, reps, status)
	if err != nil {
		log.Printf("⚠️ Failed to insert job to DB (job_id=%s): %v", jobID, err)
	} else {
		log.Printf("✅ Inserted job to DB: job_id=%s, status=%s", jobID, status)
	}
	return err
}

// GetAllTranscodedJobs lists recent jobs for frontend display
func GetAllTranscodedJobs(limit int) ([]TranscodedJob, error) {
	rows, err := db.Query(`
		SELECT job_id, stream_name, input_url, codec, representations, mpd_url, status, created_at, updated_at
		FROM transcoding_jobs
		ORDER BY created_at DESC
		LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []TranscodedJob
	for rows.Next() {
		var job TranscodedJob
		err := rows.Scan(
			&job.JobID,
			&job.StreamName,
			&job.InputURL,
			&job.Codec,
			&job.Representations,
			&job.MPDURL,
			&job.Status,
			&job.CreatedAt,
			&job.UpdatedAt,
		)
		if err != nil {
			log.Printf("⚠️ Scan error: %v", err)
			continue
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}
