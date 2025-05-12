package main

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

var queueURL string

func main() {
	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String("us-east-1"),
		Endpoint:    aws.String("http://localhost:4566"),
		Credentials: credentials.NewStaticCredentials("test", "test", ""),
	}))

	svc := sqs.New(sess)

	resultURL, err := svc.CreateQueue(&sqs.CreateQueueInput{
		QueueName: aws.String("test-queue"),
	})

	if err != nil {
		log.Fatalf("Unable to create queue: %v", err)
	}

	queueURL = *resultURL.QueueUrl

	sendMessage(svc, "Hello, SQS!")
	sendMessage(svc, "Hello, SQS2!")
	receiveMessage(svc)
	receiveMessage(svc)
}

func sendMessage(svc *sqs.SQS, message string) {
	_, err := svc.SendMessage(&sqs.SendMessageInput{
		QueueUrl:    &queueURL,
		MessageBody: aws.String(message),
	})

	if err != nil {
		log.Fatalf("failed to send message, %v", err)
	}

	log.Println("message sent")
}

func receiveMessage(svc *sqs.SQS) {
	result, err := svc.ReceiveMessage(&sqs.ReceiveMessageInput{
		QueueUrl:            &queueURL,
		MaxNumberOfMessages: aws.Int64(1),
		WaitTimeSeconds:     aws.Int64(5),
	})

	if err != nil {
		log.Fatalf("failed to receive message, %v", err)
	}

	if len(result.Messages) == 0 {
		fmt.Println("No messages received")
		return
	}

	for _, message := range result.Messages {
		log.Printf("message received: %s, id: %s, len: %d", *message.Body, *message.MessageId, len(result.Messages))

		_, err := svc.DeleteMessage(&sqs.DeleteMessageInput{
			QueueUrl:      &queueURL,
			ReceiptHandle: message.ReceiptHandle,
		})
		if err != nil {
			log.Fatalf("Unable to delete message: %v", err)
		}
		fmt.Println("Message deleted successfully!")
	}
}
