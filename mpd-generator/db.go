package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

// InitDB establishes a connection to the existing SQLite DB.
func InitDB() {
	dbPath := os.Getenv("SQLITE_DB_PATH")
	if dbPath == "" {
		log.Fatal("❌ SQLITE_DB_PATH environment variable is not set")
	}

	var err error
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("❌ Failed to open SQLite DB: %v", err)
	}

	log.Println("📁 SQLite DB initialized (mpd-generator).")
}

// UpdateMPDUrl updates only the mpd_url field for a given job_id.
func UpdateMPDUrl(jobID, mpdURL string) error {
	stmt := `
	UPDATE transcoding_jobs
	SET mpd_url = ?, updated_at = CURRENT_TIMESTAMP
	WHERE job_id = ?;`

	_, err := DB.Exec(stmt, mpdURL, jobID)
	if err != nil {
		return fmt.Errorf("❌ Failed to update MPD URL for job %s: %w", jobID, err)
	}

	log.Printf("✅ Updated MPD URL for job %s", jobID)
	return nil
}
