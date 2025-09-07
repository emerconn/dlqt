package main

import (
	"context"
	"fmt"
	"log"

	"dlqt/internal/servicebus"

	"github.com/urfave/cli/v3"
)

func seedMessages(ctx context.Context, cmd *cli.Command) error {
	namespace := cmd.String("namespace")
	queue := cmd.String("queue")
	numMessages := cmd.Int("num-messages")

	log.Println("namespace:", namespace)
	log.Println("queue:", queue)
	log.Println("number of messages:", numMessages)

	client, err := servicebus.GetClient(namespace)
	if err != nil {
		return fmt.Errorf("failed to get Service Bus client: %w", err)
	}

	messages := make([]string, numMessages)
	for i := range messages {
		messages[i] = fmt.Sprintf("testMessage%d", i+1)
	}
	log.Printf("seeding %d messages", len(messages))
	if err := servicebus.SendMessageBatch(ctx, client, queue, messages[:]); err != nil {
		return fmt.Errorf("failed to send messages: %w", err)
	}

	if !cmd.Bool("no-dlq") {
		log.Println("moving messages to dead-letter queue")
		if err := servicebus.DeadLetterMessages(ctx, client, queue, len(messages)); err != nil {
			return fmt.Errorf("failed to dead-letter messages: %w", err)
		}
	} else {
		log.Println("skipping dead-lettering messages")
	}

	return nil
}
