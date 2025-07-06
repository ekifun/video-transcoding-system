func ConsumeTranscodeJobs(topic string) {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{ /* ... */ })
	if err != nil {
		log.Fatalf("❌ Kafka consumer init failed: %v", err)
	}
	defer consumer.Close()

	consumer.SubscribeTopics([]string{topic}, nil)

	for {
		msg, err := consumer.ReadMessage(-1)
		if err != nil {
			log.Printf("❌ Error reading msg: %v", err)
			continue
		}

		var job TranscodeJob
		if err := json.Unmarshal(msg.Value, &job); err != nil {
			log.Printf("❌ Failed to unmarshal job: %v", err)
			continue
		}

		go HandleTranscodeJob(job) // defined in ffmpeg.go
	}
}
