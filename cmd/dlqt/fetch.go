package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"dlqt/internal/servicebus"
	"github.com/urfave/cli/v3"
)

func fetch(ctx context.Context, c *cli.Command) error {
	namespace := c.String("namespace")
	queue := c.String("queue")

	log.Printf("Fetching one message from DLQ for namespace: %s, queue: %s", namespace, queue)

	// Get Azure Service Bus client
	client, err := servicebus.GetClient(namespace)
	if err != nil {
		return fmt.Errorf("failed to get Service Bus client: %w", err)
	}
	defer client.Close(ctx)

	// Fetch one message from the dead letter queue
	message, err := servicebus.FetchDeadLetterMessage(ctx, client, queue)
	if err != nil {
		return fmt.Errorf("failed to fetch message from DLQ: %w", err)
	}

	if message == nil {
		log.Println("No messages found in dead letter queue")
		return nil
	}

	// Output the message details as JSON
	messageDetails := map[string]interface{}{
		"messageID":      message.MessageID,
		"body":          string(message.Body),
		"enqueuedTime":  message.EnqueuedTime,
		"deliveryCount": message.DeliveryCount,
		"properties":    message.ApplicationProperties,
	}

	jsonOutput, err := json.MarshalIndent(messageDetails, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal message details: %w", err)
	}

	fmt.Println(string(jsonOutput))
	log.Printf("Successfully fetched message: %s", message.MessageID)
	
	return nil
}
