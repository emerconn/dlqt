package main

import (
	"context"
	"fmt"
	"log"

	"dlqt/internal/servicebus"
	"github.com/urfave/cli/v3"
)

func retrigger(ctx context.Context, c *cli.Command) error {
	namespace := c.String("namespace")
	queue := c.String("queue")
	messageID := c.String("message-id")

	if messageID == "" {
		return fmt.Errorf("message-id is required")
	}

	log.Printf("Retriggering message %s from DLQ for namespace: %s, queue: %s", messageID, namespace, queue)

	// Get Azure Service Bus client
	client, err := servicebus.GetClient(namespace)
	if err != nil {
		return fmt.Errorf("failed to get Service Bus client: %w", err)
	}
	defer client.Close(ctx)

	// Retrigger the message from DLQ
	err = servicebus.RetriggerDeadLetterMessage(ctx, client, queue, messageID)
	if err != nil {
		return fmt.Errorf("failed to retrigger message: %w", err)
	}

	log.Printf("Successfully retriggered message %s", messageID)
	return nil
}
