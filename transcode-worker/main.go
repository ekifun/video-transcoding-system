package main

import (
	"log"
	"os"
)

func main() {
	log.Println("🚀 Starting Transcoder Worker...")

	if err := InitKafka(); err != nil {
		log.Fatalf("❌ Failed to initialize Kafka: %v", err)
		os.Exit(1)
	}

	log.Println("📡 Subscribing to 'transcode-jobs' topic...")
	ConsumeTranscodeJobs("transcode-jobs")
}
