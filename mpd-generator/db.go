package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

// InitDB initializes the SQLite database and creates the transcoded_jobs table.
func InitDB() {
	dbPath := os.Getenv("SQLITE_DB_PATH") // Correct environment variable key
	if dbPath == "" {
		log.Fatal("❌ SQLITE_DB_PATH environment variable is not set")
	}

	var err error
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("❌ Failed to open SQLite DB: %v", err)
	}

	createTable := `
	CREATE TABLE IF NOT EXISTS transcoded_jobs (
		job_id TEXT PRIMARY KEY,
		stream_name TEXT,
		original_url TEXT,
		codec TEXT,
		representations TEXT,
		mpd_url TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = DB.Exec(createTable)
	if err != nil {
		log.Fatalf("❌ Failed to create transcoded_jobs table: %v", err)
	}
	log.Println("📁 SQLite DB initialized.")
}

// SaveJobToDB inserts or updates a job record into the transcoded_jobs table.
func SaveJobToDB(jobID, streamName, originalURL, codec, representations, mpdURL string) error {
	stmt := `
	INSERT OR REPLACE INTO transcoded_jobs
	(job_id, stream_name, original_url, codec, representations, mpd_url)
	VALUES (?, ?, ?, ?, ?, ?);`

	_, err := DB.Exec(stmt, jobID, streamName, originalURL, codec, representations, mpdURL)
	if err != nil {
		return fmt.Errorf("❌ DB insert error: %w", err)
	}
	return nil
}
