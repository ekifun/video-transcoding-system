package main

import (
    "encoding/json"
    "log"

    "github.com/confluentinc/confluent-kafka-go/kafka"
)

// NOTE: Ensure this struct is also defined in `model.go` and imported properly.
type TranscodeJob struct {
    JobID          string `json:"job_id"`
    InputURL       string `json:"input_url"`
    Representation string `json:"representation"`
    Resolution     string `json:"resolution"`
    Bitrate        string `json:"bitrate"`
    Codec          string `json:"codec"`
    OutputPath     string `json:"output_path"`
}

var producer *kafka.Producer

func InitKafka() error {
    var err error
    producer, err = kafka.NewProducer(&kafka.ConfigMap{
        "bootstrap.servers": "localhost:9092",
    })
    if err != nil {
        return err
    }

    // Start a goroutine to handle Kafka delivery reports
    go func() {
        for e := range producer.Events() {
            switch ev := e.(type) {
            case *kafka.Message:
                if ev.TopicPartition.Error != nil {
                    log.Printf("❌ Delivery failed: %v\n", ev.TopicPartition.Error)
                } else {
                    log.Printf("✅ Message delivered to %v\n", ev.TopicPartition)
                }
            }
        }
    }()
    return nil
}

func PublishJob(topic string, job TranscodeJob) error {
    payload, err := json.Marshal(job)
    if err != nil {
        return err
    }

    return producer.Produce(&kafka.Message{
        TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
        Key:            []byte(job.JobID),
        Value:          payload,
    }, nil)
}
