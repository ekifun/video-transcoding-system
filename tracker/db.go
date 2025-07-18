package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

// InitDB initializes SQLite DB and ensures transcoding_jobs table exists.
func InitDB(dbPath string) {
	var err error
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("❌ Failed to open SQLite DB: %v", err)
	}

	createTable := `
	CREATE TABLE IF NOT EXISTS transcoding_jobs (
		job_id TEXT PRIMARY KEY,
		stream_name TEXT,
		input_url TEXT,
		codec TEXT,
		representations TEXT,
		mpd_url TEXT,
		status TEXT,
		worker_id TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP
	);
	`

	_, err = DB.Exec(createTable)
	if err != nil {
		log.Fatalf("❌ Failed to create transcoding_jobs table: %v", err)
	}

	log.Println("📁 SQLite DB initialized and transcoding_jobs table ready.")
}

// InsertOrUpdateJob performs an upsert into transcoding_jobs table.
func InsertOrUpdateJob(jobID, streamName, inputURL, codec, representations, workerID, status string) error {
	stmt := `
	INSERT INTO transcoding_jobs (
		job_id, stream_name, input_url, codec, representations,
		worker_id, status, created_at, updated_at
	)
	VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	ON CONFLICT(job_id) DO UPDATE SET
		stream_name       = excluded.stream_name,
		input_url         = excluded.input_url,
		codec             = excluded.codec,
		representations   = excluded.representations,
		worker_id         = excluded.worker_id,
		status            = excluded.status,
		updated_at        = CURRENT_TIMESTAMP;
	`

	_, err := DB.Exec(stmt, jobID, streamName, inputURL, codec, representations, workerID, status)
	if err != nil {
		return fmt.Errorf("❌ Failed to insert or update job %s: %w", jobID, err)
	}

	log.Printf("✅ DB upsert: job_id=%s, status=%s", jobID, status)
	return nil
}

// UpdateJobStatus updates only the job status in DB.
func UpdateJobStatus(jobID, status string) error {
	stmt := `
	UPDATE transcoding_jobs
	SET status = ?, updated_at = CURRENT_TIMESTAMP
	WHERE job_id = ?;
	`

	_, err := DB.Exec(stmt, status, jobID)
	if err != nil {
		return fmt.Errorf("❌ Failed to update status for job %s: %w", jobID, err)
	}

	log.Printf("✅ DB status update: job_id=%s, status=%s", jobID, status)
	return nil
}

// UpdateMPDUrl sets the mpd_url for a given job.
func UpdateMPDUrl(jobID, mpdURL string) error {
	stmt := `
	UPDATE transcoding_jobs
	SET mpd_url = ?, updated_at = CURRENT_TIMESTAMP
	WHERE job_id = ?;
	`

	_, err := DB.Exec(stmt, mpdURL, jobID)
	if err != nil {
		return fmt.Errorf("❌ Failed to update MPD URL for job %s: %w", jobID, err)
	}

	log.Printf("✅ DB MPD URL update: job_id=%s, mpd_url=%s", jobID, mpdURL)
	return nil
}

// GetJobByID fetches existing metadata for safe updates.
func GetJobByID(jobID string) (map[string]string, error) {
	stmt := `
	SELECT stream_name, input_url, codec, representations, worker_id, status
	FROM transcoding_jobs
	WHERE job_id = ?;
	`

	row := DB.QueryRow(stmt, jobID)

	var streamName, inputURL, codec, representations, workerID, status sql.NullString

	err := row.Scan(&streamName, &inputURL, &codec, &representations, &workerID, &status)
	if err != nil {
		if err == sql.ErrNoRows {
			return make(map[string]string), nil
		}
		return nil, fmt.Errorf("❌ Failed to fetch job %s: %w", jobID, err)
	}

	return map[string]string{
		"stream_name":     nullStringToString(streamName),
		"input_url":       nullStringToString(inputURL),
		"codec":           nullStringToString(codec),
		"representations": nullStringToString(representations),
		"worker_id":       nullStringToString(workerID),
		"status":          nullStringToString(status),
	}, nil
}

func nullStringToString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

// SafeUpdateJobMetadata prevents overwriting valid DB metadata with missing Redis values.
func SafeUpdateJobMetadata(jobID, streamName, inputURL, codec, representations, workerID, status string) error {
	existing, err := GetJobByID(jobID)
	if err != nil {
		return err
	}

	if streamName == "" {
		streamName = existing["stream_name"]
	}
	if inputURL == "" {
		inputURL = existing["input_url"]
	}
	if codec == "" {
		codec = existing["codec"]
	}
	if representations == "" {
		representations = existing["representations"]
	}
	if workerID == "" {
		workerID = existing["worker_id"]
	}
	if status == "" {
		status = existing["status"]
	}

	return InsertOrUpdateJob(jobID, streamName, inputURL, codec, representations, workerID, status)
}
