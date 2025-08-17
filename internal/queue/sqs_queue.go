package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"gn-indexer/internal/domain"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

const SQSLongPollingSec = 20

// SQSQueue implements EventQueue interface using AWS SQS
type SQSQueue struct {
	client   *sqs.SQS
	queueURL string
	config   *QueueConfig
}

// NewSQSQueue creates a new SQS queue instance
func NewSQSQueue(config *QueueConfig) (*SQSQueue, error) {
	// Create AWS session
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(config.Region),
		Credentials: credentials.NewStaticCredentials(config.AccessKeyID, config.SecretAccessKey, ""),
		Endpoint:    aws.String(config.EndpointURL),
		DisableSSL:  aws.Bool(true), // LocalStack uses HTTP
	})
	if err != nil {
		return nil, fmt.Errorf("create aws session: %w", err)
	}

	// Create SQS client
	sqsClient := sqs.New(sess)

	// Get queue URL
	queueURL, err := getQueueURL(sqsClient, config.QueueName)
	if err != nil {
		return nil, fmt.Errorf("get queue URL: %w", err)
	}

	log.Printf("SQSQueue: connected to queue %s at %s", config.QueueName, queueURL)

	return &SQSQueue{
		client:   sqsClient,
		queueURL: queueURL,
		config:   config,
	}, nil
}

// SendEvent sends a parsed event to the SQS queue
func (q *SQSQueue) SendEvent(ctx context.Context, event *domain.ParsedEvent) error {
	// Convert event to JSON
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	// Prepare message
	message := &sqs.SendMessageInput{
		QueueUrl:    aws.String(q.queueURL),
		MessageBody: aws.String(string(eventJSON)),
		MessageAttributes: map[string]*sqs.MessageAttributeValue{
			"EventType": {
				DataType:    aws.String("String"),
				StringValue: aws.String(event.Type),
			},
			"TokenPath": {
				DataType:    aws.String("String"),
				StringValue: aws.String(event.TokenPath),
			},
		},
	}

	// Send message
	result, err := q.client.SendMessageWithContext(ctx, message)
	if err != nil {
		return fmt.Errorf("send message: %w", err)
	}

	log.Printf("SQSQueue: sent event %s to queue, message ID: %s", event.Type, *result.MessageId)
	return nil
}

// ReceiveEvents receives multiple events from the SQS queue
func (q *SQSQueue) ReceiveEvents(ctx context.Context) ([]*domain.ParsedEvent, error) {
	// Prepare receive message input
	input := &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(q.queueURL),
		MaxNumberOfMessages: aws.Int64(int64(q.config.MaxReceiveMessages)),
		VisibilityTimeout:   aws.Int64(int64(q.config.VisibilityTimeout)),
		WaitTimeSeconds:     aws.Int64(SQSLongPollingSec), // Long polling
		MessageAttributeNames: []*string{
			aws.String("All"),
		},
	}

	// Receive messages
	result, err := q.client.ReceiveMessageWithContext(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("receive messages: %w", err)
	}

	// Check if any messages received
	if len(result.Messages) == 0 {
		return nil, nil // No messages available
	}

	log.Printf("SQSQueue: received %d messages from queue", len(result.Messages))

	// Process all messages
	var events []*domain.ParsedEvent
	for _, message := range result.Messages {
		// Parse event from message body
		var event domain.ParsedEvent
		if err := json.Unmarshal([]byte(*message.Body), &event); err != nil {
			log.Printf("SQSQueue: failed to unmarshal message %s: %v", *message.MessageId, err)
			continue // Skip invalid message
		}

		log.Printf("SQSQueue: parsed event %s from message ID: %s", event.Type, *message.MessageId)

		// Delete message from queue (acknowledge)
		_, err = q.client.DeleteMessage(&sqs.DeleteMessageInput{
			QueueUrl:      aws.String(q.queueURL),
			ReceiptHandle: message.ReceiptHandle,
		})
		if err != nil {
			log.Printf("SQSQueue: failed to delete message %s: %v", *message.MessageId, err)
			continue // Skip failed deletion
		}

		events = append(events, &event)
	}

	log.Printf("SQSQueue: successfully processed %d events", len(events))
	return events, nil
}

// Close closes the SQS connection
func (q *SQSQueue) Close() error {
	// SQS client doesn't need explicit closing
	log.Printf("SQSQueue: connection closed")
	return nil
}

// getQueueURL gets the URL for the specified queue
func getQueueURL(client *sqs.SQS, queueName string) (string, error) {
	result, err := client.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: aws.String(queueName),
	})
	if err != nil {
		return "", fmt.Errorf("get queue URL for %s: %w", queueName, err)
	}
	return *result.QueueUrl, nil
}
