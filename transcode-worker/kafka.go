package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type TranscodeStatus struct {
	JobID          string `json:"job_id"`
	Representation string `json:"representation"`
	Status         string `json:"status"` // e.g., "done", "failed"
}

var kafkaProducer *kafka.Producer
var kafkaStatusTopic = "transcode-status"

// InitKafka initializes Kafka producer
func InitKafka() error {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:9092"
	}

	var err error
	kafkaProducer, err = kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": brokers,
	})
	if err != nil {
		return err
	}

	log.Println("✅ Kafka producer initialized")
	return nil
}

// PublishStatus publishes a job status to Kafka
func PublishStatus(jobID, representation, status string) {
	msg := TranscodeStatus{
		JobID:          jobID,
		Representation: representation,
		Status:         status,
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		log.Printf("❌ Failed to marshal status message: %v", err)
		return
	}

	err = kafkaProducer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &kafkaStatusTopic,
			Partition: kafka.PartitionAny,
		},
		Value: payload,
	}, nil)

	if err != nil {
		log.Printf("❌ Failed to publish status to Kafka: %v", err)
		return
	}

	log.Printf("📤 Published status to Kafka: jobID=%s, rep=%s, status=%s", jobID, representation, status)
}

// ConsumeTranscodeJobs reads from the Kafka job topic and dispatches transcode jobs
func ConsumeTranscodeJobs(topic string) {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:9092"
	}

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": brokers,
		"group.id":          "transcode-worker-group",
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		log.Fatalf("❌ Kafka consumer init failed: %v", err)
	}
	defer consumer.Close()

	err = consumer.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		log.Fatalf("❌ Failed to subscribe to topic: %v", err)
	}

	log.Printf("🎧 Listening for jobs on topic: %s", topic)

	for {
		msg, err := consumer.ReadMessage(-1)
		if err != nil {
			log.Printf("❌ Error reading message: %v", err)
			continue
		}

		var job TranscodeJob
		if err := json.Unmarshal(msg.Value, &job); err != nil {
			log.Printf("❌ Failed to unmarshal job: %v", err)
			continue
		}

		log.Printf("🆕 Received job: %+v", job)
		go HandleTranscodeJob(job) // from ffmpeg.go
	}
}
