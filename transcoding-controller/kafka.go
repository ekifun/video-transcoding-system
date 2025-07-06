package main

import (
    "encoding/json"
    "log" // <- used below
    "github.com/confluentinc/confluent-kafka-go/kafka"
)

var producer *kafka.Producer

func InitKafka() error {
    var err error
    producer, err = kafka.NewProducer(&kafka.ConfigMap{
        "bootstrap.servers": "kafka:9092"
    })
    if err != nil {
        return err
    }

    go func() {
        for e := range producer.Events() {
            switch ev := e.(type) {
            case *kafka.Message:
                if ev.TopicPartition.Error != nil {
                    log.Printf("❌ Delivery failed: %v\n", ev.TopicPartition.Error)
                } else {
                    log.Printf("✅ Delivered to %v\n", ev.TopicPartition)
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
