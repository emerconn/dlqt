package main

import (
	"context"
	"fmt"
	"log"

	"dlqt/internal/servicebus"

	"github.com/urfave/cli/v3"
)

func purgeMessages(ctx context.Context, c *cli.Command) error {
	namespace := c.String("namespace")
	queue := c.String("queue")

	log.Println("namespace:", namespace)
	log.Println("queue:", queue)

	client, err := servicebus.GetClient(namespace)
	if err != nil {
		return fmt.Errorf("failed to get Service Bus client: %w", err)
	}

	if !c.Bool("no-queue") {
		log.Println("purging queue")
		if err := servicebus.PurgeQueue(ctx, client, queue); err != nil {
			return fmt.Errorf("failed to purge queue '%s': %w", queue, err)
		}
	}

	if !c.Bool("no-dlq") {
		log.Println("purging dead-letter queue")
		if err := servicebus.PurgeDeadLetterQueue(ctx, client, queue); err != nil {
			return fmt.Errorf("failed to purge dead-letter queue for '%s': %w", queue, err)
		}
	}

	return nil
}
