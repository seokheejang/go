// cmd/server/main.go
package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"sqs-pause/queue"

	"github.com/gin-gonic/gin"
)

var (
	sqsQueue      *queue.SQSQueue
	currentHeight int64
)

func main() {
	r := gin.Default()

	r.POST("/pause", func(c *gin.Context) {
		sqsQueue.PauseSending()
		c.JSON(http.StatusOK, gin.H{"status": "paused"})
	})

	r.POST("/resume", func(c *gin.Context) {
		sqsQueue.ResumeSending()
		c.JSON(http.StatusOK, gin.H{"status": "resumed"})
	})

	r.POST("/reset", func(c *gin.Context) {
		currentHeight = 0
		log.Println("Resetting current height to 0")
		c.JSON(http.StatusOK, gin.H{"status": "reset"})
	})

	sqsCfg := &queue.SQSQueueConfig{
		Queue:               "test-queue.fifo",
		Region:              "us-east-1",
		Endpoint:            "http://localstack:4566",
		IsFIFO:              true,
		MaxNumberOfMessages: 1,
		WaitTimeSeconds:     1,
		MessageGroupID:      "testgroup",
	}

	q, err := queue.NewSQSQueue(sqsCfg)
	if err != nil {
		panic(err)
	}
	sqsQueue = q

	go func() {
		for {
			heightStr := strconv.FormatInt(currentHeight, 10)
			sqsQueue.PublishMessage(fmt.Sprintf("block_height: %s", heightStr))
			currentHeight++
			time.Sleep(2 * time.Second)
		}
	}()

	r.Run(":8080")
}
