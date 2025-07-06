package main

import (
    "context"
	"log"
	"os"
    "os/signal"
    "syscall"
)

func main() {
    log.Println("🚀 Starting Transcoder Worker...")

    if err := InitKafka(); err != nil {
        log.Fatalf("❌ Failed to initialize Kafka: %v", err)
        os.Exit(1)
    }

    log.Println("📡 Subscribing to 'transcode-jobs' topic...")

    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()

    go ConsumeTranscodeJobs("transcode-jobs")

    <-ctx.Done()
    log.Println("🛑 Graceful shutdown signal received")
    // TODO: Close Kafka producer/consumer if needed
}

