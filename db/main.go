package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB() {
	dbPath := os.Getenv("/app/db/data")
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
