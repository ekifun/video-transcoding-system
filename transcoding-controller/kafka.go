package main

import (
	"encoding/json"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

var producer *kafka.Producer

func InitKafka() error {
	var err error
	producer, err = kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
	})
	return err
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
