// queue/sqs.go
package queue

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"sqs-pause/config"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
)

type MessageDeleteOptions string

const (
	AlwaysDelete MessageDeleteOptions = "Always"
	AutoDelete   MessageDeleteOptions = "Auto"
	NoneDelete   MessageDeleteOptions = "None"
)

func (m MessageDeleteOptions) String() string {
	return string(m)
}

func (m MessageDeleteOptions) IsAlways() bool {
	return m == AlwaysDelete
}

func (m MessageDeleteOptions) IsAuto() bool {
	return m == AutoDelete
}

func (m MessageDeleteOptions) IsNone() bool {
	return m == NoneDelete
}

type SQSQueueConfig struct {
	Endpoint            string
	AccountID           string
	Queue               string
	Region              string
	MessageGroupID      string
	MaxNumberOfMessages int64
	WaitTimeSeconds     int64
	IsFIFO              bool
}

func NewSQSConfigBy(conf *config.SQSConfig) *SQSQueueConfig {
	return &SQSQueueConfig{
		Endpoint:            conf.Endpoint,
		AccountID:           conf.AccountID,
		Queue:               conf.Queue,
		Region:              conf.Region,
		IsFIFO:              conf.IsFIFO,
		MessageGroupID:      conf.MessageGroupID,
		MaxNumberOfMessages: conf.MaxNumberOfMessages,
		WaitTimeSeconds:     conf.WaitTimeSeconds,
	}
}

type sqsClient interface {
	SendMessage(input *sqs.SendMessageInput) (*sqs.SendMessageOutput, error)
	ReceiveMessage(input *sqs.ReceiveMessageInput) (*sqs.ReceiveMessageOutput, error)
	DeleteMessage(input *sqs.DeleteMessageInput) (*sqs.DeleteMessageOutput, error)
	CreateQueue(input *sqs.CreateQueueInput) (*sqs.CreateQueueOutput, error)
}

type SQSQueue struct {
	client    sqsClient
	config    *SQSQueueConfig
	queueURL  string
	mu        sync.Mutex
	paused    bool
	receiving bool
	buffer    []string
}

func NewSQSQueue(config *SQSQueueConfig) (*SQSQueue, error) {
	client, err := createSQSClient(config)
	if err != nil {
		return nil, err
	}

	attributes := make(map[string]*string)
	if config.IsFIFO {
		attributes["FifoQueue"] = aws.String("true")
	}

	output, err := client.CreateQueue(&sqs.CreateQueueInput{
		QueueName:  aws.String(config.Queue),
		Attributes: attributes,
	})
	if err != nil {
		return nil, err
	}

	q := &SQSQueue{
		client:    client,
		config:    config,
		queueURL:  *output.QueueUrl,
		paused:    false,
		receiving: true,
	}

	return q, nil
}

func (q *SQSQueue) PauseSending() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.paused = true
	log.Info("SQS sending paused")
}

func (q *SQSQueue) ResumeSending() {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.paused = false
	log.Info("SQS sending resumed")

	// Flush buffered messages
	for _, msg := range q.buffer {
		input := &sqs.SendMessageInput{
			QueueUrl:    aws.String(q.queueURL),
			MessageBody: aws.String(msg),
		}
		if q.config.IsFIFO {
			input.MessageGroupId = aws.String(q.config.MessageGroupID)
			input.MessageDeduplicationId = aws.String(uuid.New().String())
		}

		_, err := q.client.SendMessage(input)
		if err != nil {
			log.Errorf("Failed to flush buffered message: %v", err)
		} else {
			log.Infof("Flushed buffered message: %s", msg)
		}
	}

	q.buffer = nil
}

func (q *SQSQueue) PauseReceiving() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.receiving = false
	log.Info("SQS receiving paused")
}

func (q *SQSQueue) ResumeReceiving() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.receiving = true
	log.Info("SQS receiving resumed")
}

func (q *SQSQueue) PublishMessage(message string) error {
	if message == "" {
		return fmt.Errorf("message cannot be empty")
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	if q.paused {
		log.Warn("SQS sending is paused. Skipping message")
		q.buffer = append(q.buffer, message)
		return nil
	}

	input := &sqs.SendMessageInput{
		QueueUrl:    aws.String(q.queueURL),
		MessageBody: aws.String(message),
	}
	if q.config.IsFIFO {
		input.MessageGroupId = aws.String(q.config.MessageGroupID)
		input.MessageDeduplicationId = aws.String(uuid.New().String())
	}

	_, err := q.client.SendMessage(input)
	return err
}

func createSQSClient(config *SQSQueueConfig) (*sqs.SQS, error) {
	sess, err := session.NewSession(&aws.Config{
		Region:   aws.String(config.Region),
		Endpoint: aws.String(config.Endpoint),
	})
	if err != nil {
		return nil, err
	}
	return sqs.New(sess), nil
}

func (q *SQSQueue) StartListening(handler func(message string) error, deleteOptions ...MessageDeleteOptions) error {
	deleteOption := AutoDelete
	if len(deleteOptions) > 0 {
		deleteOption = deleteOptions[0]
	}

	errChan := make(chan error)

	// Start connection monitoring
	go q.monitorConnection(errChan)

	for {
		q.mu.Lock()
		client := q.client
		q.mu.Unlock()

		resp, err := client.ReceiveMessage(&sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(q.queueURL),
			MaxNumberOfMessages: aws.Int64(q.config.MaxNumberOfMessages),
			WaitTimeSeconds:     aws.Int64(q.config.WaitTimeSeconds),
		})
		if err != nil {
			if (errors.Is(err, &sqs.TooManyEntriesInBatchRequest{}) ||
				errors.Is(err, &sqs.QueueDoesNotExist{}) ||
				errors.Is(err, &sqs.OverLimit{}) ||
				errors.Is(err, &sqs.QueueDeletedRecently{}) ||
				errors.Is(err, &sqs.TooManyEntriesInBatchRequest{})) {
				log.Infof("Queue does not exist or over limit. Retrying in 5 seconds...")

				time.Sleep(5 * time.Second)

				errChan <- err

				continue
			}

			log.Errorf("Failed to receive message: %v", err)

			continue
		}

		for _, msg := range resp.Messages {
			if err := handler(*msg.Body); err != nil {
				log.Infof("Failed to process message: %v", err)
			}

			if deleteOption.IsNone() {
				continue
			}

			// Delete the message after successful processing
			_, err := client.DeleteMessage(&sqs.DeleteMessageInput{
				QueueUrl:      aws.String(q.queueURL),
				ReceiptHandle: msg.ReceiptHandle,
			})
			if err != nil {
				log.Infof("Failed to delete message: %v", err)
			}
		}
	}
}

func (q *SQSQueue) reconnectClient() error {
	q.mu.Lock()
	defer q.mu.Unlock()

	log.Info("Attempting to reconnect to SQS...")

	client, err := createSQSClient(q.config)
	if err != nil {
		return err
	}

	q.client = client
	log.Info("Reconnected to SQS successfully.")

	return nil
}

func (q *SQSQueue) monitorConnection(errChan chan error) {
	for {
		err := <-errChan
		if err != nil {
			log.Infof("SQS connection error: %v. Reconnecting...", err)

			for {
				if err = q.reconnectClient(); err == nil {
					break
				}

				log.Infof("Reconnect failed: %v. Retrying in 5 seconds...", err)

				time.Sleep(5 * time.Second)
			}
		}
	}
}
