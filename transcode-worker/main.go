func main() {
	log.Println("🚀 Starting Transcoder Worker...")
	InitKafka()

	ConsumeTranscodeJobs("transcode-jobs") // defined in kafka.go
}
