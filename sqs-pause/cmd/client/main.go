// cmd/client/main.go
package main

import (
	"fmt"
	"log"

	"sqs-pause/queue"
)

func main() {
	sqsCfg := &queue.SQSQueueConfig{
		Queue:               "test-queue.fifo",
		Region:              "us-east-1",
		Endpoint:            "http://localstack:4566",
		IsFIFO:              false,
		MaxNumberOfMessages: 1,
		WaitTimeSeconds:     1,
		MessageGroupID:      "testgroup",
	}

	q, err := queue.NewSQSQueue(sqsCfg)
	if err != nil {
		log.Fatalf("failed to create SQSQueue: %v", err)
	}

	handler := func(msg string) error {
		fmt.Println("[handler] received message:", msg)
		return nil
	}

	if err := q.StartListening(handler); err != nil {
		log.Fatalf("failed to start listening: %v", err)
	}
}
