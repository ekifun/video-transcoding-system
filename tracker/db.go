package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

// InitDB initializes SQLite DB and creates the transcoding_jobs table if it doesn't exist.
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
	);`

	_, err = DB.Exec(createTable)
	if err != nil {
		log.Fatalf("❌ Failed to create transcoding_jobs table: %v", err)
	}

	log.Println("📁 SQLite DB initialized.")
}

// InsertOrUpdateJob inserts a new job or updates metadata if jobID exists.
func InsertOrUpdateJob(jobID, streamName, inputURL, codec, representations, workerID, status string) error {
	stmt := `
	INSERT INTO transcoding_jobs
	(job_id, stream_name, input_url, codec, representations, worker_id, status, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	ON CONFLICT(job_id) DO UPDATE SET
		stream_name = excluded.stream_name,
		input_url = excluded.input_url,
		codec = excluded.codec,
		representations = excluded.representations,
		worker_id = excluded.worker_id,
		status = excluded.status,
		updated_at = CURRENT_TIMESTAMP;
	`

	_, err := DB.Exec(stmt, jobID, streamName, inputURL, codec, representations, workerID, status)
	if err != nil {
		return fmt.Errorf("❌ Failed to insert or update job metadata: %w", err)
	}
	return nil
}

// UpdateJobStatus updates only the status and sets updated_at.
func UpdateJobStatus(jobID, status string) error {
	stmt := `
	UPDATE transcoding_jobs
	SET status = ?, updated_at = CURRENT_TIMESTAMP
	WHERE job_id = ?;
	`
	_, err := DB.Exec(stmt, status, jobID)
	if err != nil {
		return fmt.Errorf("❌ Failed to update job status: %w", err)
	}
	return nil
}

// UpdateMPDUrl updates the mpd_url field after MPD generation.
func UpdateMPDUrl(jobID, mpdURL string) error {
	stmt := `
	UPDATE transcoding_jobs
	SET mpd_url = ?, updated_at = CURRENT_TIMESTAMP
	WHERE job_id = ?;
	`
	_, err := DB.Exec(stmt, mpdURL, jobID)
	if err != nil {
		return fmt.Errorf("❌ Failed to update MPD URL: %w", err)
	}
	return nil
}
